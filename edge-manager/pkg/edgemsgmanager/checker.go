// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager handler
package edgemsgmanager

import (
	"fmt"
	"math"

	"huawei.com/mindx/common/checker"

	"huawei.com/mindxedge/base/common"
)

// NewDownloadChecker is the struct to init a DownloadChecker
func NewDownloadChecker() *DownloadChecker {
	return &DownloadChecker{}
}

// DownloadChecker is the checker to check the message of download request
type DownloadChecker struct {
	modelChecker checker.ModelChecker
}

func (d *DownloadChecker) init() {
	d.modelChecker.Checker = checker.GetAndChecker(
		checker.GetUniqueListChecker("SerialNumbers",
			checker.GetRegChecker("", `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`, true),
			1, common.MaxNode, true),
		checker.GetStringChoiceChecker("SoftwareName",
			[]string{common.MEFEdge}, true),
		GetDownloadInfoChecker("DownloadInfo", true),
	)
}

// Check is the main func for a checker to execute checking operation
func (d *DownloadChecker) Check(data interface{}) checker.CheckResult {
	d.init()

	checkResult := d.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("software downloadinfo check failed: %s", checkResult.Reason))
	}

	return checker.NewSuccessResult()
}

type downloadInfoChecker struct {
	modelChecker checker.ModelChecker
}

func (d *downloadInfoChecker) init() {
	const minPwdLength = 1
	const maxPwdLength = 20
	d.modelChecker.Checker = checker.GetAndChecker(
		checker.GetHttpsUrlChecker("Package", true),
		checker.GetHttpsUrlChecker("SignFile", true),
		checker.GetHttpsUrlChecker("CrlFile", true),
		checker.GetRegChecker("UserName", "^[a-zA-Z0-9]{6,32}$", true),
		checker.GetListChecker("Password",
			checker.GetUintChecker("", 0, math.MaxUint8, true),
			minPwdLength,
			maxPwdLength,
			true,
		),
	)
}

// Check [method] check main function
func (d *downloadInfoChecker) Check(data interface{}) checker.CheckResult {
	d.init()

	checkResult := d.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("software downloadinfo check failed: %s", checkResult.Reason))
	}

	return checker.NewSuccessResult()
}

// GetDownloadInfoChecker [method] software download info checker
func GetDownloadInfoChecker(field string, required bool) *downloadInfoChecker {
	return &downloadInfoChecker{
		modelChecker: checker.ModelChecker{Field: field, Required: required},
	}
}

type upgradeChecker struct {
	modelChecker checker.ModelChecker
}

func newUpgradeChecker() *upgradeChecker {
	return &upgradeChecker{}
}

func (u *upgradeChecker) init() {
	u.modelChecker.Checker = checker.GetAndChecker(
		checker.GetUniqueListChecker("SerialNumbers",
			checker.GetRegChecker("", `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`, true),
			1, common.MaxNode, true),
		checker.GetStringChoiceChecker("SoftwareName",
			[]string{common.MEFEdge}, true),
	)
}

func (u *upgradeChecker) Check(data interface{}) checker.CheckResult {
	u.init()

	checkResult := u.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("software downloadinfo check failed: %s", checkResult.Reason))
	}

	return checker.NewSuccessResult()
}

type certNameChecker struct {
}

func newCertNameChecker() *certNameChecker {
	return &certNameChecker{}
}

func (c *certNameChecker) Check(certName string) bool {
	certSupportList := []string{common.WsCltName, common.SoftwareCertName, common.ImageCertName, common.NginxCertName}
	for _, certSupport := range certSupportList {
		if certName == certSupport {
			return true
		}
	}
	return false
}
