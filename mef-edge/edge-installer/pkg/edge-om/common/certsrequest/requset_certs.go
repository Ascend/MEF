// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package certsrequest requests certs for edge
package certsrequest

import (
	"errors"
	"path/filepath"
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

const (
	requestCertWaitTime = 30 * time.Second
	maxRetryTimes       = 3
)

var (
	certNames     = []string{constants.SoftwareCertName, constants.ImageCertName}
	requiredCerts map[string]struct{}
	// certLock for r/w of requiredCerts
	certLock sync.Mutex
	// requestLock allows only one go-routine to request remote cert from center
	requestLock sync.Mutex
)

// SetReceivedCert [method] delete required certs from request array
func SetReceivedCert(certName string) {
	certLock.Lock()
	if requiredCerts != nil {
		delete(requiredCerts, certName)
	}
	certLock.Unlock()
}

func resetRequiredCerts() {
	certLock.Lock()
	requiredCerts = make(map[string]struct{})
	for _, certName := range certNames {
		requiredCerts[certName] = struct{}{}
	}
	certLock.Unlock()
}

// RequestCertsFromCenter update image and software certs for MEFEdge when net manager type is MEF
func RequestCertsFromCenter() {
	resetRequiredCerts()

	if !requestLock.TryLock() {
		hwlog.RunLog.Info("reconnect to the center, reset required certs' array")
		return
	}
	defer requestLock.Unlock()
	certRootDir, err := path.GetCompSpecificDir(constants.ConfigCertPathName)
	if err != nil {
		hwlog.RunLog.Errorf("get cert root directory failed, error: %v", err)
		return
	}
	requestCount := 0
	for ; requestCount < maxRetryTimes; requestCount++ {
		// wait center send certs first
		time.Sleep(requestCertWaitTime)
		// all required certs are received
		if len(requiredCerts) == 0 {
			hwlog.RunLog.Info("all certs received successfully from center")
			return
		}
		hwlog.RunLog.Info("start to request certs from center")
		certLock.Lock()
		for certName := range requiredCerts {
			if err := sendDownloadCertMsg(certName); err != nil {
				hwlog.RunLog.Errorf("send cert %s massage failed", certName)
			}
		}
		certLock.Unlock()
	}
	if requestCount == maxRetryTimes {
		makeSureLocalCerts(certRootDir)
	}
}

// makeSureLocalCerts check and restored local certs when request certs from center failed
func makeSureLocalCerts(certDir string) []string {
	names := make([]string, 0, len(certNames))
	for _, certName := range certNames {
		certFilePath := filepath.Join(certDir, certName, constants.RootCaName)
		if _, err := certutils.GetCertContentWithBackup(certFilePath); err != nil {
			hwlog.RunLog.Warnf("%s cert not exist", certName)
			names = append(names, certName)
		}
	}
	return names
}

func sendDownloadCertMsg(certName string) error {
	sendMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create new message failed, error: %v", err)
		return err
	}
	sendMsg.SetRouter(constants.ModEdgeOm, constants.InnerClient, constants.OptGet, constants.ResDownloadCert)
	if err = sendMsg.FillContent(certName); err != nil {
		hwlog.RunLog.Errorf("fill cert name into content failed: %v", err)
		return errors.New("fill cert name into content failed")
	}
	if err = modulemgr.SendMessage(sendMsg); err != nil {
		hwlog.RunLog.Errorf("%s sends message to %s failed", constants.ModEdgeOm, constants.InnerClient)
		return err
	}
	return nil
}
