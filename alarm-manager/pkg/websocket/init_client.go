// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package websocket this file for initialize websocket client
package websocket

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"alarm-manager/pkg/alarmmanager"
	"alarm-manager/pkg/utils"

	"huawei.com/mindxedge/base/common"
)

var proxy *websocketmgr.WsClientProxy

const (
	msgRate   = 40
	burstSize = 100
)

func initClient() error {
	certInfo := certutils.TlsCertInfo{
		KmcCfg:     kmc.GetDefKmcCfg(),
		RootCaPath: utils.RootCaPath,
		CertPath:   utils.ServerCertPath,
		KeyPath:    utils.ServerKeyPath,
		SvrFlag:    false,
		WithBackup: true,
	}

	proxyConfig, err := websocketmgr.InitProxyConfig(common.AlarmManagerWsMoudle, common.EdgeMgrDns,
		common.EdgeManagerInnerWsPort, certInfo)
	if err != nil {
		hwlog.RunLog.Errorf("init proxy config failed, error: %v", err)
		return errors.New("init proxy config failed")
	}

	proxyConfig.RegModInfos(getRegModuleInfoList())
	if err := proxyConfig.SetRpsLimiterCfg(msgRate, burstSize); err != nil {
		hwlog.RunLog.Errorf("init websocket rps limiter config failed: %v", err)
		return fmt.Errorf("init websocket rps limiter config failed: %v", err)
	}
	proxy = &websocketmgr.WsClientProxy{
		ProxyCfg: proxyConfig,
	}
	proxy.SetDisconnCallback(clearAllAlarms)
	if err = proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("init alarm-manager client failed: %v", err)
		return errors.New("init alarm-manager client failed")
	}
	return nil
}

func clearAllAlarms(interface{}) {
	if err := alarmmanager.AlarmDbInstance().DeleteEdgeAlarm(); err != nil {
		hwlog.RunLog.Errorf("clear alarm info table failed: %s", err.Error())
		return
	}
	hwlog.RunLog.Info("edge-manager disconnected, clear all alarms from edge-manager and MEFEdge success")
}
