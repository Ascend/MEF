// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager"
)

const (
	name           = "server_edge_ctl"
	wsWriteTimeout = 10 * time.Minute
)

var serverSender websocketmgr.WsSvrSender
var initFlag bool

// InitServer init server
func InitServer() error {
	certInfo := certutils.TlsCertInfo{
		KmcCfg:     kmc.GetDefKmcCfg(),
		RootCaPath: constants.RootCaPath,
		CertPath:   constants.ServerCertPath,
		KeyPath:    constants.ServerKeyPath,
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
	proxyConfig.ReadTimeout = wsWriteTimeout
	proxyConfig.WriteTimeout = wsWriteTimeout
	proxy := &websocketmgr.WsServerProxy{
		ProxyCfg: proxyConfig,
	}

	proxy.AddDefaultHandler()
	if err = proxy.AddHandler(constants.LogUploadUrl, logmanager.HandleUpload); err != nil {
		hwlog.RunLog.Error("add handler failed")
		return errors.New("add handler failed")
	}
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
		RootCaPath:    constants.RootCaPath,
		CertPath:      constants.ServerCertPath,
		KeyPath:       constants.ServerKeyPath,
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
