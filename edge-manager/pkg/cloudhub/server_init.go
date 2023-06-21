// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package cloudhub server init
package cloudhub

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/httpsmgr"
	centerutil "huawei.com/mindxedge/base/mef-center-install/pkg/util"

	"edge-manager/pkg/util"
)

const (
	name     = "server_edge_ctl"
	maxRetry = 10
	waitTime = 5 * time.Second
)

var serverSender websocketmgr.WsSvrSender
var initFlag bool
var certPathDir = filepath.Join("/home/data/config", centerutil.WebsocketCerts)

// InitServer init server
func InitServer() error {
	checkAndSetWsSvcCert()
	rootCaBytes, err := getWsRootCert()
	if err != nil {
		return err
	}
	certInfo := certutils.TlsCertInfo{
		KmcCfg:        kmc.GetDefKmcCfg(),
		RootCaContent: rootCaBytes,
		CertPath:      filepath.Join(certPathDir, centerutil.ServiceName),
		KeyPath:       filepath.Join(certPathDir, centerutil.KeyFileName),
		SvrFlag:       true,
	}
	go authServer(rootCaBytes)

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
	if err = proxy.AddHandler(util.ConnCheckUrl, checkConn); err != nil {
		hwlog.RunLog.Errorf("add handler failed")
		return fmt.Errorf("add handler failed")
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

func authServer(rootCaBytes []byte) {
	authCertInfo := certutils.TlsCertInfo{
		KmcCfg:        kmc.GetDefKmcCfg(),
		RootCaContent: rootCaBytes,
		CertPath:      filepath.Join(certPathDir, centerutil.ServiceName),
		KeyPath:       filepath.Join(certPathDir, centerutil.KeyFileName),
		SvrFlag:       true,
		IgnoreCltCert: true,
	}
	NewClientAuthService(server.authPort, authCertInfo).Start()
}

func checkAndSetWsSvcCert() {
	keyPath := filepath.Join(certPathDir, centerutil.KeyFileName)
	certPath := filepath.Join(certPathDir, centerutil.ServiceName)
	if utils.IsExist(keyPath) && utils.IsExist(certPath) {
		hwlog.RunLog.Info("check websocket server certs success")
		return
	}
	hwlog.RunLog.Warn("check websocket server certs failed, start to create")
	svcCertStr, err := getWsSvcCert(keyPath)
	if err != nil {
		return
	}
	if err = common.WriteData(certPath, []byte(svcCertStr)); err != nil {
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
		},
	}
	var svcCertStr string
	san := certutils.CertSan{DnsName: []string{common.EdgeMgrDns}}
	ips, err := common.GetHostIpV4()
	if err != nil {
		return "", err
	}
	san.IpAddr = ips
	csr, err := certutils.CreateCsr(keyPath, common.WsSerSerName, nil, san)
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

func checkConn(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	hwlog.RunLog.Info("successfully receive connection test req from mef edge")
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
