// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller software manager info db service
package edgeinstaller

import (
	"time"

	"edge-manager/pkg/database"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

// UpgradeTableSfwInfo updates software manager info
func UpgradeTableSfwInfo(updateTableSfwInfo *SoftwareMgrInfo) common.RespMsg {
	var resp common.RespMsg
	resp = CheckUpdateTableSfwInfo(updateTableSfwInfo)
	if resp.Status != common.Success {
		hwlog.RunLog.Error("check updating table software manager info error")
		return common.RespMsg{Status: "", Msg: "check updating table software manager info error", Data: nil}
	}

	count, err := database.GetItemCount(SoftwareMgrInfo{})
	if err != nil {
		hwlog.RunLog.Errorf("get item count in table software manager info error: %v", err)
		return common.RespMsg{Status: "", Msg: "get item count in table software manager info error", Data: nil}
	}

	if count == 0 {
		err = createSfwMgrInfoDb(updateTableSfwInfo)
		if err != nil {
			hwlog.RunLog.Errorf("create an item in table software manager info error: %v", err)
			return common.RespMsg{Status: "", Msg: "create an item in table software manager info error", Data: nil}
		}
		return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
	}

	if err = updateInSfwMgrInfo(updateTableSfwInfo); err != nil {
		hwlog.RunLog.Errorf("update connection info in table software manager info failed, error: %v", err)
		return common.RespMsg{Status: "", Msg: "update table software manager error", Data: nil}
	}

	hwlog.RunLog.Info("update connection info in table software manager info success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

// CreateTableSfwInfo create table software manager info
func CreateTableSfwInfo() error {
	now := time.Now()
	sfwManagerInfo := &SoftwareMgrInfo{
		Address:   SoftwareIP,
		Port:      SoftwarePort,
		Route:     SoftRoute,
		CreatedAt: now.Format(common.TimeFormat),
		UpdatedAt: now.Format(common.TimeFormat),
	}
	if err := createSfwMgrInfoDb(sfwManagerInfo); err != nil {
		hwlog.RunLog.Errorf("create software manager info table failed, error: %v", err)
		return err
	}

	return nil
}
