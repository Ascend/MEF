package parachecker

import (
	"fmt"

	"huawei.com/mindxedge/base/common/checker/checker"
)

type EdgeUpgradeInfoChecker struct {
	modelChecker checker.ModelChecker
}

func (ec *EdgeUpgradeInfoChecker) init() {
	ec.modelChecker.Required = true
	ec.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("NodeIDs", nameReg, true),
		checker.GetRegChecker("SoftWareName", descriptionReg, true),
	)
}

// Check [method] for create app checker
func (ec *EdgeUpgradeInfoChecker) Check(data interface{}) checker.CheckResult {
	ec.init()
	checkResult := ec.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("check edge upgrade info failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

type DownloadInfoChecker struct {
	modelChecker checker.ModelChecker
}

func (ec *DownloadInfoChecker) init() {
	ec.modelChecker.Required = true
	ec.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Url", nameReg, true),
		checker.GetRegChecker("UserName", descriptionReg, true),
		checker.GetRegChecker("Password", descriptionReg, true),
	)
}

func (ec *DownloadInfoChecker) Check(data interface{}) checker.CheckResult {
	ec.init()
	checkResult := ec.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("check download info failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
