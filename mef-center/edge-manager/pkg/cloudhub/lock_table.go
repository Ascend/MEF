// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cloudhub for
package cloudhub

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
)

const (
	lockTime        = 5
	maxAuthFiledNum = 5000
	unlock          = 0
	nolockingFlag   = 0
	lockingFlag     = 1
)

// AuthFailedRecord token error then record ip
type AuthFailedRecord struct {
	IP         string `gorm:"type:char(32);unique;not null"`
	ErrorTimes int    `gorm:"type:integer;not null"`
}

// LockRecord lock ip record
type LockRecord struct {
	IP       string `gorm:"type:char(32)"`
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
	isLock() (bool, error)
	deleteLockRecord() error
	UnlockRecords() error
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
	lockInfo, err := c.getAndInitLockRecord()
	if err != nil {
		return true, err
	}
	if lockInfo == nil || lockInfo.LockTime == unlock {
		return false, nil
	}

	return true, nil
}

func (c *lockRepositoryImpl) getAndInitLockRecord() (*LockRecord, error) {
	var lockInfo []LockRecord
	err := c.db().Model(LockRecord{}).Find(&lockInfo).Error
	if err != nil || len(lockInfo) > 1 {
		hwlog.RunLog.Errorf("get lock record error %v", err)
		return nil, errors.New("get lock record error")
	}
	if len(lockInfo) != 0 {
		return &lockInfo[0], nil
	}
	if err := c.db().Model(LockRecord{}).Create(LockRecord{LockTime: unlock}).Error; err != nil {
		hwlog.RunLog.Error("create lock info error")
		return nil, err
	}
	return nil, nil
}

func (c *lockRepositoryImpl) recordFailed(ip string) error {
	record, err := c.getFailedRecord(ip)
	if err != nil && err != gorm.ErrRecordNotFound {
		hwlog.RunLog.Error("get auth failed record from db error")
		return errors.New("get auth failed record from db error")
	}
	if err == gorm.ErrRecordNotFound {
		count, err := common.GetItemCount(AuthFailedRecord{})
		if err != nil {
			hwlog.RunLog.Error("get auth failed record num error")
			return errors.New("get auth failed record num error")
		}
		if count > maxAuthFiledNum {
			hwlog.RunLog.Errorf("auth failed ip record has exceed %d, lock token auth function", maxAuthFiledNum)
			if err := c.lock(ip); err != nil {
				return fmt.Errorf("lock failed trigger by %s", ip)
			}
			return nil
		}

		return c.createFailedRecord(ip)
	}

	if record.ErrorTimes < lockTime {
		return c.updateFailedRecord(ip)
	}
	if err := c.lock(ip); err != nil {
		return fmt.Errorf("lock failed trigger by %s", ip)
	}
	return nil
}

func (c *lockRepositoryImpl) lock(ip string) error {
	return database.Transaction(c.db(), func(tx *gorm.DB) error {
		var lockInfo []LockRecord
		if err := tx.Model(LockRecord{}).Find(&lockInfo).Error; err != nil {
			hwlog.RunLog.Error("get lock info error")
			return err
		}
		if len(lockInfo) != 0 && lockInfo[0].LockTime == unlock {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(LockRecord{}).
				Updates(map[string]interface{}{"LockTime": time.Now().Unix()}).Error; err != nil {
				hwlog.RunLog.Error("update lock info error")
				return err
			}
		}
		if !atomic.CompareAndSwapInt64(&lockFlag, nolockingFlag, lockingFlag) {
			hwlog.RunLog.Error("token is in locking status, try it later")
			return fmt.Errorf("token is in locking status, try it later")
		}
		go doLock()
		hwlog.OpLog.Warnf("[%s@%s] %s has too much token auth failed record, lock token auth function",
			constants.MefCenterUserName, constants.LocalHost, ip)
		hwlog.RunLog.Warnf("%s has too much token auth failed record, lock token auth function", ip)
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

func (c *lockRepositoryImpl) UnlockRecords() error {
	return c.db().Session(&gorm.Session{AllowGlobalUpdate: true}).Model(LockRecord{}).
		Updates(map[string]interface{}{"LockTime": unlock}).Error
}

func (c *lockRepositoryImpl) deleteLockRecord() error {
	if err := c.db().Session(&gorm.Session{AllowGlobalUpdate: true}).
		Model(LockRecord{}).Delete(&LockRecord{}).Error; err != nil {
		return fmt.Errorf("delete lock info error")
	}
	return nil
}
