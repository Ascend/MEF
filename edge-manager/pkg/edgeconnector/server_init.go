// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector server init
package edgeconnector

import (
	"errors"
	"os"
	"path"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/websocketmgr"
)

const (
	name                      = "server_edge_ctl"
	rootCaNameValidEdge       = "valid_edge_root.crt"
	rootCaBackUpNameValidEdge = "valid_edge_root_backup.crt"
	serviceCaName             = "server.crt"
	keyFileName               = "server.key"
	pwdFileName               = "server.pwd"
)

var serverSender websocketmgr.WsSvrSender
var initFlag bool

// InitServer init server
func InitServer() error {
	certPath, ok := common.GetCertDir()
	if !ok {
		hwlog.RunLog.Error("get server cert path failed")
		return errors.New("get server cert path failed")
	}

	certInfo := websocketmgr.CertPathInfo{
		RootCaPath:  path.Join(certPath, rootCaNameValidEdge),
		SvrCertPath: path.Join(certPath, serviceCaName),
		SvrKeyPath:  path.Join(certPath, keyFileName),
		ServerFlag:  true,
	}

	podIp := os.Getenv("POD_IP")
	if podIp == "" {
		hwlog.RunLog.Error("get edge-manager pod ip failed")
		return errors.New("get edge-manager pod ip failed")
	}

	proxyConfig, err := websocketmgr.InitProxyConfig(name, podIp, connector.wsPort, certInfo)
	if err != nil {
		hwlog.RunLog.Errorf("init proxy config failed: %v", err)
		return err
	}
	proxyConfig.RegModInfos(getRegModuleInfoList())
	proxy := &websocketmgr.WsServerProxy{
		ProxyCfg: proxyConfig,
	}
	serverSender.SetProxy(proxy)
	if err = proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("proxy.Start failed: %v", err)
		return err
	}

	initFlag = true
	return nil
}

// GetSvrSender get server sender
func GetSvrSender() websocketmgr.WsSvrSender {
	if !initFlag {
		if err := InitServer(); err != nil {
			return websocketmgr.WsSvrSender{}
		}
	}
	return serverSender
}
