// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub for
package cloudhub

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"

	"gorm.io/gorm"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/database"
)

const (
	lockTime = 9
)

// AuthFailedRecord token error then record ip
type AuthFailedRecord struct {
	IP         string `gorm:"type:char(32);unique;not null"`
	ErrorTimes int    `gorm:"type:integer;not null"`
}

//  LockRecord lock ip record
type LockRecord struct {
	IP       string `gorm:"type:char(32);unique;not null"`
	LockTime int64  `gorm:"type:integer"`
}

var (
	repositoryInitOnce sync.Once
	lockRepository     LockRepository
)

type lockRepositoryImpl struct {
	db *gorm.DB
}

// LockRepository for config method to operate db
type LockRepository interface {
	recordFailed(string) error
	isLock(string) (bool, error)
	deleteOneLockRecord(string) error
	UnlockRecords() error
	updateLockTime(string) error
	getFailedRecord(string) (*AuthFailedRecord, error)
	createFailedRecord(ip string) error
	updateFailedRecord(ip string) error
	createLockRecord(ip string) error
	deleteFailedRecord(ip string) error
}

// LockRepositoryInstance returns the singleton instance of token lock service
func LockRepositoryInstance() LockRepository {
	repositoryInitOnce.Do(func() {
		lockRepository = &lockRepositoryImpl{db: database.GetDb()}
	})
	return lockRepository
}

func (c *lockRepositoryImpl) isLock(ip string) (bool, error) {
	var lockInfo LockRecord
	err := c.db.Model(LockRecord{}).Where("ip = ?", ip).First(&lockInfo).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	} else if err != nil {
		hwlog.RunLog.Error(err)
		return true, fmt.Errorf("get lock info from ip %s error", ip)
	}
	if time.Now().Unix() > lockInfo.LockTime {
		if err := c.deleteOneLockRecord(ip); err != nil {
			return false, err
		}
		return false, nil
	}
	if err := c.updateLockTime(ip); err != nil {
		return true, fmt.Errorf("update %s lock time error", ip)
	}
	return true, nil
}

func (c *lockRepositoryImpl) recordFailed(ip string) error {
	record, err := c.getFailedRecord(ip)
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.New("get auth failed record from db error")
	}
	if err == gorm.ErrRecordNotFound {
		return c.createFailedRecord(ip)
	}

	if record.ErrorTimes+1 < lockTime {
		return c.updateFailedRecord(ip)
	}
	if err := c.createLockRecord(ip); err != nil {
		return errors.New("create lock record to db error")
	}
	hwlog.OpLog.Infof("lock edge (%s)", ip)
	if err := c.deleteFailedRecord(ip); err != nil {
		return err
	}
	return nil
}

func (c *lockRepositoryImpl) getFailedRecord(ip string) (*AuthFailedRecord, error) {
	var failedRecord AuthFailedRecord
	err := c.db.Model(AuthFailedRecord{}).Where("ip = ?", ip).First(&failedRecord).Error
	return &failedRecord, err
}

func (c *lockRepositoryImpl) createFailedRecord(ip string) error {
	record := AuthFailedRecord{
		IP:         ip,
		ErrorTimes: 1,
	}
	if createErr := c.db.Model(AuthFailedRecord{}).Create(record).Error; createErr != nil {
		return errors.New("create auth failed record to db error")
	}
	return nil
}

func (c *lockRepositoryImpl) updateFailedRecord(ip string) error {
	oldRecord, err := c.getFailedRecord(ip)
	if err != nil {
		return errors.New("get failed record from db error")
	}
	record := map[string]interface{}{
		"ErrorTimes": oldRecord.ErrorTimes + 1,
	}
	if err := c.db.Model(AuthFailedRecord{}).Where("ip = ?", ip).UpdateColumns(record).Error; err != nil {
		return errors.New("update auth failed record to db error")
	}
	return nil
}

func (c *lockRepositoryImpl) deleteFailedRecord(ip string) error {
	if err := c.db.Model(AuthFailedRecord{}).Where("ip = ?", ip).Delete(&AuthFailedRecord{}).Error; err != nil {
		return fmt.Errorf("delete failed record(%s) from db error", ip)
	}
	return nil
}

func (c *lockRepositoryImpl) UnlockRecords() error {
	return c.db.Model(LockRecord{}).Where("lock_time < ?", time.Now().Unix()).Delete(LockRecord{}).Error
}

func (c *lockRepositoryImpl) deleteOneLockRecord(ip string) error {
	if err := c.db.Model(LockRecord{}).Where("ip = ?", ip).Delete(&LockRecord{}).Error; err != nil {
		return fmt.Errorf("delete lock info from ip:%s error", ip)
	}
	hwlog.OpLog.Infof("edge (%s) is unlock", ip)
	return nil
}

func (c *lockRepositoryImpl) updateLockTime(ip string) error {
	updatedColumns := map[string]interface{}{
		"LockTime": time.Now().Add(common.LockInterval).Unix(),
	}
	return c.db.Model(LockRecord{}).Where("ip = ?", ip).UpdateColumns(updatedColumns).Error
}

func (c *lockRepositoryImpl) createLockRecord(ip string) error {
	record := LockRecord{
		IP:       ip,
		LockTime: time.Now().Add(common.LockInterval).Unix(),
	}
	return c.db.Model(LockRecord{}).Create(record).Error
}
