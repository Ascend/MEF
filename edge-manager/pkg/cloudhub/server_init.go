// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
)

const (
	name = "server_edge_ctl"
)

var serverSender websocketmgr.WsSvrSender
var initFlag bool

// InitServer init server
func InitServer() error {
	certInfo := certutils.TlsCertInfo{
		KmcCfg:     kmc.GetDefKmcCfg(),
		RootCaPath: util.RootCaPath,
		CertPath:   util.ServerCertPath,
		KeyPath:    util.ServerKeyPath,
	}
	go authServer()

	podIp, err := common.GetPodIP()
	if err != nil {
		hwlog.RunLog.Errorf("get edge manager pod ip failed: %s", err.Error())
		return errors.New("get edge manager pod ip")
	}
	proxyConfig, err := websocketmgr.InitProxyConfig(name, podIp, server.wsPort, certInfo)
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
		hwlog.RunLog.Errorf("proxy.Start failed: %v", err)
		return errors.New("proxy.Start failed")
	}
	hwlog.RunLog.Infof("cloudhub server start success")
	initFlag = true
	return nil
}

func authServer() {
	authCertInfo := certutils.TlsCertInfo{
		KmcCfg:        kmc.GetDefKmcCfg(),
		RootCaPath:    util.RootCaPath,
		CertPath:      util.ServerCertPath,
		KeyPath:       util.ServerKeyPath,
		SvrFlag:       true,
		IgnoreCltCert: true,
	}
	NewClientAuthService(server.authPort, authCertInfo).Start()
}

// GetSvrSender get server sender
func GetSvrSender() (websocketmgr.WsSvrSender, error) {
	if !initFlag {
		if err := InitServer(); err != nil {
			hwlog.RunLog.Errorf("init websocket server failed before sending message to mef-edge, error: %v", err)
			return websocketmgr.WsSvrSender{}, errors.New("init websocket server failed before sending message to mef-edge")
		}
	}
	return serverSender, nil
}
