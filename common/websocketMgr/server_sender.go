package websocket

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type WsSvrSender struct {
	proxy NetProxyIntf
}

func (wss *WsSvrSender) SetProxy(proxy NetProxyIntf) {
	wss.proxy = proxy
}

func (wss *WsSvrSender) Send(clientId string, msg *model.Message) error {
	data, err := json.Marshal(msg)
	cltMsg := WsSvrMessage{
		Msg: &WsMessage{
			MsgType: websocket.TextMessage,
			Value:   data,
		},
		ClientName: clientId,
	}
	err = wss.proxy.Send(cltMsg)
	if err != nil {
		return err
	}
	return nil
}
