// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr for deal every handler
package handlermgr

import (
	"sync"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/handler"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

var handlerMgr handler.MsgHandler
var regOnce sync.Once
var registerInfoList = []handler.RegisterInfo{
	{MsgOpt: constants.OptGet, MsgRes: constants.ResConfig, Handler: new(getConfigHandler)},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.ResImageCertInfo, Handler: new(saveCertHandler)},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.ResNpuSharing, Handler: new(npuSharingHandler)},
	{MsgOpt: constants.OptRestart, MsgRes: constants.ActionPod, Handler: new(restartPodHandler)},
	{MsgOpt: constants.OptRaw, MsgRes: constants.ActionModelFiles, Handler: new(operateModelFileHandler)},
	{MsgOpt: constants.OptUpdate, MsgRes: constants.ActionModelFiles, Handler: new(modelFileHandler)},
}

// GetHandlerMgr get handler manager
func GetHandlerMgr() *handler.MsgHandler {
	regOnce.Do(registerHandler)
	return &handlerMgr
}

func registerHandler() {
	handlerMgr = handler.MsgHandler{}
	for _, reg := range registerInfoList {
		handlerMgr.Register(reg)
	}
}

func sendHandlerReplyMsg(msg *model.Message) error {
	msg.SetRouter(constants.ModEdgeOm, constants.InnerClient, msg.GetOption(), msg.GetResource())
	err := modulemgr.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}
