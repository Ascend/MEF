// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller handler
package edgeinstaller

import (
	"fmt"
	"sync"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/handlerbase"
)

var handlerMgr handlerbase.HandlerMgr
var regOnce sync.Once
var registerInfoList = []handlerbase.RegisterInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResEdgeCoreConfig, Handler: new(configHandler)},
	{MsgOpt: common.OptPost, MsgRes: common.ResDownLoadSoftware, Handler: new(upgradeHandler)},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadCert, Handler: new(certHandler)},
}

// GetHandlerMgr get handler manager
func GetHandlerMgr() *handlerbase.HandlerMgr {
	regOnce.Do(registerHandler)
	return &handlerMgr
}

func registerHandler() {
	handlerMgr = handlerbase.HandlerMgr{}
	for _, reg := range registerInfoList {
		handlerMgr.Register(reg)
	}
}

func sendMessage(msg *model.Message, resp string) error {
	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("edge-installer new message failed, error: %v", err)
		return fmt.Errorf("edge-installer new message failed, error: %v", err)
	}

	respMsg.SetNodeId(msg.GetNodeId())
	respMsg.FillContent(resp)
	respMsg.SetRouter(common.NodeMsgManagerName, common.NodeMsgManagerName, common.OptPost, msg.GetResource())

	if err = modulemgr.SendMessage(respMsg); err != nil {
		hwlog.RunLog.Errorf("edge-installer send message failed, error: %v", err)
		return fmt.Errorf("edge-installer send message failed, error: %v", err)
	}

	return nil
}

func sendResponse(msg *model.Message, resp string) error {
	newResponse, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("edge-installer new response failed, error: %v", err)
		return fmt.Errorf("edge-installer new response failed, error: %v", err)
	}

	newResponse.FillContent(resp)
	newResponse.SetRouter(common.EdgeInstallerName, common.NodeMsgManagerName, common.OptPost, msg.GetResource())

	if err = modulemgr.SendAsyncMessage(newResponse); err != nil {
		hwlog.RunLog.Errorf("edge-installer send sync message failed, error: %v", err)
		return fmt.Errorf("edge-installer send sync message failed, error: %v", err)
	}

	return nil
}
