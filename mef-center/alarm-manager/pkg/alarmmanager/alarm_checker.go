// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package alarmmanager

import (
	"fmt"
	"math"

	"huawei.com/mindx/common/checker"

	"alarm-manager/pkg/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
)

// DealAlarmsReqChecker is the checker for dealing alarms request
type DealAlarmsReqChecker struct {
	checker checker.ModelChecker
}

// NewDealAlarmsReqChecker is the func to create DealAlarmsReqChecker
func NewDealAlarmsReqChecker() *DealAlarmsReqChecker {
	return &DealAlarmsReqChecker{}
}

func (dac *DealAlarmsReqChecker) init() {
	dac.checker.Checker = checker.GetAndChecker(
		checker.GetOrChecker(
			checker.GetSnChecker("Sn", true),
			checker.GetStringChoiceChecker("Sn", []string{alarms.CenterSn}, true),
		),
		checker.GetIpV4Checker("Ip", true),
		checker.GetListChecker("Alarms", NewDealAlarmChecker(), 0, maxOneNodeAlarmCount, true),
	)
}

// Check is the main func to start the check for DealAlarmsReqChecker
func (dac *DealAlarmsReqChecker) Check(data interface{}) checker.CheckResult {
	dac.init()
	checkResult := dac.checker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("deal alarms req check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

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
	alarmNameReg := "^[a-z0-9A-Z- _]{0,64}$"
	resourceReg := "^[a-z0-9A-Z- _]{0,256}$"
	const (
		detailedInformationLength = 256
		suggestionLength          = 512
		reasonLength              = 256
		impactLength              = 256
		minLength                 = 0
	)
	dac.checker.Checker = checker.GetAndChecker(
		checker.GetStringChoiceChecker("Type", []string{alarms.AlarmType, alarms.EventType}, true),
		checker.GetRegChecker("AlarmId", alarmIdReg, true),
		checker.GetRegChecker("AlarmName", alarmNameReg, true),
		checker.GetRegChecker("Resource", resourceReg, true),
		checker.GetStringChoiceChecker("PerceivedSeverity",
			[]string{alarms.MajorSeverity, alarms.MinorSeverity, alarms.CriticalSeverity, alarms.OkSeverity}, true),
		checker.GetStringChoiceChecker("NotificationType",
			[]string{alarms.ClearFlag, alarms.AlarmFlag, alarms.EventFlag}, true),
		checker.GetStringLengthChecker("DetailedInformation", minLength, detailedInformationLength, true),
		checker.GetStringLengthChecker("Suggestion", minLength, suggestionLength, true),
		checker.GetStringLengthChecker("Reason", minLength, reasonLength, true),
		checker.GetStringLengthChecker("Impact", minLength, impactLength, true),
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

// AlarmListerChecker checks for listing alarms and events
type AlarmListerChecker struct {
	modelChecker checker.ModelChecker
}

// NewAlarmListerChecker gen a new AlarmListerChecker
func NewAlarmListerChecker() *AlarmListerChecker {
	return &AlarmListerChecker{}
}

func (alc *AlarmListerChecker) init() {
	alc.modelChecker.Required = true

	alc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetUintChecker("PageNum", common.DefaultPage, math.MaxInt32, true),
		checker.GetUintChecker("PageSize", common.DefaultMinPageSize, common.DefaultMaxPageSize, true),
		checker.GetOrChecker(
			checker.GetSnChecker("Sn", true),
			checker.GetStringChoiceChecker("Sn", []string{alarms.CenterSn}, true),
		),
		checker.GetUintChecker("GroupId", 0, math.MaxUint32, true),
		checker.GetStringChoiceChecker("IfCenter", []string{utils.TrueStr, utils.FalseStr, ""}, true),
	)
}

// Check checking all params
func (alc *AlarmListerChecker) Check(data utils.ListAlarmOrEventReq) checker.CheckResult {
	alc.init()

	if data.IfCenter != utils.TrueStr {
		if data.Sn != "" && data.GroupId != 0 {
			return checker.NewFailedResult("sn and groupId can't exist at the same " +
				"time when ifCenter is not true")
		}
	}

	checkResult := alc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("alarm lister checker failed: %v", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

// NewGetAlarmChecker gen new checker
func NewGetAlarmChecker() *checker.UintChecker {
	return checker.GetUintChecker("", 1, math.MaxUint32, true)
}
