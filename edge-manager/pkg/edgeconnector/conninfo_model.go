// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the table conn_infos operation
package edgeconnector

import (
	"time"

	"edge-manager/pkg/database"

	"huawei.com/mindxedge/base/common"
)

func createConnInfoDb(connInfo *ConnInfo) error {
	return database.GetDb().Model(ConnInfo{}).Create(connInfo).Error
}

func updateInConnInfo(node *ConnInfo) error {
	defer common.ClearSliceByteMemory(node.Password)
	if err := database.GetDb().Model(ConnInfo{}).Where("id = 1").Updates(&ConnInfo{
		Address:   node.Address,
		Port:      node.Port,
		Username:  node.Username,
		Password:  node.Password,
		UpdatedAt: time.Now().Format(TimeFormat),
	}).Error; err != nil {
		return err
	}

	return nil
}
