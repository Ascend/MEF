package websocketmgr

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type WsCltSender struct {
	proxy NetProxyIntf
}

func (wcs *WsCltSender) SetProxy(proxy NetProxyIntf) {
	wcs.proxy = proxy
}

func (wcs *WsCltSender) Send(msg *model.Message) error {
	data, err := json.Marshal(msg)
	sendMsg := wsMessage{
		MsgType: websocket.TextMessage,
		Value:   data,
	}
	err = wcs.proxy.Send(sendMsg)
	if err != nil {
		hwlog.RunLog.Errorf("websocket client send data failed: %v\n", err)
		return err
	}
	return nil
}
