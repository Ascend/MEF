// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

import (
	"encoding/json"
	"sync"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// WsMsgHandler msg handler
type WsMsgHandler struct {
	handLock    sync.Mutex
	handlersMap map[string]string
}

func (wh *WsMsgHandler) register(regHandler RegisterModuleInfo) {
	wh.handLock.Lock()
	defer wh.handLock.Unlock()
	if wh.handlersMap == nil {
		wh.handlersMap = make(map[string]string)
	}
	wh.handlersMap[regHandler.MsgOpt+":"+regHandler.MsgRes] = regHandler.ModuleName
}

func (wh *WsMsgHandler) handleMsg(msgBytes []byte) {
	var msg model.Message
	err := json.Unmarshal(msgBytes, &msg)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal message failed, error: %v", err)
		return
	}
	if msg.GetParentId() == "" {
		msgOpt := msg.GetOption()
		msgRes := msg.GetResource()
		key := msgOpt + ":" + msgRes
		moduleName := wh.handlersMap[key]
		if moduleName == "" {
			hwlog.RunLog.Errorf("no register msg Handler [MsgOpt = %v, MsgRes = %v]", msgOpt, msgRes)
			return
		}
		msg.SetRouter("websocket", moduleName, msgOpt, msgRes)
	}
	err = modulemanager.SendMessage(&msg)
	if err != nil {
		hwlog.RunLog.Errorf("send module message failed, error: %v", err)
		return
	}
}
