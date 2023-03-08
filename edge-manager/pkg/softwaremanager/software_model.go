// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager for db operate
package softwaremanager

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"edge-manager/pkg/database"
	"huawei.com/mindxedge/base/common"
)

var (
	repositoryInitOnce sync.Once
	sftRepository      sftRepositoryItf
)

// sftRepositoryImpl app service struct
type sftRepositoryImpl struct {
	db *gorm.DB
}

// SftRepository for app method to operate db
type sftRepositoryItf interface {
	insertOrUpdate(key string, value string) error
	query(key string) (string, error)
}

// sftRepositoryInstance returns the singleton instance of software service
func sftRepositoryInstance() sftRepositoryItf {
	repositoryInitOnce.Do(func() {
		sftRepository = &sftRepositoryImpl{db: database.GetDb()}
	})
	return sftRepository
}

func (a *sftRepositoryImpl) insertOrUpdate(key string, value string) error {
	now := time.Now().Format(common.TimeFormat)
	softwareInfo := SoftwareInfo{
		Key:       key,
		Value:     value,
		UpdatedAt: now,
	}
	var count int64
	if err := database.GetDb().Model(SoftwareInfo{}).Where(SoftwareInfo{Key: key}).Count(&count).
		Error; err != nil {
		return fmt.Errorf("count software info failed: %v", err)
	}
	if count > 0 {
		if err := database.GetDb().Model(&softwareInfo).Updates(softwareInfo).Error; err != nil {
			return fmt.Errorf("update software info failed: %v", err)
		}
	} else {
		softwareInfo.CreatedAt = now
		if err := database.GetDb().Create(&softwareInfo).Error; err != nil {
			return fmt.Errorf("create software info failed: %v", err)
		}
	}
	return nil
}

func (a *sftRepositoryImpl) query(key string) (string, error) {
	var softwareInfo SoftwareInfo

	if err := database.GetDb().Where(SoftwareInfo{Key: key}).First(&softwareInfo).Error; err != nil &&
		err != gorm.ErrRecordNotFound {
		return "", fmt.Errorf("query software info failed: %v", err)
	}

	return softwareInfo.Value, nil
}
