// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector operation table ConnInfo according to the request body
package edgeconnector

import (
	"time"

	"edge-manager/pkg/database"
	"edge-manager/pkg/util"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// UpdateTableConnInfo updates item in table conn_infos
func UpdateTableConnInfo(message *model.Message) common.RespMsg {
	updateConnInfo, resp := checkUpdateConnInfoMsg(message)
	if resp.Status != common.Success {
		hwlog.RunLog.Error("check message for updating conn_infos failed")
		return common.RespMsg{Status: common.Success, Msg: "check message for updating conn_infos failed", Data: nil}
	}

	// todo password encryption
	node := &ConnInfo{
		Address:   updateConnInfo.Address,
		Port:      updateConnInfo.Port,
		Username:  updateConnInfo.Username,
		Password:  updateConnInfo.Password,
		CreatedAt: time.Now().Format(TimeFormat),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	defer common.ClearSliceByteMemory(node.Password)
	count, err := database.GetItemCount(ConnInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get item count in table conn_infos error: %v", err)
		return common.RespMsg{Status: "", Msg: "get item count in table conn_infos error", Data: nil}
	}

	if count == 0 {
		err = createConnInfoDb(node)
		if err != nil {
			hwlog.RunLog.Errorf("create an item in table conn_infos error: %v", err)
			return common.RespMsg{Status: "", Msg: "create an item in table conn_infos error", Data: nil}
		}
		return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
	}

	if err = updateInConnInfo(node); err != nil {
		hwlog.RunLog.Errorf("update connection info in table conn_infos failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: "db update error", Data: nil}
	}

	hwlog.RunLog.Info("update connection info in table conn_infos success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func checkUpdateConnInfoMsg(message *model.Message) (*UpdateConnInfo, common.RespMsg) {
	if !util.CheckInnerMsg(message) {
		hwlog.RunLog.Error("message receive from module is invalid")
		return nil, common.RespMsg{Status: "", Msg: "check inner message error", Data: nil}
	}

	updateConnInfo, ok := message.GetContent().(UpdateConnInfo)
	defer common.ClearSliceByteMemory(updateConnInfo.Password)
	if !ok {
		hwlog.RunLog.Error("convert to UpdateConnInfo failed")
		return nil, common.RespMsg{Status: "", Msg: "convert to UpdateConnInfo error", Data: nil}
	}

	if err := updateConnInfo.checkBaseInfo(); err != nil {
		hwlog.RunLog.Error("check base info error")
		return nil, common.RespMsg{Status: "", Msg: "check base info error", Data: nil}
	}

	return &updateConnInfo, common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// UpdateUserConnInfo updates item in table conn_infos
func UpdateUserConnInfo(updateInfo UpdateInfo) common.RespMsg {
	// todo password encryption
	node := &ConnInfo{
		Username:  updateInfo.Username,
		Password:  []byte(updateInfo.Password),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	defer common.ClearSliceByteMemory(node.Password)
	count, err := database.GetItemCount(ConnInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get item count in table conn_infos error: %v", err)
		return common.RespMsg{Status: "", Msg: "get item count in table conn_infos error", Data: nil}
	}

	if count == 0 {
		if err = createConnInfoDb(node); err != nil {
			hwlog.RunLog.Errorf("create an item in table conn_infos error: %v", err)
			return common.RespMsg{Status: "", Msg: "create an item in table conn_infos error", Data: nil}
		}
		return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
	}

	if err = updateUserInfoInConnInfo(node); err != nil {
		hwlog.RunLog.Errorf("update user info in table conn_infos failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: "db update error", Data: nil}
	}

	hwlog.RunLog.Info("update user info in table conn_infos success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
