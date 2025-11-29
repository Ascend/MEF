// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package modulemgr

import (
	"encoding/json"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/limiter"
	"huawei.com/mindx/common/modulemgr/model"
)

const (
	handleSyncMsgTimeout = 30 * time.Second
)

// RegisterModuleInfo register module info
type RegisterModuleInfo struct {
	Src        string
	MsgOpt     string
	MsgRes     string
	ModuleName string
	// if no limit config set, no request limitation on this message
	Rps   float64
	Burst int
}

// GetMessageKey implementation for MessageHandlerIntf, get message identifier
func (msg *RegisterModuleInfo) GetMessageKey() string {
	return msg.MsgOpt + ":" + msg.MsgRes
}

// GetMessageHandler implementation for MessageHandlerIntf, get message handler module name
func (msg *RegisterModuleInfo) GetMessageHandler() string {
	return msg.ModuleName
}

// GetRpsLimiter implementation for MessageHandlerIntf, get related rps limiter, if not set, return nil
func (msg *RegisterModuleInfo) GetRpsLimiter() limiter.IndependentLimiter {
	if msg.Rps > 0 && msg.Burst > 0 {
		return limiter.NewRpsLimiter(msg.Rps, msg.Burst)
	}
	return nil
}

// MsgOutModuleRouter find a dest websocket connection for msg
type MsgOutModuleRouter struct {
	Src      string
	MsgOpt   string
	MsgRes   string
	ConnName string
}

// MsgHandler message handler
type MsgHandler struct {
	handLock           sync.Mutex
	handlersMap        map[string]string
	handlersLimiterMap map[string]limiter.IndependentLimiter
}

// Register - register a handler module and related limiter (if set) for a message
func (wh *MsgHandler) Register(regHandler MessageHandlerIntf) {
	wh.handLock.Lock()
	defer wh.handLock.Unlock()
	if wh.handlersMap == nil {
		wh.handlersMap = make(map[string]string)
	}
	handlerKey := regHandler.GetMessageKey()
	wh.handlersMap[handlerKey] = regHandler.GetMessageHandler()

	// init rps limiter for each message
	rpsLimiter := regHandler.GetRpsLimiter()
	if rpsLimiter == nil {
		return
	}
	if wh.handlersLimiterMap == nil {
		wh.handlersLimiterMap = make(map[string]limiter.IndependentLimiter)
	}
	wh.handlersLimiterMap[handlerKey] = rpsLimiter
}

// HandleMsg - process a message by it's registered handler
func (wh *MsgHandler) HandleMsg(msgBytes []byte, pi model.MsgPeerInfo) []byte {
	var msg model.Message
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		hwlog.RunLog.Errorf("unmarshal message failed, error: %v", err)
		return nil
	}
	msg.SetPeerInfo(pi)
	if msg.GetParentId() == "" {
		// key rule [opt + ":" + res] must be sync with MessageHandlerIntf GetMessageKey() implementation
		msgOpt := msg.GetOption()
		msgRes := msg.GetResource()
		key := msgOpt + ":" + msgRes

		if rpsLimiter, ok := wh.handlersLimiterMap[key]; ok && !rpsLimiter.Allow() {
			return nil
		}

		moduleName, ok := wh.handlersMap[key]
		if !ok {
			hwlog.RunLog.Errorf("no register msg handler [MsgOpt = %v, MsgRes = %v]", msgOpt, msgRes)
			return nil
		}
		msg.SetRouter("websocket", moduleName, msgOpt, msgRes)
	}
	// sync request message
	if msg.GetIsSync() && msg.GetParentId() == "" {
		ret, err := SendSyncMessage(&msg, handleSyncMsgTimeout)
		if err != nil {
			hwlog.RunLog.Errorf("get result of msg [option: %s resource: %s] from [%s] failed: %v", msg.GetOption(),
				msg.GetResource(), msg.GetNodeId(), err)
			return nil
		}
		data, err := json.Marshal(ret)
		if err != nil {
			hwlog.RunLog.Errorf("marshal message failed: %v", err)
			return nil
		}
		return data
	}
	// async message and sync resp message
	if err := SendMessage(&msg); err != nil {
		hwlog.RunLog.Errorf("send module message failed, error: %v", err)
	}

	return nil
}
