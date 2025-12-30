// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package job consists of jobs needed. jobs related to connection should in this file
package job

import (
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
)

const (
	// PodStatusInterval interval to report pod status
	PodStatusInterval = 30 * time.Second
)

// SyncPodStatus tell FD the pod's status at a fixed interval which query from db
func SyncPodStatus() error {
	msg, err := common.NewFDPodStatusMsg("metaManager", "resource",
		constants.OptUpdate, constants.ResPodStatus)
	if err != nil {
		hwlog.RunLog.Errorf("sync pod status failed: %v", err)
		return err
	}
	if err := modulemgr.SendAsyncMessage(msg); err != nil {
		hwlog.RunLog.Errorf("send sync pod status msg failed: %v", err)
		return err
	}
	return nil
}

// SyncNodeStatus tell FD the node's status when mef is connected to device-om
func SyncNodeStatus() error {
	msg, err := common.NewFDNodeStatusMsg(constants.ModDeviceOm, constants.ModDeviceOm,
		constants.OptUpdate, constants.ResNodeStatus)
	if err != nil {
		hwlog.RunLog.Errorf("sync node status failed: %v", err)
		return err
	}
	if err := modulemgr.SendAsyncMessage(msg); err != nil {
		hwlog.RunLog.Errorf("send sync node status msg failed: %v", err)
		return err
	}
	return nil
}
