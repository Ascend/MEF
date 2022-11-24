// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector operation table ConnInfo according to the request body
package edgeconnector

import (
	"edge-manager/pkg/util"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// UpdateTableConnInfo updates item in table conn_infos
func UpdateTableConnInfo(msg *model.Message) *ConnInfo {
	if !util.CheckInnerMsg(msg) {
		hwlog.RunLog.Error("message receive from module is invalid")
		return nil
	}
	updateConnInfo, ok := msg.GetContent().(UpdateConnInfo)
	defer common.ClearSliceByteMemory(updateConnInfo.Password)
	if !ok {
		hwlog.RunLog.Error("convert to UpdateConnInfo failed")
		return nil
	}
	if err := updateConnInfo.checkBaseInfo(); err != nil {
		return nil
	}
	// Password encryption
	node := &ConnInfo{
		Address:   updateConnInfo.Address,
		Port:      updateConnInfo.Port,
		Username:  updateConnInfo.UserName,
		Password:  updateConnInfo.Password,
		CreatedAt: time.Now().Format(TimeFormat),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	defer common.ClearSliceByteMemory(node.Password)
	count, err := getItemCount(ConnInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get item count in table conn_infos error: %v", err)
		return nil
	}
	if count == 0 {
		err = createConnInfoDb(node)
		if err != nil {
			hwlog.RunLog.Errorf("create an item in table conn_infos error: %v", err)
			return nil
		}
		return node
	}
	if err = updateInConnInfo(node); err != nil {
		hwlog.RunLog.Errorf("update connection info in table conn_infos failed, error: %v", err)
		return nil
	}
	hwlog.RunLog.Info("update connection info in table conn_infos success")
	return node
}
