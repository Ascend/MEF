// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"errors"
	"math"
	"path"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/common/websocketmgr"

	"edge-manager/pkg/util"
)

const (
	name              = "server_edge_ctl"
	certPathDir       = "/home/data/config/websocket-certs"
	rootNameValidEdge = "root_edge.crt"
	serviceName       = "server.crt"
	keyFileName       = "server.key"
	retryTime         = 30
	maxRetry          = math.MaxInt
	waitTime          = 5 * time.Second
)

var serverSender websocketmgr.WsSvrSender
var initFlag bool

// InitServer init server
func InitServer() error {
	checkAndSetWsSvcCert()
	rootCaBytes, err := getWsRootCert()
	if err != nil {
		return err
	}
	certInfo := certutils.TlsCertInfo{
		KmcCfg:        common.GetDefKmcCfg(),
		RootCaContent: rootCaBytes,
		CertPath:      path.Join(certPathDir, serviceName),
		KeyPath:       path.Join(certPathDir, keyFileName),
		SvrFlag:       true,
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
	serverSender.SetProxy(proxy)
	if err = proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("proxy.Start failed: %v", err)
		return errors.New("proxy.Start failed")
	}

	initFlag = true
	return nil
}

func checkAndSetWsSvcCert() {
	keyPath := path.Join(certPathDir, keyFileName)
	certPath := path.Join(certPathDir, serviceName)
	if utils.IsExist(keyPath) && utils.IsExist(certPath) {
		hwlog.RunLog.Info("check websocket server certs success")
		return
	}
	hwlog.RunLog.Warn("check websocket server certs failed, start to create")
	svcCertStr, err := getWsSvcCert(keyPath)
	if err != nil {
		return
	}
	err = common.WriteData(certPath, []byte(svcCertStr))
	if err != nil {
		hwlog.RunLog.Errorf("save cert for websocket service cert failed: %v", err)
		return
	}

	hwlog.RunLog.Info("create cert for websocket service success")
}

func getWsSvcCert(keyPath string) (string, error) {
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: util.RootCaPath,
			CertPath:   util.ServerCertPath,
			KeyPath:    util.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	var svcCertStr string
	var err error
	san := certutils.CertSan{DnsName: []string{common.EdgeMgrDns}}
	ips, err := common.GetHostIpV4()
	if err != nil {
		return "", err
	}
	san.IpAddr = ips
	csr, err := certutils.CreateCsr(keyPath, common.WsSerName, nil, san)
	if err != nil {
		hwlog.RunLog.Errorf("create websocket service cert csr failed: %v", err)
		return "", err
	}
	for i := 0; i < maxRetry; i++ {
		svcCertStr, err = reqCertParams.ReqIssueSvrCert(common.WsSerName, csr)
		if err == nil {
			break
		}
		time.Sleep(waitTime)
	}
	if svcCertStr == "" {
		hwlog.RunLog.Errorf("issue svcCertStr for websocket service cert failed: %v", err)
		return "", err
	}
	return svcCertStr, nil
}

func getWsRootCert() ([]byte, error) {
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath: util.RootCaPath,
			CertPath:   util.ServerCertPath,
			KeyPath:    util.ServerKeyPath,
			SvrFlag:    false,
		},
	}
	var rootCaStr string
	var err error
	for i := 0; i < maxRetry; i++ {
		rootCaStr, err = reqCertParams.GetRootCa(common.WsCltName)
		if err == nil {
			break
		}
		time.Sleep(waitTime)
	}
	if rootCaStr == "" {
		hwlog.RunLog.Errorf("get valid root ca for websocket service failed: %v", err)
		return nil, err
	}

	return []byte(rootCaStr), nil
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
