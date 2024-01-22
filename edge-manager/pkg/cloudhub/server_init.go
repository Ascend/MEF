// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	name           = "server_edge_ctl"
	wsWriteTimeout = 10 * time.Minute

	wsMaxThroughput    = 100 * common.MB
	wsThroughputPeriod = 30 * time.Second
)

var serverSender websocketmgr.WsSvrSender
var initFlag bool

// InitServer init server
func InitServer() (*websocketmgr.WsServerProxy, error) {
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
		return nil, errors.New("get edge manager pod ip")
	}
	server.serverIp = podIp
	go authServer()

	proxyConfig, err := websocketmgr.InitProxyConfig(name, server.serverIp, server.wsPort, certInfo)
	if err != nil {
		hwlog.RunLog.Errorf("init proxy config failed: %v", err)
		return nil, errors.New("init proxy config failed")
	}
	proxyConfig.RegModInfos(getRegModuleInfoList())
	proxyConfig.ReadTimeout = wsWriteTimeout
	proxyConfig.WriteTimeout = wsWriteTimeout
	proxyConfig.MaxThroughputPerPeriod = wsMaxThroughput
	proxyConfig.ThroughputPeriod = wsThroughputPeriod
	proxy := &websocketmgr.WsServerProxy{
		ProxyCfg: proxyConfig,
	}

	proxy.AddDefaultHandler()
	proxy.SetDisconnCallback(clearAlarm)
	if err = proxy.AddHandler(constants.LogUploadUrl, logmanager.HandleUpload); err != nil {
		hwlog.RunLog.Error("add handler failed")
		return nil, errors.New("add handler failed")
	}
	serverSender.SetProxy(proxy)
	if err = proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("proxy.Start failed: %v", err)
		return nil, errors.New("proxy.Start failed")
	}
	hwlog.RunLog.Info("cloudhub server start success")
	initFlag = true
	return proxy, nil
}

func authServer() {
	authCertInfo := certutils.TlsCertInfo{
		KmcCfg:        kmc.GetDefKmcCfg(),
		RootCaPath:    constants.RootCaPath,
		CertPath:      constants.ServerCertPath,
		KeyPath:       constants.ServerKeyPath,
		SvrFlag:       true,
		IgnoreCltCert: true,
		WithBackup:    true,
	}

	for count := 0; count < constants.ServerInitRetryCount; count++ {
		NewClientAuthService(server.serverIp, server.authPort, authCertInfo).Start()
		hwlog.RunLog.Error("start auth server failed. Restart server later")
		time.Sleep(constants.ServerInitRetryInterval)
	}
	hwlog.RunLog.Error("start auth server failed after maximum number of retry")
}

// GetSvrSender get server sender
func GetSvrSender() (websocketmgr.WsSvrSender, error) {
	if !initFlag {
		if _, err := InitServer(); err != nil {
			hwlog.RunLog.Errorf("init websocket server failed before sending message to edge, error: %v", err)
			return websocketmgr.WsSvrSender{}, errors.New("init websocket server failed before sending message to edge")
		}
	}
	return serverSender, nil
}

func clearAlarm(arg interface{}) {
	sn, ok := arg.(string)
	if !ok {
		hwlog.RunLog.Error("clear alarm failed: sn is invalid type")
		return
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create new msg failed: %s", err.Error())
		return
	}

	clearStruct := requests.ClearNodeAlarmReq{
		Sn: sn,
	}
	if err = msg.FillContent(clearStruct, true); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return
	}
	msg.SetNodeId(common.AlarmManagerWsMoudle)
	msg.SetRouter(common.CloudHubName, common.CloudHubName, common.Delete, requests.ClearOneNodeAlarmRouter)

	const (
		clearWaitTime = 30 * time.Second
		maxRetryTimes = 10
	)

	for i := 0; i < maxRetryTimes; i++ {
		ret, err := modulemgr.SendSyncMessage(msg, common.ResponseTimeout)
		if err != nil {
			hwlog.RunLog.Errorf("send clear node alarm msg to alarm-manager failed: %s", err.Error())
			continue
		}

		var content string
		if err = ret.ParseContent(&content); err != nil {
			hwlog.RunLog.Errorf("parse content failed: %v", err)
			time.Sleep(clearWaitTime)
			continue
		}
		if content != common.OK {
			hwlog.RunLog.Warnf("clear node: %s alarm failed", sn)
			time.Sleep(clearWaitTime)
			continue
		}
		hwlog.RunLog.Infof("clear alarm for node %s success", sn)
		break
	}
}
