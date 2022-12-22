// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// WsSvrSender websocket server sender
type WsSvrSender struct {
	proxy NetProxyIntf
}

// SetProxy set sender proxy
func (wss *WsSvrSender) SetProxy(proxy NetProxyIntf) {
	wss.proxy = proxy
}

// Send sends message
func (wss *WsSvrSender) Send(clientId string, msg *model.Message) error {
	data, err := json.Marshal(msg)
	cltMsg := wsSvrMessage{
		Msg: &wsMessage{
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
