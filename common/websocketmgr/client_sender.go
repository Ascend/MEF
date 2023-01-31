// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package websocketmgr

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/modulemanager/model"
)

// WsCltSender websocket client sender
type WsCltSender struct {
	proxy NetProxyIntf
}

// SetProxy websocket client sender set proxy
func (wcs *WsCltSender) SetProxy(proxy NetProxyIntf) {
	wcs.proxy = proxy
}

// Send websocket sender send message
func (wcs *WsCltSender) Send(msg *model.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("wesocket client send failed, json marshal message error: %v", err)
	}
	sendMsg := wsMessage{
		MsgType: websocket.TextMessage,
		Value:   data,
	}
	err = wcs.proxy.Send(sendMsg)
	if err != nil {
		hwlog.RunLog.Errorf("websocket client send data failed: %v", err)
		return err
	}
	return nil
}
