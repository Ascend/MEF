// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package innerwebsocket

import (
	"context"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"edge-manager/pkg/constants"

	"huawei.com/mindxedge/base/common"
)

const (
	name = "inner_ws_server"
)

var serverSender websocketmgr.WsSvrSender

const (
	msgRate   = 40
	burstSize = 100
)

// WsInnerServer wraps the struct WebSocketServer
type WsInnerServer struct {
	WsPort int
	Ctx    context.Context
}

// InnerWsServer is the server of the inner websocket
var InnerWsServer WsInnerServer

// InitInnerWsServer init server
func InitInnerWsServer(innerWsPort int) error {
	InnerWsServer = WsInnerServer{
		WsPort: innerWsPort,
		Ctx:    context.Background(),
	}
	certInfo := certutils.TlsCertInfo{
		KmcCfg:     kmc.GetDefKmcCfg(),
		RootCaPath: constants.RootCaPath,
		CertPath:   constants.ServerCertPath,
		KeyPath:    constants.ServerKeyPath,
		SvrFlag:    true,
		WithBackup: true,
	}

	podIp, err := common.GetPodIP()
	if err != nil {
		hwlog.RunLog.Errorf("get edge manager pod ip failed: %s", err.Error())
		return errors.New("get edge manager pod ip")
	}
	proxyConfig, err := websocketmgr.InitProxyConfig(name, podIp, InnerWsServer.WsPort, certInfo)
	if err != nil {
		hwlog.RunLog.Errorf("init proxy config failed: %v", err)
		return errors.New("init proxy config failed")
	}
	proxyConfig.RegModInfos(getRegModuleInfoList())
	proxy := &websocketmgr.WsServerProxy{
		ProxyCfg: proxyConfig,
	}

	proxy.AddDefaultHandler()
	serverSender.SetProxy(proxy)
	if err = proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("proxy start failed: %v", err)
		return errors.New("proxy start failed")
	}

	if err = proxy.SetLimiter(websocketmgr.NewMsgLimiter(msgRate, burstSize)); err != nil {
		hwlog.RunLog.Errorf("set msg Limiter failed: %s", err.Error())
		return errors.New("set msg limiter failed")
	}

	hwlog.RunLog.Info("cloudhub server start success")
	return nil
}

// getWsSender returns a websocket server sender
func getWsSender() websocketmgr.WsSvrSender {
	return serverSender
}

// SendMessageByInnerWs send a message to specified module
// message is the message need tobe sent, should pre set router and content
// moduleName is the name of the ws client
func sendMessageByInnerWs(message *model.Message, moduleName string) error {
	if message == nil {
		return fmt.Errorf("failed to send message to %s while message is nil", moduleName)
	}
	message.SetNodeId(moduleName)
	sender := getWsSender()
	return sender.Send(message.GetNodeId(), message)
}
