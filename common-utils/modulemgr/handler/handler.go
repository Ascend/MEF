// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handler for
package handler

import (
	"errors"
	"fmt"
	"sync"

	"huawei.com/mindx/common/modulemgr/model"
)

// MsgHandler message handler manager
type MsgHandler struct {
	handlersMap map[string]HandleBase
	handLock    sync.Mutex
}

// Register registered handler info
func (hm *MsgHandler) Register(regHandler RegisterInfo) {
	if hm == nil {
		return
	}
	hm.handLock.Lock()
	defer hm.handLock.Unlock()
	if hm.handlersMap == nil {
		hm.handlersMap = make(map[string]HandleBase)
	}
	hm.handlersMap[regHandler.MsgOpt+":"+regHandler.MsgRes] = regHandler.Handler
}

// Unregister unregistered handler info
func (hm *MsgHandler) Unregister(msgType string) {
	if hm == nil {
		return
	}
	hm.handLock.Lock()
	defer hm.handLock.Unlock()
	if hm.handlersMap == nil {
		return
	}
	delete(hm.handlersMap, msgType)
}

// Process execute handle process
func (hm *MsgHandler) Process(msg *model.Message) error {
	if hm == nil {
		return errors.New("msg handler is nil")
	}
	hm.handLock.Lock()
	defer hm.handLock.Unlock()
	msgOpt := msg.GetOption()
	msgRes := msg.GetResource()
	key := msgOpt + ":" + msgRes
	handler, ok := hm.handlersMap[key]
	if !ok || handler == nil {
		return fmt.Errorf("no register msg Handler[MsgOpt=%v, MsgRes=%v]", msgOpt, msgRes)
	}
	postHandler, ok := handler.(PostHandleBase)
	if ok {
		return hm.doPost(msg, postHandler)
	}
	return handler.Handle(msg)
}

func (hm *MsgHandler) doPost(msg *model.Message, handler PostHandleBase) error {
	if err := handler.Parse(msg); err != nil {
		handler.PrintOpLogFail()
		return err
	}
	if err := handler.Check(msg); err != nil {
		handler.PrintOpLogFail()
		return err
	}
	if err := handler.Handle(msg); err != nil {
		handler.PrintOpLogFail()
		return err
	}
	handler.PrintOpLogOk()
	return nil
}
