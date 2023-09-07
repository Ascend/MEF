// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package innerserver

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"edge-manager/pkg/constants"

	"huawei.com/mindxedge/base/common"
)

const (
	name = "inner_server"
)

var serverSender websocketmgr.WsSvrSender

const (
	msgRate   = 40
	burstSize = 100
)

// InitServer init server
func InitServer() error {
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

func getWsSender() websocketmgr.WsSvrSender {
	return serverSender
}
