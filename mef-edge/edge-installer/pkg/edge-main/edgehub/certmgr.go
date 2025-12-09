// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package edgehub this file for edge hub module register
package edgehub

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/certmgr"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/edge-main/common/cloudcert"
	"edge-installer/pkg/edge-main/common/configpara"
)

// cert types and update result types
const (
	CertTypeEdgeCa            = "EdgeCa"
	CertTypeEdgeSvc           = "EdgeSvc"
	UpdateStatusSuccess       = 2
	UpdateStatusFail          = 3
	httpReqTryInterval        = time.Minute
	httpReqTryMaxTime         = 5
	inUpdating          int64 = 1
	notUpdating         int64 = 0
)

var (
	edgeSvcCertUpdating int64
	edgeCaCertUpdating  int64
)

type certUpdateResult struct {
	CertType   string `json:"certType"`
	Sn         string `json:"sn"`
	ResultCode int64  `json:"resultCode"`
	Desc       string `json:"desc"`
}

type certUpdatePayload struct {
	CertType  string `json:"certType"`
	CaContent string `json:"caContent"`
}

func doCertUpdate(msg *model.Message) error {
	var updateErr error
	payload, err := parseCertUpdatePayload(msg)
	if err != nil {
		updateErr = fmt.Errorf("parse update notify message payload error: %v", err)
		hwlog.RunLog.Error(updateErr)
		return updateErr
	}
	skipReportResult := false
	defer func() {
		if err := reportCertUpdateResult(skipReportResult, updateErr, payload, msg); err != nil {
			hwlog.RunLog.Errorf("report cert update result failed error: %v", payload.CertType)
		}
	}()
	switch payload.CertType {
	case CertTypeEdgeCa:
		// in case of repeated cert update notify received when cert update is in process
		if !atomic.CompareAndSwapInt64(&edgeCaCertUpdating, notUpdating, inUpdating) {
			hwlog.RunLog.Warn("edge root ca is in updating... skip repeatedly update operation")
			skipReportResult = true
			return nil
		}
		updateErr = doRootCaCertUpdate(*payload)
		doCertUpdateOptLog(updateErr, payload.CertType, msg)
		atomic.StoreInt64(&edgeCaCertUpdating, notUpdating)
		return updateErr
	case CertTypeEdgeSvc:
		// in case of repeated cert update notify received when cert update is in process
		if !atomic.CompareAndSwapInt64(&edgeSvcCertUpdating, notUpdating, inUpdating) {
			hwlog.RunLog.Warn("edge service cert is in updating... skip repeatedly update operation")
			skipReportResult = true
			return nil
		}
		updateErr = doServiceCertUpdate()
		doCertUpdateOptLog(updateErr, payload.CertType, msg)
		atomic.StoreInt64(&edgeSvcCertUpdating, notUpdating)
		return updateErr
	default:
		updateErr = fmt.Errorf("invalid update notify cert type: %v", payload.CertType)
		hwlog.RunLog.Error(updateErr)
		return updateErr
	}
}

func parseCertUpdatePayload(msg *model.Message) (*certUpdatePayload, error) {
	var payload certUpdatePayload
	if err := msg.ParseContent(&payload); err != nil {
		return nil, fmt.Errorf("parse update notify message content error: %v", err)
	}
	return &payload, nil
}

func getNewCertViaWs() (*certutils.TlsCertInfo, error) {
	newCertInfo, err := cloudcert.GetEdgeHubCertInfo(true)
	if err != nil {
		hwlog.RunLog.Errorf("get edgehub back cert path failed: %v", err)
		return nil, errors.New("get edgehub back cert path failed")
	}

	csrData, err := certutils.CreateCsr(newCertInfo.KeyPath, constants.MefCertCommonNamePrefix,
		newCertInfo.KmcCfg, certutils.CertSan{})
	if err != nil {
		hwlog.RunLog.Errorf("generate edgehub csr data failed: %v", err)
		return nil, errors.New("generate edgehub csr data failed")
	}

	csrMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message error: %v", err)
		return nil, errors.New("create message error")
	}
	csrMsg.SetRouter(
		constants.ModEdgeHub,
		constants.ModEdgeHub,
		constants.OptPost,
		constants.ResEdgeCert)
	installConfig := configpara.GetInstallerConfig()
	csrMsg.SetNodeId(installConfig.SerialNumber)
	if err = csrMsg.FillContent(csrData); err != nil {
		hwlog.RunLog.Errorf("fill csr data into content failed: %v", err)
		return nil, errors.New("fill csr data into content failed")
	}
	resp, err := modulemgr.SendSyncMessage(csrMsg, constants.WsSycMsgWaitTime)
	if err != nil {
		hwlog.RunLog.Errorf("send sync message to cloudhub error: %v", err)
		return nil, errors.New("send sync message to cloudhub error")
	}
	var certStr string
	if err = resp.ParseContent(&certStr); err != nil {
		hwlog.RunLog.Errorf("get resp content failed: %v", err)
		return nil, errors.New("get resp content failed")
	}

	if err = certmgr.CheckSrvCert(certStr, newCertInfo.KeyPath, false); err != nil {
		return nil, err
	}
	if err = fileutils.WriteData(newCertInfo.CertPath, []byte(certStr)); err != nil {
		hwlog.RunLog.Errorf("write new cert data error: %v", err)
		return nil, errors.New("write new cert data error")
	}
	return newCertInfo, nil
}

