// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub for
package cloudhub

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

const (
	lockTime = 5
)

// AuthFailedRecord token error then record ip
type AuthFailedRecord struct {
	IP         string `gorm:"type:char(32);unique;not null"`
	ErrorTimes int    `gorm:"type:integer;not null"`
}

// LockRecord lock ip record
type LockRecord struct {
	IP       string `gorm:"type:char(32);unique;not null"`
	LockTime int64  `gorm:"type:integer"`
}

var (
	repositoryInitOnce sync.Once
	lockRepository     LockRepository
)

type lockRepositoryImpl struct {
}

// LockRepository for config method to operate db
type LockRepository interface {
	recordFailed(string) error
	isLock(string) (bool, error)
	deleteOneLockRecord(string) error
	findUnlockRecords() ([]LockRecord, error)
	UnlockRecords(string) error
	updateLockTime(string) error
	getFailedRecord(string) (*AuthFailedRecord, error)
	createFailedRecord(string) error
	updateFailedRecord(string) error
	createLockRecord(string) error
	deleteFailedRecord(string) error
	authPass(string) error
}

// LockRepositoryInstance returns the singleton instance of token lock service
func LockRepositoryInstance() LockRepository {
	repositoryInitOnce.Do(func() {
		lockRepository = &lockRepositoryImpl{}
	})
	return lockRepository
}

func (c *lockRepositoryImpl) db() *gorm.DB {
	return database.GetDb()
}

func (c *lockRepositoryImpl) isLock(ip string) (bool, error) {
	var lockInfo LockRecord
	err := c.db().Model(LockRecord{}).Where("ip = ?", ip).First(&lockInfo).Error
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
		hwlog.OpLog.Infof("edge (%s) is unlock", ip)
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

	if record.ErrorTimes < lockTime {
		return c.updateFailedRecord(ip)
	}
	if err := c.createLockRecord(ip); err != nil {
		return fmt.Errorf("create lock record to db error ip(%s)", ip)
	}
	hwlog.OpLog.Warnf("%s has too much auth failed record, lock this edge device", ip)
	if err := c.deleteFailedRecord(ip); err != nil {
		return err
	}
	return nil
}

func (c *lockRepositoryImpl) authPass(ip string) error {
	if err := c.deleteFailedRecord(ip); err != nil {
		return err
	}
	return c.deleteOneLockRecord(ip)
}

func (c *lockRepositoryImpl) getFailedRecord(ip string) (*AuthFailedRecord, error) {
	var lockInfo AuthFailedRecord
	err := c.db().Model(AuthFailedRecord{}).Where("ip = ?", ip).First(&lockInfo).Error
	return &lockInfo, err
}

func (c *lockRepositoryImpl) createFailedRecord(ip string) error {
	record := AuthFailedRecord{
		IP:         ip,
		ErrorTimes: 1,
	}
	if createErr := c.db().Model(AuthFailedRecord{}).Create(record).Error; createErr != nil {
		return fmt.Errorf("create auth failed record to db error, ip(%s)", ip)
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
	if err := c.db().Model(AuthFailedRecord{}).Where("ip = ?", ip).UpdateColumns(record).Error; err != nil {
		return fmt.Errorf("update auth failed record to db error, ip(%s)", ip)
	}
	return nil
}

func (c *lockRepositoryImpl) deleteFailedRecord(ip string) error {
	if err := c.db().Model(AuthFailedRecord{}).Where("ip = ?", ip).Delete(&AuthFailedRecord{}).Error; err != nil {
		return fmt.Errorf("delete failed record(%s) from db error", ip)
	}
	return nil
}

func (c *lockRepositoryImpl) findUnlockRecords() ([]LockRecord, error) {
	var unlockIP []LockRecord
	return unlockIP, c.db().Model(LockRecord{}).Where("lock_time < ?", time.Now().Unix()).Find(&unlockIP).Error
}

func (c *lockRepositoryImpl) UnlockRecords(ip string) error {
	return c.db().Model(LockRecord{}).Where("lock_time < ? and ip = ?", time.Now().Unix(), ip).Delete(LockRecord{}).Error
}

func (c *lockRepositoryImpl) deleteOneLockRecord(ip string) error {
	if err := c.db().Model(LockRecord{}).Where("ip = ?", ip).Delete(&LockRecord{}).Error; err != nil {
		return fmt.Errorf("delete lock info from ip:%s error", ip)
	}
	return nil
}

func (c *lockRepositoryImpl) updateLockTime(ip string) error {
	updatedColumns := map[string]interface{}{
		"LockTime": time.Now().Add(common.LockInterval).Unix(),
	}
	return c.db().Model(LockRecord{}).Where("ip = ?", ip).UpdateColumns(updatedColumns).Error
}

func (c *lockRepositoryImpl) createLockRecord(ip string) error {
	record := LockRecord{
		IP:       ip,
		LockTime: time.Now().Add(common.LockInterval).Unix(),
	}
	return c.db().Model(LockRecord{}).Create(record).Error
}
