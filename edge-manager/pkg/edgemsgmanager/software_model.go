// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager software manager info db module
package edgemsgmanager

import (
	"edge-manager/pkg/database"
)

func createSfwMgrInfoDb(sfwMgrInfo *SoftwareMgrInfo) error {
	return database.GetDb().Model(SoftwareMgrInfo{}).Create(sfwMgrInfo).Error
}

func updateInSfwMgrInfo(sfwMgrInfo *SoftwareMgrInfo) error {
	if err := database.GetDb().Model(SoftwareMgrInfo{}).Where("id = 1").Updates(&sfwMgrInfo).Error; err != nil {
		return err
	}
	return nil
}

func readInSfwMgrInfo(sfwMgrInfo *SoftwareMgrInfo) error {
	if err := database.GetDb().Model(SoftwareMgrInfo{}).Where("id = 1").Find(&sfwMgrInfo).Error; err != nil {
		return err
	}
	return nil
}
