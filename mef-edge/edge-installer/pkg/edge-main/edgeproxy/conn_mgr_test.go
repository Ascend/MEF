// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package edgeproxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

func TestWebSocket(t *testing.T) {
	conn, server := CreateWebsocket(getAsyncMessage)
	defer func() {
		server.Close()
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}()
	connName := "testConn"
	convey.Convey("Creating a WebSocket and registry it and try get it", t, func() {
		err := RegistryConn(connName, conn)
		convey.So(err, convey.ShouldBeNil)
		_, err = GetConnByModName(connName)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("Sending Message to WebSocket", t, func() {
		msg := CreateMessage()
		err := SendMsgToWs(msg, connName)
		convey.So(err, convey.ShouldResemble, nil)
	})
	convey.Convey("Sending PingMessage to WebSocket and deregister it", t, func() {
		err := SendHeartbeatToPeer(websocket.PingMessage, "ping date", connName)
		convey.So(err, convey.ShouldResemble, nil)
		err = UnRegistryConn(connName)
		convey.So(err, convey.ShouldBeNil)
	})

}

func CreateWebsocket(getMessage func() *model.Message) (*websocket.Conn, *httptest.Server) {
	// 创建一个虚拟的 WebSocket 服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 升级连接为WebSocket连接
		upgrade := &websocket.Upgrader{}
		conn, err := upgrade.Upgrade(w, r, http.Header{})
		if err != nil {
			panic(err)
		}
		// 构造消息
		msg := getMessage()

		// 将Message对象编码为JSON格式
		jsonData, err := json.Marshal(msg)
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
			return
		}
		// 发送JSON数据
		err = conn.WriteMessage(websocket.TextMessage, jsonData)
		// 接收一条消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("When receiving a message,An error occurred: %s\n", err)
			return
		}
		// 检查消息类型是否正确
		if messageType != websocket.TextMessage {
			fmt.Printf("When checking the parameter type ,An error occurred: %s\n", err)
			return
		}
		// 检查接收到的消息是否正确
		var receivedMessage model.Message
		err = json.Unmarshal(message, &receivedMessage)
		if err != nil {
			fmt.Printf("When Unmarshal, An error occurred: %s\n", err)
			return
		}
		fmt.Printf("receivedMessage: %s\n", message)
	}))

	// 创建一个WebSocket连接
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}

	return conn, server
}

func CreateMessage() *model.Message {
	message, err := model.NewMessage()
	if err != nil {
		panic(err)
	}
	if err = message.FillContent("content"); err != nil {
		panic(err)
	}
	message.SetKubeEdgeRouter(
		constants.ControllerModule,
		constants.ResourceModule,
		constants.OptUpdate,
		"websocket/pod/abc",
	)
	return message
}

func initLog() {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := util.InitHwLogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
	}
}

func getAsyncMessage() *model.Message {

	cntBytes := []byte(`Hi!`)

	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, %v", err)
		return nil
	}

	respMsg.KubeEdgeRouter = model.MessageRoute{
		Source:    constants.SourceHardware,
		Group:     constants.GroupHub,
		Operation: constants.OptUpdate,
		Resource:  constants.ActionSecret,
	}
	respMsg.Header.ID = respMsg.Header.Id
	respMsg.Header.Sync = false
	respMsg.SetRouter(constants.CfgRestore, constants.ModDeviceOm, constants.OptUpdate, constants.ActionSecret)
	if err = respMsg.FillContent(cntBytes); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return nil
	}
	return respMsg
}
