// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certchecker cert config checker
package certchecker

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"sync"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
)

var lock sync.Mutex

// NewImportCertChecker [method] for getting import cert checker struct
func NewImportCertChecker() *importCertChecker {
	return &importCertChecker{}
}

// NewDeleteCertChecker [method] for getting delete cert checker struct
func NewDeleteCertChecker() *deleteCertChecker {
	return &deleteCertChecker{}
}

// NewIssueCertChecker [method] for getting issue cert checker struct
func NewIssueCertChecker() *issueCertChecker {
	return &issueCertChecker{}
}

type importCertChecker struct {
	certChecker checker.ModelChecker
}

func (icc *importCertChecker) init() {
	icc.certChecker.Checker = checker.GetAndChecker(
		GetStringChecker("CertName", checkIfCanImport, true),
		GetStringChecker("Cert", certContentChecker, true),
	)
}

func (icc *importCertChecker) Check(data interface{}) checker.CheckResult {
	icc.init()
	checkResult := icc.certChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("cert checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

type deleteCertChecker struct {
	certChecker checker.ModelChecker
}

func (dcc *deleteCertChecker) init() {
	dcc.certChecker.Checker = checker.GetAndChecker(
		GetStringChecker("Type", checkIfCanDelete, true),
	)
}

func (dcc *deleteCertChecker) Check(data interface{}) checker.CheckResult {
	dcc.init()
	checkResult := dcc.certChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("cert checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

type issueCertChecker struct {
	certChecker checker.ModelChecker
}

func (icc *issueCertChecker) init() {
	icc.certChecker.Checker = checker.GetAndChecker(
		GetStringChecker("CertName", CheckCertName, true),
		GetStringChecker("Csr", csrChecker, true),
	)
}

func (icc *issueCertChecker) Check(data interface{}) checker.CheckResult {
	icc.init()
	checkResult := icc.certChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("cert checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func csrChecker(csr string) error {
	csrLen := len(csr)
	if csrLen < minCsrLen || csrLen > maxCsrLen {
		hwlog.RunLog.Errorf("csr checker check failed: the length is not in range [%d, %d]", minCsrLen, maxCsrLen)
		return fmt.Errorf("the length is out of range [%d, %d]", minCsrLen, maxCsrLen)
	}
	pattern := regexp.MustCompile(csrReg)
	if match := pattern.MatchString(csr); !match {
		hwlog.RunLog.Error("csr checker check failed: not meet regex")
		return errors.New("csr is not meet regex required")
	}
	if _, err := base64.StdEncoding.DecodeString(csr); err != nil {
		hwlog.RunLog.Errorf("base64 decode csr failed, error:%v", err)
		return err
	}
	return nil
}

func certContentChecker(certContent string) error {
	lock.Lock()
	defer lock.Unlock()
	// base64 decode root certificate content
	caBase64, err := base64.StdEncoding.DecodeString(certContent)
	if err != nil {
		hwlog.RunLog.Errorf("base64 decode ca content failed, error:%v", err)
		return err
	}

	caMgr, err := x509.NewCaChainMgr(caBase64)
	if err != nil {
		hwlog.RunLog.Errorf("new ca chain mgr failed: %s", err.Error())
		return err
	}

	if err = caMgr.CheckCertChain(); err != nil {
		hwlog.RunLog.Errorf("check importing certs failed: %s", err.Error())
		return err
	}

	return nil
}

var onlineImportMap = map[string]bool{
	common.WsSerName:        false,
	common.WsCltName:        false,
	common.SoftwareCertName: true,
	common.ImageCertName:    true,
	common.NginxCertName:    false,
	common.InnerName:        false,
	common.NorthernCertName: false,
}

var certDeleteNames = []string{common.SoftwareCertName, common.ImageCertName}

var certExportNames = []string{common.WsSerName}

var certInfoNames = map[string]struct{}{
	common.NorthernCertName: {},
}

// CheckCertName check cert name if valid
func CheckCertName(certName string) error {
	_, ok := onlineImportMap[certName]
	if !ok {
		return errors.New("target cert is invalid")
	}
	return nil
}

func checkIfCanImport(certName string) error {
	allowed, ok := onlineImportMap[certName]
	if !ok {
		return errors.New("target cert is invalid")
	}
	if !allowed {
		return errors.New("target cert is no allowed to be imported online")
	}
	return nil
}

func checkIfCanDelete(certName string) error {
	for _, name := range certDeleteNames {
		if name == certName {
			return nil
		}
	}
	return errors.New("target cert is invalid to be deleted")
}

// CheckIfCanExport check cert name if can export
func CheckIfCanExport(certName string) bool {
	for _, name := range certExportNames {
		if name == certName {
			return true
		}
	}
	return false
}

// CheckIfCanGetInfo check cert name if can get summary info
func CheckIfCanGetInfo(certName string) bool {
	_, ok := certInfoNames[certName]
	return ok
}
