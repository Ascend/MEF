// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package certchecker cert config checker
package certchecker

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker/checker"
)

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
		GetStringChecker("Type", CheckCertName, true),
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

func csrChecker(csr string) bool {
	csrLen := len(csr)
	if csrLen < minCsrLen || csrLen > maxCsrLen {
		hwlog.RunLog.Errorf("csr checker check failed: the length is not in range [%d, %d]", minCsrLen, maxCsrLen)
		return false
	}
	pattern := regexp.MustCompile(csrReg)
	if match := pattern.MatchString(csr); !match {
		hwlog.RunLog.Error("csr checker check failed: not meet regex")
		return false
	}
	if _, err := base64.StdEncoding.DecodeString(csr); err != nil {
		hwlog.RunLog.Errorf("base64 decode csr failed, error:%v", err)
		return false
	}
	return true
}

func certContentChecker(certContent string) bool {
	// base64 decode root certificate content
	caBase64, err := base64.StdEncoding.DecodeString(certContent)
	if err != nil {
		hwlog.RunLog.Errorf("base64 decode ca content failed, error:%v", err)
		return false
	}
	if len(caBase64) == 0 || len(caBase64) > maxCertSize {
		hwlog.RunLog.Error("valid ca file size failed")
		return false
	}
	// verifying root certificate content
	if err := x509.VerifyCaCert(caBase64, x509.InvalidNum); err != nil {
		hwlog.RunLog.Errorf("valid ca certification failed, error:%v", err)
		return false
	}
	return true
}

var certImportMap = map[string]bool{
	common.WsSerName:        false,
	common.WsCltName:        false,
	common.SoftwareCertName: true,
	common.ImageCertName:    true,
	common.NginxCertName:    true,
	common.InnerName:        false,
}

// CheckCertName check use id if valid
func CheckCertName(certName string) bool {
	_, ok := certImportMap[certName]
	return ok
}

func checkIfCanImport(certName string) bool {
	v, ok := certImportMap[certName]
	if !ok {
		return false
	}
	return v
}
