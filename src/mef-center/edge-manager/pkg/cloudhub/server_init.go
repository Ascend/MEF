// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cloudhub server init
package cloudhub

import (
	"errors"
	"fmt"
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
	name             = "server_edge_ctl"
	largeFileTimeout = 10 * time.Minute

	wsMaxThroughput    = 100 * common.MB
	wsThroughputPeriod = 30 * time.Second
)

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
	proxyConfig.SetTimeout(largeFileTimeout, largeFileTimeout, 0)
	if err := proxyConfig.SetBandwidthLimiterCfg(wsMaxThroughput, wsThroughputPeriod); err != nil {
		hwlog.RunLog.Errorf("init bandwidth limiter config failed: %v", err)
		return nil, fmt.Errorf("init bandwidth limiter config failed: %v", err)
	}
	proxy := &websocketmgr.WsServerProxy{
		ProxyCfg: proxyConfig,
	}
	if server.maxClientNum > 0 {
		if err := proxy.SetConnLimiter(server.maxClientNum); err != nil {
			hwlog.RunLog.Errorf("init connection limiter failed: %v", err)
			return nil, fmt.Errorf("init connection limiter failed: %v", err)
		}
	}
	proxy.AddDefaultHandler()
	proxy.SetDisconnCallback(clearAlarm)
	proxy.SetOnConnCallback(syncCertsToEdgeNode)
	if err = proxy.AddHandler(constants.LogUploadUrl, logmanager.HandleUpload); err != nil {
		hwlog.RunLog.Error("add handler failed")
		return nil, errors.New("add handler failed")
	}
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

func clearAlarm(peerInfo websocketmgr.WebsocketPeerInfo) {
	sn := peerInfo.Sn

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

func syncCertsToEdgeNode(peerInfo websocketmgr.WebsocketPeerInfo) {
	hwlog.RunLog.Infof("start to send certs to edge node[name=%s][ip=%s]", peerInfo.Sn, peerInfo.Ip)
	certs := []string{common.SoftwareCertName, common.ImageCertName}
	for _, cert := range certs {
		if err := sendCertMsg(cert, peerInfo); err != nil {
			hwlog.RunLog.Warnf("sync [%s] cert to edge node[name=%s][ip=%s] failed, %v", cert, peerInfo.Sn, peerInfo.Ip, err)
			continue
		}
	}
}

func sendCertMsg(cert string, peerInfo websocketmgr.WebsocketPeerInfo) error {
	msg, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("create message failed, %v", err)
	}
	msg.SetNodeId(peerInfo.Sn)
	msg.SetRouter(common.CloudHubName, common.NodeMsgManagerName, common.OptGet, common.ResDownLoadCert)

	if err := msg.FillContent(cert); err != nil {
		return fmt.Errorf("fill message content failed, %v", err)
	}
	if err := modulemgr.SendMessage(msg); err != nil {
		return fmt.Errorf("send message failed, %v", err)
	}
	return nil
}
