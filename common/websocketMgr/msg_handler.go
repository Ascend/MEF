package websocket

import (
	"encoding/json"
	"sync"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type WsMsgHandler struct {
	handLock    sync.Mutex
	handlersMap map[string]string
}

func (wh *WsMsgHandler) Register(regHandler RegisterModuleInfo) {
	wh.handLock.Lock()
	defer wh.handLock.Unlock()
	if wh.handlersMap == nil {
		wh.handlersMap = make(map[string]string)
	}
	wh.handlersMap[regHandler.MsgOpt+":"+regHandler.MsgRes] = regHandler.ModuleName
}

func (wh *WsMsgHandler) handleMsg(msgBytes []byte) {
	wh.handLock.Lock()
	defer wh.handLock.Unlock()

	var msg *model.Message
	err := json.Unmarshal(msgBytes, &msg)
	if err != nil {
		hwlog.RunLog.Errorf("Unmarshal message failed: %v\n", err)
		return
	}
	msgOpt := msg.GetOption()
	msgRes := msg.GetResource()
	key := msgOpt + ":" + msgRes
	moduleName := wh.handlersMap[key]
	if moduleName == "" {
		hwlog.RunLog.Errorf("no register msg Handler[MsgOpt=%v, MsgRes=%v]", msgOpt, msgRes)
		return
	}
	msg.SetRouter("websocket", moduleName, msgOpt, msgRes)
	err = modulemanager.SendMessage(msg)
	if err != nil {
		hwlog.RunLog.Errorf("send module message failed: %v\n", err)
		return
	}
}
