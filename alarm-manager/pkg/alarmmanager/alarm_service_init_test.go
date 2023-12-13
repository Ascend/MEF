// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for init packages testing
package alarmmanager

import (
	"math"
	"strconv"
	"time"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common/alarms"
)

var (
	groupNodesMap = map[string]bool{testSn1: true}
	// ensure testSn is in db
	testSns        = []string{testSn1, testSn2, ""}
	defaultAlarmID uint64
	defaultEventID uint64
)

const (
	NodeNums        = 4
	TypesOfSeverity = 3
	Possibility     = 10
	HalfPossibility = 5
)

func randSetOneAlarm(alarm *AlarmInfo) {
	severities := []string{"MINOR", "MAJOR", "CRITICAL"}
	defaultInfo := "ALARM DEFAULT INFO"
	defaultSuggest := "ALARM DEFAULT SUGGESTION"
	defaultReason := "ALARM DEFAULT Reason"
	defaultImpact := "ALARM DEFAULT Impact"
	defaultAlarmName := "ALARM DEFAULT NAME"
	defaultAlarmResource := "ALARM DEFAULT RESOURCE"
	randType := alarms.AlarmType
	var randNum, randNodeId, randTypeSe int
	randNodeId, err1 := randIntn(NodeNums)
	randTypeSe, err2 := randIntn(TypesOfSeverity - 1)
	randNum, err3 := randIntn(Possibility)
	if err1 != nil || err2 != nil || err3 != nil {
		hwlog.RunLog.Error("failed to generate random id")
		return
	}
	if randNum < HalfPossibility {
		randType = alarms.EventType
	}
	alarm.AlarmType = randType
	alarm.CreatedAt = time.Now()
	alarm.SerialNumber = strconv.Itoa(randNodeId)
	alarmId, err := randIntn(math.MaxUint32)
	if err != nil {
		hwlog.RunLog.Error("failed to generate alarm id")
		return
	}
	alarm.AlarmId = strconv.Itoa(alarmId)
	alarm.AlarmName = defaultAlarmName
	typeSe := randTypeSe
	if typeSe >= len(severities) {
		return
	}
	alarm.PerceivedSeverity = severities[typeSe]
	alarm.DetailedInformation = defaultInfo
	alarm.Suggestion = defaultSuggest
	alarm.Reason = defaultReason
	alarm.Impact = defaultImpact
	alarm.Resource = defaultAlarmResource
}
