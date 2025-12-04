// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certmgr this file for cert manager
package certmgr

import (
	"errors"
	"os"
	"path"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

// CertManager the cert manager struct
type CertManager struct {
	certDir        string
	certName       string
	certBackUpName string
}

// NewCertMgr create cert manager instance
func NewCertMgr(certDir, certName, certBackUpName string) *CertManager {
	return &CertManager{
		certDir:        certDir,
		certName:       certName,
		certBackUpName: certBackUpName,
	}
}

// IsCertExist check whether the cert exists
func (cm *CertManager) IsCertExist() bool {
	certPath := path.Join(cm.certDir, cm.certName)
	return fileutils.IsExist(certPath)
}

// LoadCert load file content from main file path or back up file path
func (cm *CertManager) LoadCert() ([]byte, error) {
	certPath := path.Join(cm.certDir, cm.certName)
	certBackUpPath := path.Join(cm.certDir, cm.certBackUpName)
	buInstance, err := x509.NewBKPInstance([]byte{}, certPath, certBackUpPath)
	if err != nil {
		hwlog.RunLog.Errorf("create backup instance failed, error: %v", err)
		return nil, errors.New("create backup instance failed")
	}
	data, err := buInstance.ReadFromDisk(constants.CertFileMode, true)
	if err != nil {
		hwlog.RunLog.Errorf("read cert data from disk failed, error: %v", err)
		return nil, errors.New("read cert data from disk failed")
	}
	return data, nil
}

// SaveCertByFile import and backup the cert by temp cert file
func (cm *CertManager) SaveCertByFile(tempCertPath string, mod ...os.FileMode) error {
	certBytes, err := fileutils.LoadFile(tempCertPath)
	if err != nil {
		return err
	}
	var certFileMode os.FileMode = constants.CertFileMode
	if len(mod) > 0 {
		certFileMode = mod[0]
	}
	return cm.SaveCertByContent(certBytes, certFileMode)
}

// SaveCertByContent import and backup the cert by cert content
func (cm *CertManager) SaveCertByContent(certByte []byte, mod os.FileMode) error {
	if err := fileutils.CreateDir(cm.certDir, constants.CertDirMode); err != nil {
		hwlog.RunLog.Errorf("create cert directory [%s] failed, error: %v", cm.certDir, err)
		return errors.New("create cert directory failed")
	}

	certPath := path.Join(cm.certDir, cm.certName)
	certBackUpPath := path.Join(cm.certDir, cm.certBackUpName)
	buInstance, err := x509.NewBKPInstance(certByte, certPath, certBackUpPath)
	if err != nil {
		hwlog.RunLog.Errorf("create backup instance failed, error: %v", err)
		return errors.New("create backup instance failed")
	}
	if err = buInstance.WriteToDisk(mod, false); err != nil {
		hwlog.RunLog.Errorf("write cert data to disk failed, error: %v", err)
		return errors.New("write cert data to disk failed")
	}

	return nil
}

// CheckSrvCert is the func to check received service cert
// allowFutureEffective indicates whether allowed the cert to be valid when 'now' is before cert.NotBefore time
// while allowed,maximum 24 hours ahead
// current is only effect on cloud-core-cert
func CheckSrvCert(certStr, keyPath string, allowFutureEffective bool) error {
	kmcConfig, err := util.GetKmcConfig("")
	if err != nil {
		hwlog.RunLog.Errorf("get kmc config failed: %v", err)
		return errors.New("get kmc config failed")
	}
	checkTask := x509.CheckSvcCertTask{
		KeyPath:              keyPath,
		SvcCertData:          []byte(certStr),
		KmcConfig:            kmcConfig,
		AllowFutureEffective: allowFutureEffective,
	}

	if err = checkTask.RunTask(); err != nil {
		hwlog.RunLog.Errorf("check srv cert failed: %v", err)
		return errors.New("check srv cert failed")
	}
	return nil
}
