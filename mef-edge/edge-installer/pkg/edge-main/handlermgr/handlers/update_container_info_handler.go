// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlers
package handlers

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/edge-main/common"
)

const responseOK = "OK"

type updateContainerInfoHandler struct{}

// Handle updateContainerInfoHandler handle entry
func (uc *updateContainerInfoHandler) Handle(msg *model.Message) error {
	uc.sendResponse(msg, responseOK)

	hwlog.RunLog.Info("start update container info")
	var containerInfo types.UpdateContainerInfo
	if err := msg.ParseContent(&containerInfo); err != nil {
		hwlog.RunLog.Errorf("parse container info content failed, error: %v", err)
		uc.reportContainerInfoProgress("0%", "failed", "input parameter error")
		return errors.New("parse container info content failed")
	}

	updateContainerInfo := UpdateContainerInfo{containerInfo: containerInfo}
	if err := updateContainerInfo.EffectModelFile(); err != nil {
		hwlog.RunLog.Errorf("update container info failed, error: %v", err)
		uc.reportContainerInfoProgress("0%", "failed", err.Error())
		return errors.New("update container info failed")
	}

	hwlog.RunLog.Info("update container info success")
	uc.reportContainerInfoProgress("100%", "success", "")
	return nil
}

func (uc *updateContainerInfoHandler) sendResponse(msg *model.Message, respMsg string) {
	newResp, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("get new response message failed, error: %v", err)
		return
	}
	newResp.Header.IsSync = false
	newResp.Router.Destination = constants.ModDeviceOm
	newResp.Router.Option = constants.OptReport
	newResp.SetKubeEdgeRouter(constants.ControllerModule, constants.HardwareModule, constants.OptReport,
		constants.ActionContainerInfo)
	newResponse, err := common.MsgOutProcess(newResp)
	if err != nil {
		hwlog.RunLog.Errorf("message out process failed, error: %v", err)
		return
	}
	if err = newResponse.FillContent(respMsg); err != nil {
		hwlog.RunLog.Errorf("fill resp content failed: %v", err)
		return
	}
	if err = modulemgr.SendAsyncMessage(newResponse); err != nil {
		hwlog.RunLog.Errorf("send update container info handler response failed, error: %v", err)
	}
}

func (uc *updateContainerInfoHandler) reportContainerInfoProgress(percentage, result, reason string) {
	content := config.ProgressTip{
		Topic:      "container_info",
		Percentage: percentage,
		Result:     result,
		Reason:     reason,
	}
	newMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("get new report message failed, error: %v", err)
		return
	}

	newMsg.Header.ID = newMsg.Header.Id
	newMsg.SetRouter(constants.ModHandlerMgr, constants.ModDeviceOm, constants.OptUpdate, constants.ResConfigResult)
	newMsg.SetKubeEdgeRouter(constants.HardwareModule, constants.GroupHub, constants.OptUpdate,
		constants.ResConfigResult)

	if err = newMsg.FillContent(content, true); err != nil {
		hwlog.RunLog.Errorf("fill data into content failed: %v", err)
		return
	}
	if err = modulemgr.SendAsyncMessage(newMsg); err != nil {
		hwlog.RunLog.Errorf("report update container info handler progress failed, error: %v", err)
	}
}
