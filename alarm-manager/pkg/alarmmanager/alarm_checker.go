// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package alarmmanager

import (
	"fmt"

	"huawei.com/mindx/common/checker"
)

// DealAlarmChecker is the checker for add alarm msg
type DealAlarmChecker struct {
	checker checker.ModelChecker
}

// NewDealAlarmChecker is the func to create a DealAlarmChecker
func NewDealAlarmChecker() *DealAlarmChecker {
	return &DealAlarmChecker{}
}

func (dac *DealAlarmChecker) init() {
	alarmIdReg := "^0x0[0-9a-f]{7}$"
	alarmNameReg := "[a-z0-9A-Z- _]{1,64}"
	resourceReg := "[a-z0-9A-Z- _]{1,256}"
	const (
		detailedInformationLength = 256
		suggestionLength          = 512
		reasonLength              = 256
		impactLength              = 256
		minLength                 = 0
	)
	dac.checker.Checker = checker.GetAndChecker(
		checker.GetStringChoiceChecker("Type", []string{AlarmType, EventType}, true),
		checker.GetRegChecker("AlarmId", alarmIdReg, true),
		checker.GetRegChecker("AlarmName", alarmNameReg, false),
		checker.GetRegChecker("Resource", resourceReg, true),
		checker.GetStringChoiceChecker("PerceivedSeverity", []string{MajorSeverity, MinorSeverity, CriticalSeverity},
			true),
		checker.GetStringChoiceChecker("NotificationType", []string{ClearFlag, AlarmFlag}, false),
		checker.GetStringLengthChecker("DetailedInformationLength", minLength, detailedInformationLength, false),
		checker.GetStringLengthChecker("Suggestion", minLength, suggestionLength, false),
		checker.GetStringLengthChecker("Reason", minLength, reasonLength, false),
		checker.GetStringLengthChecker("Impact", minLength, impactLength, false),
	)
}

// Check is the main func to start the check for DealAlarmChecker
func (dac *DealAlarmChecker) Check(data interface{}) checker.CheckResult {
	dac.init()
	checkResult := dac.checker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("deal alarmStaticInfo check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
