// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"encoding/json"
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
	}
	go authServer()

	podIp, err := common.GetPodIP()
	if err != nil {
		hwlog.RunLog.Errorf("get edge manager pod ip failed: %s", err.Error())
		return nil, errors.New("get edge manager pod ip")
	}
	proxyConfig, err := websocketmgr.InitProxyConfig(name, podIp, server.wsPort, certInfo)
	if err != nil {
		hwlog.RunLog.Errorf("init proxy config failed: %v", err)
		return nil, errors.New("init proxy config failed")
	}
	proxyConfig.RegModInfos(getRegModuleInfoList())
	proxyConfig.ReadTimeout = wsWriteTimeout
	proxyConfig.WriteTimeout = wsWriteTimeout
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
	}
	NewClientAuthService(server.authPort, authCertInfo).Start()
}

// GetSvrSender get server sender
func GetSvrSender() (websocketmgr.WsSvrSender, error) {
	if !initFlag {
		if _, err := InitServer(); err != nil {
			hwlog.RunLog.Errorf("init websocket server failed before sending message to mef-edge, error: %v", err)
			return websocketmgr.WsSvrSender{}, errors.New("init websocket server failed before sending message to mef-edge")
		}
	}
	return serverSender, nil
}

func clearAlarm(sn string) {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create new msg failed: %s", err.Error())
		return
	}

	clearStruct := requests.ClearNodeAlarmReq{
		Sn: sn,
	}
	clearMsg, err := json.Marshal(clearStruct)
	if err != nil {
		hwlog.RunLog.Errorf("marshal msg to alarm manager failed: %s", err.Error())
		return
	}
	msg.FillContent(string(clearMsg))
	msg.SetNodeId(common.AlarmManagerClientName)
	msg.SetRouter(common.CloudHubName, common.InnerServerName, common.Delete, requests.ClearOneNodeAlarmRouter)

	const (
		clearWaitTime = 60 * time.Second
		maxRetryTimes = 20
	)

	for i := 0; i < maxRetryTimes; i++ {
		ret, err := modulemgr.SendSyncMessage(msg, common.ResponseTimeout)
		if err != nil {
			hwlog.RunLog.Errorf("send clear node alarm msg to alarm-manager failed: %s", err.Error())
			return
		}

		if content, ok := ret.GetContent().(string); !ok || content != common.OK {
			hwlog.RunLog.Warnf("clear node: %s alarm failed", sn)
			time.Sleep(clearWaitTime)
			continue
		}
		break
	}
}