func prepareNewCert(inuseCertInfo, tempCert *certutils.TlsCertInfo) error {
	if inuseCertInfo == nil || tempCert == nil {
		return errors.New("invalid cert info")
	}
	oldKeyPath, err := fileutils.CheckOriginPath(inuseCertInfo.KeyPath)
	if err != nil {
		return err
	}
	oldCertPath, err := fileutils.CheckOriginPath(inuseCertInfo.CertPath)
	if err != nil {
		return err
	}
	tempKeyPath, err := fileutils.CheckOriginPath(tempCert.KeyPath)
	if err != nil {
		return err
	}
	tempCertPath, err := fileutils.CheckOriginPath(tempCert.CertPath)
	if err != nil {
		return err
	}
	// key file need write permission before call DeleteAllFileWithConfusion
	if err := fileutils.SetPathPermission(oldKeyPath, constants.Mode600, false, false); err != nil {
		return err
	}
	if err := fileutils.DeleteAllFileWithConfusion(oldKeyPath); err != nil {
		return err
	}
	if err := fileutils.DeleteFile(oldCertPath); err != nil {
		return err
	}
	if err := fileutils.CopyFile(tempKeyPath, oldKeyPath); err != nil {
		return err
	}
	if err := fileutils.CopyFile(tempCertPath, oldCertPath); err != nil {
		return err
	}
	// key file need write permission before call DeleteAllFileWithConfusion
	if err := fileutils.SetPathPermission(tempKeyPath, constants.Mode600, false, false); err != nil {
		return err
	}
	if err := fileutils.DeleteAllFileWithConfusion(tempKeyPath); err != nil {
		return err
	}
	if err := fileutils.DeleteFile(tempCertPath); err != nil {
		return err
	}
	return nil
}

func updateRootCa(caCertBytes []byte) error {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		return err
	}
	if len(caCertBytes) == 0 {
		return fmt.Errorf("invalid new ca cert data")
	}
	if err := x509.CheckPemCertChain(caCertBytes); err != nil {
		hwlog.RunLog.Errorf("root ca cert check failed: %v", err)
		return fmt.Errorf("root ca cert check failed: %v", err)
	}
	cloudCaPath := configPathMgr.GetHubSvrRootCertPath()
	cloudCaBackupPath := configPathMgr.GetHubSvrRootCertBackupPath()
	if !fileutils.IsExist(cloudCaPath) || !fileutils.IsExist(cloudCaBackupPath) {
		return fmt.Errorf("cloud root cert or backup cert not exists")
	}
	if err = fileutils.SetPathPermission(cloudCaPath, constants.Mode600, false, false); err != nil {
		return fmt.Errorf("set cloud root cert permission error: %v", err)
	}
	if err = fileutils.SetPathPermission(cloudCaBackupPath, constants.Mode600, false, false); err != nil {
		return fmt.Errorf("set cloud backup root cert permission error: %v", err)
	}
	defer func() {
		if err = fileutils.SetPathPermission(cloudCaPath, constants.Mode400, false, false); err != nil {
			hwlog.RunLog.Errorf("reset cloud root cert permission error: %v", err)
		}
		if err = fileutils.SetPathPermission(cloudCaBackupPath, constants.Mode400, false, false); err != nil {
			hwlog.RunLog.Errorf("reset cloud backup root cert permission error: %v", err)
		}
	}()
	if err = fileutils.WriteData(cloudCaPath, caCertBytes); err != nil {
		return fmt.Errorf("write ca data to cert file error:%v", err)
	}
	if err = fileutils.WriteData(cloudCaBackupPath, caCertBytes); err != nil {
		return fmt.Errorf("write ca data to backup cert file error: %v", err)
	}
	return nil
}

