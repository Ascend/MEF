// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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

const (
	msgRate   = 40
	burstSize = 100
)

// WsInnerServer wraps the struct WebSocketServer
type WsInnerServer struct {
	WsPort int
	Ctx    context.Context
	Proxy  *websocketmgr.WsServerProxy
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
	if err = proxyConfig.SetRpsLimiterCfg(msgRate, burstSize); err != nil {
		hwlog.RunLog.Errorf("set rps limiter config failed: %s", err.Error())
		return fmt.Errorf("set rps limiter config failed: %s", err.Error())
	}
	proxy := &websocketmgr.WsServerProxy{
		ProxyCfg: proxyConfig,
	}
	proxy.AddDefaultHandler()
	if err = proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("proxy start failed: %v", err)
		return errors.New("proxy start failed")
	}
	InnerWsServer.Proxy = proxy
	hwlog.RunLog.Info("cloudhub server start success")
	return nil
}

// SendMessageByInnerWs send a message to specified module
// message is the message need tobe sent, should pre set router and content
// moduleName is the name of the ws client
func sendMessageByInnerWs(message *model.Message, moduleName string) error {
	if message == nil {
		return fmt.Errorf("failed to send message to %s while message is nil", moduleName)
	}
	message.SetNodeId(moduleName)
	if InnerWsServer.Proxy == nil {
		return fmt.Errorf("inner ws proxy is not initialized")
	}
	return InnerWsServer.Proxy.Send(message.GetNodeId(), message)
}
