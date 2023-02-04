// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller handler
package edgeinstaller

import (
	"fmt"
	"sync"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/handlerbase"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

var handlerMgr handlerbase.HandlerMgr
var regOnce sync.Once
var registerInfoList = []handlerbase.RegisterInfo{
	{MsgOpt: common.OptGet, MsgRes: common.ResEdgeCoreConfig, Handler: new(configHandler)},
	{MsgOpt: common.OptGet, MsgRes: common.ResDownLoadSoftware, Handler: new(downloadHandler)},
	{MsgOpt: common.OptPost, MsgRes: common.ResDownLoadSoftware, Handler: new(upgradeHandler)},
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
	respMsg.SetRouter(common.EdgeInstallerName, common.NodeMsgManagerName, common.OptPost, msg.GetResource())

	if err = modulemanager.SendMessage(respMsg); err != nil {
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

	if err = modulemanager.SendAsyncMessage(newResponse); err != nil {
		hwlog.RunLog.Errorf("edge-installer send sync message failed, error: %v", err)
		return fmt.Errorf("edge-installer send sync message failed, error: %v", err)
	}

	return nil
}