func reportCertUpdateResult(skipReport bool, updateErr error, payload *certUpdatePayload, msg *model.Message) error {
	// in case of reporting an incorrect result when cert update is in process then receive a repeat notify
	if skipReport {
		return nil
	}
	result := certUpdateResult{
		CertType:   payload.CertType,
		Sn:         msg.GetNodeId(),
		ResultCode: UpdateStatusSuccess,
	}
	if updateErr != nil {
		result.ResultCode = UpdateStatusFail
		result.Desc = fmt.Sprintf("cert [%v] update failed", payload.CertType)
	}
	respMsg, err := model.NewMessage()
	if err != nil {
		return fmt.Errorf("create new message failed, error: %v", err)
	}
	respMsg.SetRouter(
		constants.ModEdgeHub,
		constants.ModEdgeHub,
		constants.OptResponse,
		constants.ResCertUpdate)
	installConfig := configpara.GetInstallerConfig()
	respMsg.SetNodeId(installConfig.SerialNumber)
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("serialize cert update result message failed: %v", err)
	}
	var base64Data []byte
	base64.StdEncoding.Encode(base64Data, jsonData)
	if err = respMsg.FillContent(base64Data); err != nil {
		return fmt.Errorf("fill content failed: %v", err)
	}
	return sendMsgToServer(respMsg)
}

func doRootCaCertUpdate(payload certUpdatePayload) error {
	if err := updateRootCa([]byte(payload.CaContent)); err != nil {
		return err
	}
	if err := proxy.ProxyCfg.UpdateTlsCa([]byte(payload.CaContent)); err != nil {
		return err
	}
	return nil
}

func doServiceCertUpdate() error {
	netConfig, err := getConfig()
	if err != nil {
		return fmt.Errorf("init edge hub ws client failed: %v", err)
	}

	var tempCertInfo *certutils.TlsCertInfo
	var tryCnt int
	// try multiple times to wait for nginx new root cert is ready
	for tryCnt = 0; tryCnt < httpReqTryMaxTime; tryCnt++ {
		tempCertInfo, err = getNewCertViaWs()
		if err != nil {
			hwlog.RunLog.Errorf("getNewCertViaWs error: %v", err)
			time.Sleep(httpReqTryInterval)
			continue
		}
		if err = checkEdgeCertValid(netConfig, tempCertInfo); err != nil {
			hwlog.RunLog.Errorf("checkEdgeCertValid error: %v", err)
			time.Sleep(httpReqTryInterval)
			continue
		}
		break
	}
	if tryCnt == httpReqTryMaxTime {
		if tempCertInfo != nil {
			// cleanup temp cert and key files
			if err = fileutils.DeleteAllFileWithConfusion(tempCertInfo.KeyPath); err != nil {
				hwlog.RunLog.Errorf("cleanup temp key file [%v] failed: %v", tempCertInfo.KeyPath, err)
			}
			if err = fileutils.DeleteFile(tempCertInfo.CertPath); err != nil {
				hwlog.RunLog.Errorf("cleanup temp cert file [%v] failed: %v", tempCertInfo.CertPath, err)
			}
		}
		hwlog.RunLog.Error("edge service cert update failed. valid service cert cannot be retrieved")
		return fmt.Errorf("edge service cert update failed. valid service cert cannot be retrieved")
	}

	certInfo, err := cloudcert.GetEdgeHubCertInfo()
	if err != nil {
		hwlog.RunLog.Errorf("get old cert info failed: %v", err)
		return fmt.Errorf("get old cert info failed: %v", err)
	}
	// service cert is updated on disk, it will be used on next time tls handshake
	if err = prepareNewCert(certInfo, tempCertInfo); err != nil {
		return fmt.Errorf("prepareNewCert error: %v", err)
	}
	return nil
}

func doCertUpdateOptLog(updateErr error, certType string, msg *model.Message) {
	switch certType {
	case CertTypeEdgeCa:
		if updateErr != nil {
			hwlog.OpLog.Errorf("[%v@%v][%v %v %v, cert type: cloud root cert]", configpara.GetNetConfig().NetType,
				configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), constants.Failed)
		} else {
			hwlog.OpLog.Infof("[%v@%v][%v %v %v, cert type: cloud root cert]", configpara.GetNetConfig().NetType,
				configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), constants.Success)
		}
	case CertTypeEdgeSvc:
		if updateErr != nil {
			hwlog.OpLog.Errorf("[%v@%v][%v %v %v, cert type: edge service cert]", configpara.GetNetConfig().NetType,
				configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), constants.Failed)
		} else {
			hwlog.OpLog.Infof("[%v@%v][%v %v %v, cert type: edge service cert]", configpara.GetNetConfig().NetType,
				configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), constants.Success)
		}
	default:
		hwlog.OpLog.Errorf("[%v@%v][%v %v, invalid cert type: %v]", configpara.GetNetConfig().NetType,
			configpara.GetNetConfig().IP, msg.GetOption(), msg.GetResource(), certType)
	}
}
