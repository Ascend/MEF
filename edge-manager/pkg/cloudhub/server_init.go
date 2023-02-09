// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"errors"
	"os"
	"path"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/common/httpsmgr"
	"huawei.com/mindxedge/base/common/websocketmgr"
)

const (
	name              = "server_edge_ctl"
	certPathDir       = "/home/data/config/websocket-certs"
	rootNameValidEdge = "root_edge.crt"
	serviceName       = "server.crt"
	keyFileName       = "server.key"
	retryTime         = 30
	waitTime          = 5 * time.Second
)

var serverSender websocketmgr.WsSvrSender
var initFlag bool

// InitServer init server
func InitServer() error {
	checkAndGetWsCert()
	certInfo := certutils.TlsCertInfo{
		KmcCfg:        common.GetDefKmcCfg(),
		RootCaPath:    path.Join(certPathDir, rootNameValidEdge),
		CertPath:      path.Join(certPathDir, serviceName),
		KeyPath:       path.Join(certPathDir, keyFileName),
		SvrFlag:       true,
		IgnoreCltCert: true,
	}

	podIp := os.Getenv("POD_IP")
	if podIp == "" {
		hwlog.RunLog.Error("get edge-manager pod ip failed")
		return errors.New("get edge-manager pod ip failed")
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

func checkAndGetWsCert() {
	keyPath := path.Join(certPathDir, keyFileName)
	certPath := path.Join(certPathDir, serviceName)
	rootCaPath := path.Join(certPathDir, rootNameValidEdge)
	if utils.IsExist(keyPath) && utils.IsExist(certPath) && utils.IsExist(rootCaPath) {
		hwlog.RunLog.Info("check websocket server certs success")
		return
	}
	hwlog.RunLog.Warn("check websocket server certs failed, start to create")
	certStr, rootCaStr, err := getWsCert(keyPath)
	if err != nil {
		return
	}
	err = common.WriteData(certPath, []byte(certStr))
	if err != nil {
		hwlog.RunLog.Errorf("save cert for websocket service cert failed: %v", err)
		return
	}

	if err = common.WriteData(rootCaPath, []byte(rootCaStr)); err != nil {
		hwlog.RunLog.Errorf("save cert for websocket service cert failed: %v", err)
		return
	}
	hwlog.RunLog.Info("create cert for websocket service success")
}

func getWsCert(keyPath string) (string, string, error) {
	san := certutils.CertSan{DnsName: []string{common.EdgeMgrDns}}
	ips, err := common.GetHostIpV4()
	if err != nil {
		return "", "", err
	}
	san.IpAddr = ips
	csr, err := certutils.CreateCsr(keyPath, common.WsSerName, nil, san)
	if err != nil {
		hwlog.RunLog.Errorf("create websocket service cert csr failed: %v", err)
		return "", "", err
	}
	reqCertParams := httpsmgr.ReqCertParams{
		ClientTlsCert: certutils.TlsCertInfo{
			RootCaPath:    util.RootCaPath,
			CertPath:      util.ServerCertPath,
			KeyPath:       util.ServerKeyPath,
			SvrFlag:       false,
			IgnoreCltCert: false,
		},
	}
	var certStr, rootCaStr string
	for i := 0; i < retryTime; i++ {
		rootCaStr, err = reqCertParams.GetRootCa(common.WsCltName)
		if err == nil {
			break
		}
		time.Sleep(waitTime)
	}
	if rootCaStr == "" {
		hwlog.RunLog.Errorf("get valid root ca for websocket service failed: %v", err)
		return "", "", err
	}
	certStr, err = reqCertParams.ReqIssueSvrCert(common.WsSerName, csr)
	if err != nil {
		hwlog.RunLog.Errorf("issue certStr for websocket service cert failed: %v", err)
		return "", "", err
	}

	return certStr, rootCaStr, nil
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
