package handlerbase

import (
	"fmt"
	"sync"

	"huawei.com/mindxedge/base/modulemanager/model"
)

type HandlerMgr struct {
	handLock    sync.Mutex
	handlersMap map[string]HandleBase
}

func (hm *HandlerMgr) Register(regHandler RegisterInfo) {
	hm.handLock.Lock()
	defer hm.handLock.Unlock()
	if hm.handlersMap == nil {
		hm.handlersMap = make(map[string]HandleBase)
	}
	hm.handlersMap[regHandler.MsgOpt+":"+regHandler.MsgRes] = regHandler.Handler
}

func (hm *HandlerMgr) Unregister(msgType string) {
	hm.handLock.Lock()
	defer hm.handLock.Unlock()
	if hm.handlersMap == nil {
		return
	}
	delete(hm.handlersMap, msgType)
}

func (hm *HandlerMgr) Process(msg *model.Message) error {
	hm.handLock.Lock()
	defer hm.handLock.Unlock()
	msgOpt := msg.GetOption()
	msgRes := msg.GetResource()
	key := msgOpt + ":" + msgRes
	handler := hm.handlersMap[key]
	if handler == nil {
		return fmt.Errorf("no register msg Handler[MsgOpt=%v, MsgRes=%v]", msgOpt, msgRes)
	}
	postHandler, ok := handler.(PostHandleBase)
	if ok {
		return hm.doPost(msg, postHandler)
	}
	return handler.Handle(msg)
}

func (hm *HandlerMgr) doPost(msg *model.Message, handler PostHandleBase) error {
	err := handler.Parse(msg)
	if err != nil {
		handler.PrintOpLogFail()
		return err
	}
	err = handler.Check(msg)
	if err != nil {
		handler.PrintOpLogFail()
		return err
	}
	err = handler.Handle(msg)
	if err != nil {
		handler.PrintOpLogFail()
		return err
	}
	handler.PrintOpLogOk()
	return nil
}
