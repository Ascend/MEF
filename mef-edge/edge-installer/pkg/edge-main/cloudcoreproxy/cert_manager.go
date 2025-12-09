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

// Package cloudcoreproxy
package cloudcoreproxy

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/certmgr"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/configpara"
)

const (
	tokenPartNum = 4
)

type certManager struct {
	caFile   string
	certFile string
	keyFile  string

	token []byte
}

func newCertManager() (certManager, error) {
	cfgDir, err := path.GetCompConfigDir()
	if err != nil {
		return certManager{}, err
	}
	return certManager{
		caFile:   filepath.Join(cfgDir, constants.CloudCoreCertPathName, constants.RootCaName),
		certFile: filepath.Join(cfgDir, constants.CloudCoreCertPathName, constants.ClientCertName),
		keyFile:  filepath.Join(cfgDir, constants.CloudCoreCertPathName, constants.ClientKeyName),
	}, nil
}

func (cm *certManager) start() error {
	for {
		cm.prepareCert()

		if cm.connectTest() {
			hwlog.RunLog.Info("cloud core cert is ready")
			utils.ClearSliceByteMemory(cm.token)
			break
		}

		// delete invalid cert,get valid cert by next loop
		if err := fileutils.DeleteFile(cm.caFile); err != nil {
			hwlog.RunLog.Errorf("delete ca file failed: %v", err)
		}

		if err := fileutils.DeleteFile(cm.certFile); err != nil {
			hwlog.RunLog.Errorf("delete crt file failed: %v", err)
		}

		if err := fileutils.DeleteAllFileWithConfusion(cm.keyFile); err != nil {
			hwlog.RunLog.Errorf("delete key file failed: %v", err)
		}

		time.Sleep(constants.StartWsWaitTime)
	}

	return nil
}

func (cm *certManager) prepareCert() {
	if err := cm.getCaCert(); err != nil {
		hwlog.RunLog.Errorf("get ca cert failed: %v", err)
		return
	}

	if err := cm.getToken(); err != nil {
		hwlog.RunLog.Errorf("get token failed: %v", err)
		return
	}

	if err := cm.getServerCert(); err != nil {
		hwlog.RunLog.Errorf("get server cert failed: %v", err)
		return
	}
}

func (cm *certManager) connectTest() bool {
	if !fileutils.IsExist(cm.keyFile) || !fileutils.IsExist(cm.certFile) || !fileutils.IsExist(cm.caFile) {
		return false
	}

	hwlog.RunLog.Info("start to check connection and certs with cloud core")
	netCfg := configpara.GetNetConfig()
	serverAddr := fmt.Sprintf("https://%s:%d", netCfg.IP, constants.DefaultCloudCoreWsPort)

	certInfo, err := getWsCertInfo()
	if err != nil {
		return false
	}
	if _, err := httpsmgr.GetHttpsReq(serverAddr, *certInfo).Get(nil); err != nil &&
		!strings.HasPrefix(err.Error(), "https return error status code:") {
		hwlog.RunLog.Errorf("establish tls connection with cloud core error: %v", err)
		return false
	}

	hwlog.RunLog.Info("check connection and certs with cloud core success")
	return true
}

func (cm *certManager) getCaCert() error {
	hwlog.RunLog.Info("start to get cloud core ca cert")
	if fileutils.IsExist(cm.caFile) {
		return nil
	}

	var content string
	var err error
	if content, err = cm.getCfgFromCenter("cloud-core-ca"); err != nil {
		return err
	}

	if err = x509.CheckPemCertChain([]byte(content)); err != nil {
		hwlog.RunLog.Errorf("check cloud core ca failed: %s", err.Error())
		return errors.New("check cloud core ca failed")
	}

	if err = fileutils.WriteData(cm.caFile, []byte(content)); err != nil {
		return err
	}

	if err = fileutils.SetPathPermission(cm.caFile, constants.Mode400, false, false); err != nil {
		if err = fileutils.DeleteFile(cm.caFile); err != nil {
			hwlog.RunLog.Errorf("delete ca file failed:%v", err)
		}
		return err
	}
	hwlog.RunLog.Info("get cloud core ca cert success")
	return nil
}

func (cm *certManager) getToken() error {
	hwlog.RunLog.Info("start to get cloud core token")

	var content string
	var err error
	if content, err = cm.getCfgFromCenter("cloud-core-token"); err != nil {
		return err
	}
	defer utils.ClearStringMemory(content)

	cm.token = []byte(content)

	hwlog.RunLog.Info("get cloud core token success")
	return nil
}

func (cm *certManager) getCfgFromCenter(cfgType string) (string, error) {
	msg, err := model.NewMessage()
	if err != nil {
		return "", err
	}

	msg.SetRouter(
		constants.ModCloudCore,
		constants.ModEdgeHub,
		constants.OptGet,
		constants.ResConfig,
	)

	if err = msg.FillContent(cfgType); err != nil {
		return "", fmt.Errorf("fill config type into content failed: %v", err)
	}
	resp, err := modulemgr.SendSyncMessage(msg, constants.CenterSycMsgWaitTime)
	if err != nil {
		return "", fmt.Errorf("send sync message error: %v", err)
	}

	var data string
	if err = resp.ParseContent(&data); err != nil {
		return "", fmt.Errorf("get resp data failed: %v", err)
	}

	return data, nil
}

func (cm *certManager) getServerCert() error {
	hwlog.RunLog.Info("start to get cloud core sever cert")
	if fileutils.IsExist(cm.keyFile) && fileutils.IsExist(cm.certFile) {
		return nil
	}

	tokenParts := strings.Split(string(cm.token), ".")
	defer func() {
		for i := range tokenParts {
			utils.ClearStringMemory(tokenParts[i])
		}
	}()

	if len(tokenParts) != tokenPartNum {
		return errors.New("token part is not four")
	}

	bearerToken := "Bearer " + strings.Join(tokenParts[1:], ".")
	defer utils.ClearStringMemory(bearerToken)
	reqHeaders := map[string]interface{}{
		"Authorization": bearerToken,
	}

	kmcCfg, err := util.GetKmcConfig("")
	if err != nil {
		return errors.New("get kmc config failed")
	}

	csr, err := certutils.CreateCsr(cm.keyFile, constants.MefCertCommonNamePrefix, kmcCfg, certutils.CertSan{})
	if err != nil {
		return fmt.Errorf("failed to create CSR: %v", err)
	}

	tlsCfg := certutils.TlsCertInfo{
		RootCaPath: cm.caFile,
		RootCaOnly: true,
		WithBackup: true,
	}

	netCfg := configpara.GetNetConfig()

	content, err := httpsmgr.GetHttpsReq(fmt.Sprintf("https://%s:%d/edge.crt", netCfg.IP,
		constants.DefaultCloudCoreCertPort), tlsCfg, reqHeaders).Get(bytes.NewReader(csr))
	if err != nil {
		return err
	}

	if err = cm.saveServerCert(content); err != nil {
		return err
	}
	hwlog.RunLog.Info("get cloud core sever cert success")
	return nil
}

func (cm *certManager) saveServerCert(content []byte) error {
	certPem := certutils.PemWrapCert(content)
	if err := certmgr.CheckSrvCert(string(certPem), cm.keyFile, true); err != nil {
		hwlog.RunLog.Errorf("check received srv cert failed: %v", err)
		return errors.New("check received srv cert failed")
	}
	if err := fileutils.WriteData(cm.certFile, certPem); err != nil {
		return err
	}

	if err := fileutils.SetPathPermission(cm.certFile, constants.Mode400, false, false); err != nil {
		return err
	}
	return nil
}
