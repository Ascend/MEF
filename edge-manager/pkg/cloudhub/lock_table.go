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

	"edge-manager/pkg/constants"
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
	LockTime int64 `gorm:"type:integer"`
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
	isLock() (bool, error)
	deleteLockRecord() error
	UnlockRecords() (int64, error)
	updateLockTime() error
	getFailedRecord(string) (*AuthFailedRecord, error)
	createFailedRecord(string) error
	updateFailedRecord(string) error
	deleteOneFailedRecord(string) error
	deleteAllFailedRecord() error
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

func (c *lockRepositoryImpl) isLock() (bool, error) {
	lockInfo, err := c.getLockRecord()
	if err != nil {
		return true, err
	}
	if lockInfo == nil {
		return false, nil
	}
	if lockInfo.LockTime < time.Now().Unix() {
		if err := c.deleteLockRecord(); err != nil {
			return false, err
		}
		hwlog.RunLog.Info("token is unlock")
		hwlog.OpLog.Infof("[%s@%s] token is unlock", constants.MefCenterUserName, constants.LocalHost)
		return false, nil
	}

	if err := c.updateLockTime(); err != nil {
		return true, fmt.Errorf("update lock time error")
	}
	return true, nil
}

func (c *lockRepositoryImpl) getLockRecord() (*LockRecord, error) {
	var lockInfo []LockRecord
	err := c.db().Model(LockRecord{}).Find(&lockInfo).Error
	if err != nil || len(lockInfo) > 1 {
		hwlog.RunLog.Errorf("get lock record error %v", err)
		return nil, errors.New("get lock record error")
	}
	if len(lockInfo) == 0 {
		return nil, nil
	}
	return &lockInfo[0], nil
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
	if err := c.lock(ip); err != nil {
		return fmt.Errorf("lock %s failed", ip)
	}
	return nil
}

func (c *lockRepositoryImpl) lock(ip string) error {
	recordTime := time.Now().Add(common.LockInterval).Unix()

	return database.Transaction(c.db(), func(tx *gorm.DB) error {
		var lockInfo []LockRecord
		if err := tx.Model(LockRecord{}).Find(&lockInfo).Error; err != nil {
			hwlog.RunLog.Error("get lock info error")
			return err
		}
		if len(lockInfo) == 0 {
			if err := tx.Model(LockRecord{}).Create(LockRecord{LockTime: recordTime}).Error; err != nil {
				hwlog.RunLog.Error("create lock info error")
				return err
			}
			hwlog.OpLog.Warnf("[%s@%s] %s has too much token auth failed record, lock token auth function",
				constants.MefCenterUserName, constants.LocalHost, ip)
			hwlog.RunLog.Warnf("%s has too much token auth failed record, lock token auth function", ip)
		} else {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(LockRecord{}).
				Updates(map[string]interface{}{"LockTime": recordTime}).Error; err != nil {
				hwlog.RunLog.Error("update lock info error")
				return err
			}
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AuthFailedRecord{}).Error; err != nil {
			hwlog.RunLog.Error("delete failed record error")
			return err
		}
		return nil
	})
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

func (c *lockRepositoryImpl) deleteOneFailedRecord(ip string) error {
	if err := c.db().Model(AuthFailedRecord{}).Where("ip = ?", ip).Delete(&AuthFailedRecord{}).Error; err != nil {
		return fmt.Errorf("delete one failed record(%s) from db error", ip)
	}
	return nil
}

func (c *lockRepositoryImpl) deleteAllFailedRecord() error {
	if err := c.db().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AuthFailedRecord{}).Error; err != nil {
		return errors.New("delete failed record from db error")
	}
	return nil
}

func (c *lockRepositoryImpl) UnlockRecords() (int64, error) {
	stmt := c.db().Model(LockRecord{}).Where("lock_time < ?", time.Now().Unix()).Delete(LockRecord{})
	return stmt.RowsAffected, stmt.Error
}

func (c *lockRepositoryImpl) deleteLockRecord() error {
	if err := c.db().Session(&gorm.Session{AllowGlobalUpdate: true}).
		Model(LockRecord{}).Delete(&LockRecord{}).Error; err != nil {
		return fmt.Errorf("delete lock info error")
	}
	return nil
}

func (c *lockRepositoryImpl) updateLockTime() error {
	updatedColumns := map[string]interface{}{
		"LockTime": time.Now().Add(common.LockInterval).Unix(),
	}
	return c.db().Session(&gorm.Session{AllowGlobalUpdate: true}).Model(LockRecord{}).Updates(updatedColumns).Error
}
