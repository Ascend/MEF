// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package innerclient this file for initialize websocket client
package innerclient

import (
	"errors"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

var initFlag = false
var proxy *websocketmgr.WsClientProxy

// InitClient init client
func InitClient() error {
	certPath, err := path.GetCompSpecificDir(constants.InnerCertPathName)
	if err != nil {
		hwlog.RunLog.Errorf("get client cert path failed, error: %v", err)
		return errors.New("get client cert path failed")
	}
	kmcCfg, err := util.GetKmcConfig("")
	if err != nil {
		hwlog.RunLog.Errorf("get kmc config for client failed, error: %v", err)
		return errors.New("get kmc config for client failed")
	}
	_, err = x509.CheckCertsChainReturnContent(filepath.Join(certPath, constants.RootCaName))
	if err != nil {
		hwlog.RunLog.Errorf("check client cert failed, error: %v", err)
		return errors.New("check client cert failed")
	}
	certInfo := certutils.TlsCertInfo{
		RootCaPath: filepath.Join(certPath, constants.RootCaName),
		CertPath:   filepath.Join(certPath, constants.ClientCertName),
		KeyPath:    filepath.Join(certPath, constants.ClientKeyName),
		SvrFlag:    false,
		KmcCfg:     kmcCfg,
		WithBackup: true,
	}
	proxyConfig, err := websocketmgr.InitProxyConfig(constants.ClientIdName, constants.LocalIp,
		constants.InnerServerPort, certInfo, constants.EdgeOmSvcUrl)
	if err != nil {
		hwlog.RunLog.Errorf("init proxy config failed, error: %v", err)
		return errors.New("init proxy config failed")
	}

	proxyConfig.RegModInfos(getRegModuleInfoList())
	proxy = &websocketmgr.WsClientProxy{
		ProxyCfg: proxyConfig,
	}
	if err = proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("proxy.Start failed, error: %v", err)
		return errors.New("proxy.Start failed")
	}

	initFlag = true
	return nil
}

// GetCltSender get client sender
func GetCltSender() (*websocketmgr.WsClientProxy, error) {
	if !initFlag {
		if err := InitClient(); err != nil {
			hwlog.RunLog.Errorf("init inner-client failed before sending message to edge-main, error: %v", err)
			return nil, errors.New("init inner-client failed before sending message to edge-main")
		}
	}
	if proxy == nil {
		return nil, errors.New("init_client proxy is nil ")
	}

	return proxy, nil
}
