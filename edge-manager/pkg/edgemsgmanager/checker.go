package edgemsgmanager

import (
	"fmt"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker/checker"
)

func newDownloadChecker() *downloadChecker {
	return &downloadChecker{}
}

type downloadChecker struct {
	modelChecker checker.ModelChecker
}

func (d *downloadChecker) init() {
	d.modelChecker.Checker = checker.GetAndChecker(
		checker.GetUniqueListChecker("SerialNumbers",
			checker.GetRegChecker("", `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`, true),
			1, 1024, true),
		checker.GetStringChoiceChecker("SoftwareName",
			[]string{common.EdgeInstaller, common.EdgeCore, common.DevicePlugin}, true),
		GetDownloadInfoChecker("DownloadInfo", true),
	)
}

func (d *downloadChecker) Check(data interface{}) checker.CheckResult {
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
	d.modelChecker.Checker = checker.GetAndChecker(
		checker.GetHttpsUrlChecker("Package", true),
		checker.GetHttpsUrlChecker("SignFile", false),
		checker.GetHttpsUrlChecker("CrlFile", false),
		checker.GetRegChecker("UserName", "^[a-zA-Z][a-zA-Z0-9-_]{1,64}[a-zA-Z0-9]$", false),
		checker.GetExistChecker("Password"),
	)
}

func (d *downloadInfoChecker) Check(data interface{}) checker.CheckResult {
	d.init()

	checkResult := d.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("software downloadinfo check failed: %s", checkResult.Reason))
	}

	return checker.NewSuccessResult()
}

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
			[]string{common.EdgeInstaller, common.EdgeCore, common.DevicePlugin}, true),
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
