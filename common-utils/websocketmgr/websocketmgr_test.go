//  Copyright(C) 2024. Huawei Technologies Co.,Ltd.  All rights reserved.

package websocketmgr

import (
	"net/http"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
)

var (
	localHostIp      = "127.0.0.1"
	dtTestPort       = 8888
	serviceURL       = "/ut_test"
	msgOpt           = "GET"
	msgRes           = "DT_TEST"
	msgMod           = "DT_TEST"
	dtHandlerUrl     = "/dt_test"
	dtHandlerFunc    = func(writer http.ResponseWriter, request *http.Request) {}
	dtRps            = 100.0
	dtBurst          = 500
	dtMaxThroughput  = 1024
	dtPeriod         = time.Second
	dtMaxConnLimit   = 1500
	invalidWsMsgType = -1
	endpointName     = "UT_Test"
	serverName       = endpointName + "_Server"
	clientName       = endpointName + "_Client"
	serverProxy      *WsServerProxy
	clientProxy      *WsClientProxy
)

// TestCertStatus test case for InitProxyConfig
func TestServerProxy(t *testing.T) {
	convey.Convey("test server side proxy", t, testServerProxy)
	convey.Convey("test client side proxy", t, testClientProxy)
	convey.Convey("test sending message between server and client", t, testSendMsg)
	convey.Convey("test closing connection between server and client", t, testCloseProxy)
}

func testServerProxy() {
	tlsInfo := prepareTlsInfo(true)
	serverProxyConfig, err := InitProxyConfig(serverName, localHostIp, dtTestPort, tlsInfo)
	convey.So(err, convey.ShouldBeNil)

	// test ProxyConfig.RegModInfos
	testRegInfo := []modulemgr.MessageHandlerIntf{
		&modulemgr.RegisterModuleInfo{MsgOpt: msgOpt, MsgRes: msgRes, ModuleName: msgMod},
	}
	serverProxyConfig.RegModInfos(testRegInfo)

	// test ProxyConfig.SetRpsLimiterCfg and ProxyConfig.SetBandwidthLimiterCfg for server side
	convey.So(serverProxyConfig.SetRpsLimiterCfg(dtRps, dtBurst), convey.ShouldBeNil)
	convey.So(serverProxyConfig.SetBandwidthLimiterCfg(dtMaxThroughput, dtPeriod), convey.ShouldBeNil)

	// test WsServerProxy.GetName, WsServerProxy.AddDefaultHandler, WsServerProxy.AddHandler
	// and WsServerProxy.SetConnLimiter
	serverProxy = &WsServerProxy{
		ProxyCfg: serverProxyConfig,
	}
	convey.So(serverProxy.GetName(), convey.ShouldEqual, serverName)
	serverProxy.AddDefaultHandler()
	convey.So(serverProxy.AddHandler(dtHandlerUrl, dtHandlerFunc), convey.ShouldBeNil)
	convey.So(serverProxy.SetConnLimiter(dtMaxConnLimit), convey.ShouldBeNil)

	// test WsServerProxy callbacks setter
	serverProxy.SetOnConnCallback(nil)
	serverProxy.SetOnConnCallback(func(info WebsocketPeerInfo) {})
	serverProxy.SetDisconnCallback(nil)
	serverProxy.SetDisconnCallback(func(info WebsocketPeerInfo) {})

	// test WsServerProxy.Start
	convey.So(serverProxy.Start(), convey.ShouldBeNil)
	// wait for server is ready
	time.Sleep(time.Second)

	// test WsServerProxy callbacks getter
	convey.So(len(serverProxy.GetOnConnectCallbacks()), convey.ShouldEqual, 1)
	convey.So(len(serverProxy.GetDisconnectCallbacks()), convey.ShouldEqual, 1)
	convey.So(len(serverProxy.GetReconnectCallbacks()), convey.ShouldEqual, 0)

	// test WsServerProxy.GetBandwidthLimiter, GetProxyConfig.GetProxyConfig
	convey.So(serverProxy.GetBandwidthLimiter(), convey.ShouldNotBeNil)
	convey.So(serverProxy.GetProxyConfig(), convey.ShouldNotBeNil)
}

