// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package edgehub this file for initialize websocket client
package edgehub

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/websocketmgr"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/certmgr"
	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/cloudcert"
	"edge-installer/pkg/edge-main/common/configpara"
)

var proxy *websocketmgr.WsClientProxy
var errCloudhubAuth = errors.New("cloudhub auth failed")

func initClient(netConfig *config.NetManager) error {
	certInfo, err := cloudcert.GetEdgeHubCertInfo()
	if err != nil {
		hwlog.RunLog.Errorf("getEdgeHubCertPath failed: %v", err)
		return errors.New("getEdgeHubCertPath failed")
	}
	if err = checkAndGetCert(netConfig, certInfo); err != nil {
		hwlog.RunLog.Errorf("check edgehub client cert error: %v", err)
		return fmt.Errorf("check edgehub client cert error: %w", err)
	}
	sn := configpara.GetInstallerConfig().SerialNumber
	if sn == "" {
		hwlog.RunLog.Error("sn is invalid")
		return fmt.Errorf("sn is invalid")
	}
	proxyConfig, err := websocketmgr.InitProxyConfig(sn, netConfig.IP, netConfig.Port, *certInfo)
	if err != nil {
		hwlog.RunLog.Errorf("init proxy config failed, error: %v", err)
		return errors.New("init proxy config failed")
	}
	proxyConfig.RegModInfos(getRegModuleInfoList())
	if err := proxyConfig.SetBandwidthLimiterCfg(constants.MaxMsgThroughput, constants.MsgThroughputPeriod); err != nil {
		hwlog.RunLog.Errorf("init tps limiter config failed, error: %v", err)
		return fmt.Errorf("init tps limiter config failed, error: %v", err)
	}
	proxy = &websocketmgr.WsClientProxy{
		ProxyCfg: proxyConfig,
	}
	proxy.SetReConnCallback(reportSoftwareVersion)
	if err = proxy.Start(); err != nil {
		hwlog.RunLog.Errorf("init edgehub client failed: %v", err)
		return errors.New("init edgehub client failed")
	}

	return nil
}

func getCltSender() (*websocketmgr.WsClientProxy, error) {
	if proxy != nil && !proxy.IsConnected() {
		return nil, errors.New("edgehub ws client proxy is nil or client is not connected")
	}
	return proxy, nil
}

func checkAndGetCert(netCfg *config.NetManager, certInfo *certutils.TlsCertInfo) error {
	if netCfg == nil || certInfo == nil {
		return errors.New("invalid NetManager or TlsCertInfo parameter")
	}
	if fileutils.IsExist(certInfo.KeyPath) && fileutils.IsExist(certInfo.CertPath) {
		if err := checkEdgeCertValid(netCfg, certInfo); err == nil {
			hwlog.RunLog.Info("mef edgehub cert exists and valid")
			return nil
		}
		hwlog.RunLog.Warnf("mef edgehub cert is invalid. try to auth from center")
		if fileutils.DeleteFile(certInfo.CertPath) != nil ||
			fileutils.DeleteAllFileWithConfusion(certInfo.KeyPath) != nil {
			return fmt.Errorf("remove invalid mef edgehub cert failed")
		}
		hwlog.RunLog.Info("delete invalid edgehub cert success")
	}
	if err := getCertFromCenter(netCfg, certInfo); err != nil {
		hwlog.RunLog.Error(err)
		return fmt.Errorf("get cert from center error: %w", err)
	}
	return nil
}

func getCertFromCenter(netCfg *config.NetManager, certInfo *certutils.TlsCertInfo) error {
	hwlog.RunLog.Info("start to auth from center")
	url := fmt.Sprintf("https://%s:%d%s", netCfg.IP, netCfg.AuthPort, constants.MefCenterTokenUrl)
	tlsCfg := certutils.TlsCertInfo{
		RootCaPath: certInfo.RootCaPath,
		RootCaOnly: true,
		WithBackup: true,
	}
	tokenStr := string(netCfg.Token)
	defer utils.ClearStringMemory(tokenStr)
	reqHeaders := map[string]interface{}{
		constants.Token: tokenStr,
	}
	csrData, err := certutils.CreateCsr(certInfo.KeyPath, constants.MefCertCommonNamePrefix,
		certInfo.KmcCfg, certutils.CertSan{})
	if err != nil {
		return err
	}
	resp, err := httpsmgr.GetHttpsReq(url, tlsCfg, reqHeaders).PostJson(csrData)
	if err != nil {
		if strings.Contains(err.Error(), strconv.Itoa(http.StatusUnauthorized)) {
			return fmt.Errorf("%w: token is incorrect", errCloudhubAuth)
		}
		if strings.Contains(err.Error(), strconv.Itoa(http.StatusLocked)) {
			return fmt.Errorf("%w: ip has too many auth error, locked by center for a while", errCloudhubAuth)
		}
		return err
	}
	// svc cert obtained from center has already got a -24 hours lead by now, so will not allow graceful
	if err = certmgr.CheckSrvCert(string(resp), certInfo.KeyPath, false); err != nil {
		hwlog.RunLog.Errorf("check received srv cert failed: %v", err)
		return errors.New("check received srv cert failed")
	}
	if err = fileutils.WriteData(certInfo.CertPath, resp); err != nil {
		return err
	}
	hwlog.RunLog.Info("get edgehub client cert from center success")
	return nil
}

// handshake with edge manager to check connection and certs
func checkEdgeCertValid(netCfg *config.NetManager, certInfo *certutils.TlsCertInfo) error {
	if netCfg == nil || certInfo == nil {
		return errors.New("invalid NetManager or TlsCertInfo")
	}
	if !fileutils.IsExist(certInfo.KeyPath) || !fileutils.IsExist(certInfo.CertPath) {
		return errors.New("edge cert or key file not exists")
	}
	hwlog.RunLog.Info("start to check connection and certs with edge manager")
	serverAddr := fmt.Sprintf("https://%s:%d", netCfg.IP, netCfg.Port)

	if _, err := httpsmgr.GetHttpsReq(serverAddr, *certInfo).Get(nil); err != nil {
		hwlog.RunLog.Errorf("establish tls connection with mef center error: %v", err)
		return errors.New("establish tls connection with mef center failed, please check net or cert")
	}
	hwlog.RunLog.Info("check connection and certs with edge manager success")
	return nil
}

func reportSoftwareVersion() {
	msg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:      constants.ModEdgeHub,
		Destination: constants.ModEdgeOm,
		Operation:   constants.OptReport,
		Resource:    constants.InnerSoftwareVersion,
		Content:     nil,
	})
	if err != nil {
		hwlog.RunLog.Errorf("report software version failed, create message error: %v", err)
		return
	}
	if err = modulemgr.SendAsyncMessage(msg); err != nil {
		hwlog.RunLog.Errorf("report software version failed, send message error: %v", err)
		return
	}
}
