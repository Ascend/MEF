// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for alarm-manager module checker
package alarmmanager

import (
	"fmt"
	"math"

	"huawei.com/mindx/common/checker"

	"alarm-manager/pkg/utils"

	"huawei.com/mindxedge/base/common"
)

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
			checker.GetStringChoiceChecker("Sn", []string{""}, true),
		),
		checker.GetUintChecker("GroupId", 0, math.MaxInt64, true),
		checker.GetStringChoiceChecker("IfCenter", []string{"true", "false", ""}, true),
	)
}

// Check checking all params
func (alc *AlarmListerChecker) Check(data utils.ListAlarmOrEventReq) checker.CheckResult {
	alc.init()

	if data.IfCenter != trueStr {
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
	return checker.GetUintChecker("", 1, math.MaxInt64, true)
}