func testClientProxy() {
	clientTlsInfo := prepareTlsInfo(false)
	clientProxyConfig, err := InitProxyConfig(clientName, localHostIp, dtTestPort, clientTlsInfo)
	convey.So(err, convey.ShouldBeNil)

	// test ProxyConfig.SetRpsLimiterCfg and ProxyConfig.SetBandwidthLimiterCfg for client side
	convey.So(clientProxyConfig.SetRpsLimiterCfg(dtRps, dtBurst), convey.ShouldBeNil)
	convey.So(clientProxyConfig.SetBandwidthLimiterCfg(dtMaxThroughput, dtPeriod), convey.ShouldBeNil)

	clientHeader := http.Header{}
	clientHeader.Add("X-Real-IP", localHostIp)
	clientHeader.Add("clientName", clientName)
	clientProxyConfig.headers = clientHeader
	clientProxy = &WsClientProxy{ProxyCfg: clientProxyConfig}

	// test WsClientProxy callbacks setter
	clientProxy.SetDisconnCallback(nil)
	clientProxy.SetDisconnCallback(func(info WebsocketPeerInfo) {})
	clientProxy.SetReConnCallback(nil)
	clientProxy.SetReConnCallback(func() {})

	// test WsClientProxy.Start, ws client will be connected with ws server
	convey.So(clientProxy.Start(), convey.ShouldBeNil)
	// wait for connection is ready
	time.Sleep(time.Second)

	// test WsClientProxy.GetAllPeers and GetPeer, check connection counter
	if serverProxy == nil {
		panic("serverProxy is not initialized")
	}

	allConnectedClients, err := serverProxy.GetAllPeers()
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(allConnectedClients), convey.ShouldEqual, 1)
	convey.So(serverProxy, convey.ShouldNotBeNil)
	clientInfo, err := serverProxy.GetPeer(clientName)
	convey.So(err, convey.ShouldBeNil)
	convey.So(clientInfo, convey.ShouldNotBeNil)

	// test WsClientProxy callbacks getter
	convey.So(len(clientProxy.GetReconnectCallbacks()), convey.ShouldEqual, 1)
	convey.So(len(clientProxy.GetDisconnectCallbacks()), convey.ShouldEqual, 1)
	convey.So(len(clientProxy.GetOnConnectCallbacks()), convey.ShouldEqual, 0)

	// test WsClientProxy GetName, IsConnected, GetBandwidthLimiter, GetProxyConfig
	convey.So(clientProxy.GetName(), convey.ShouldEqual, clientName)
	convey.So(clientProxy.IsConnected(), convey.ShouldEqual, true)
	convey.So(clientProxy.GetBandwidthLimiter(), convey.ShouldNotBeNil)
	convey.So(clientProxy.GetProxyConfig(), convey.ShouldNotBeNil)
}

func testSendMsg() {
	if serverProxy == nil {
		panic("clientProxy is not initialized")
	}
	if clientProxy == nil {
		panic("clientProxy is not initialized")
	}
	// replace modulemgr.HandleMessageIntf with fake message handler for server side
	patchForServer := gomonkey.ApplyMethodFunc(
		serverProxy.GetProxyConfig().handlerMgr, "HandleMsg", fakeHandleWebSocketMsg)
	defer patchForServer.Reset()
	// replace modulemgr.HandleMessageIntf with fake message handler for client side
	patchForClient := gomonkey.ApplyMethodFunc(
		clientProxy.GetProxyConfig().handlerMgr, "HandleMsg", fakeHandleWebSocketMsg)
	defer patchForClient.Reset()

	// test WsClientProxy.Send and WsServerProxy.Send
	msg, err := model.NewMessage()
	convey.So(err, convey.ShouldBeNil)
	msg.SetRouter("DT_TEST", msgMod, msgOpt, msgRes)
	convey.So(msg.FillContent("DT test message from client"), convey.ShouldBeNil)
	convey.So(clientProxy.Send(msg), convey.ShouldBeNil)
	convey.So(clientProxy.Send(msg, invalidWsMsgType), convey.ShouldNotBeNil)

	msg, err = model.NewMessage()
	convey.So(err, convey.ShouldBeNil)
	msg.SetRouter("DT_TEST", msgMod, msgOpt, msgRes)
	convey.So(msg.FillContent("DT test message from server"), convey.ShouldBeNil)
	convey.So(serverProxy.Send(clientName, msg), convey.ShouldBeNil)
	convey.So(serverProxy.Send(clientName, msg, invalidWsMsgType), convey.ShouldNotBeNil)
}

func testCloseProxy() {
	if serverProxy == nil {
		panic("serverProxy is not initialized")
	}
	if clientProxy == nil {
		panic("clientProxy is not initialized")
	}
	// test WsClientProxy.Stop and WsServerProxy.Stop
	convey.So(clientProxy.Stop(), convey.ShouldBeNil)
	// wait for client side closes it's connection
	time.Sleep(time.Second)
	convey.So(serverProxy.Stop(), convey.ShouldBeNil)
}

func fakeHandleWebSocketMsg(_ []byte, _ model.MsgPeerInfo) []byte {
	hwlog.RunLog.Info("DT TEST fake message handler")
	return nil
}
