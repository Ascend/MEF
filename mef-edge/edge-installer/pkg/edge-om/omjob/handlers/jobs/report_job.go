// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package jobs consist of some jobs used by edge-om
package jobs

import (
	"context"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
)

const reportCheckInterval = 45 * time.Second

// StartReportJob method for StartReportJob
func StartReportJob(ctx context.Context) {
	go func() {
		tick := time.NewTicker(reportCheckInterval)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				hwlog.RunLog.Warn("report job stop")
				return
			case <-tick.C:
				ReportCapability()
			}
		}
	}()
}

// ReportCapability method for report capability
func ReportCapability() {
	hwlog.RunLog.Info("begin report edge capability")
	var info config.StaticInfo
	info.ProductCapabilityEdge = config.GetCapabilityMgr().GetCaps()
	newResponse, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("get new response failed: %v", err)
		return
	}
	newResponse.Header.Version = ""
	newResponse.SetRouter(constants.ModEdgeOm, constants.InnerClient, constants.OptUpdate, constants.ResStatic)
	if err = newResponse.FillContent(info, true); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return
	}
	newResp, err := common.MsgOutProcess(newResponse)
	if err != nil {
		hwlog.RunLog.Errorf("message out process failed, error:%v", err)
		return
	}
	newResp.KubeEdgeRouter.Source = constants.SourceHardware
	if config.NetMgr.NetType == constants.MEF {
		newResp.KubeEdgeRouter.Group = constants.ModEdgeHub
		newResp.KubeEdgeRouter.Operation = constants.OptReport
	} else {
		newResp.KubeEdgeRouter.Group = constants.GroupHub
		newResp.KubeEdgeRouter.Operation = constants.OptUpdate
	}
	newResp.KubeEdgeRouter.Resource = constants.ResStatic
	if err := modulemgr.SendAsyncMessage(newResp); err != nil {
		hwlog.RunLog.Errorf("send async message to device-om failed, error:%v", err)
		return
	}
	hwlog.RunLog.Info("report edge capability success")
}
