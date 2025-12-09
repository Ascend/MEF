// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package alarmmanager for package main test
package alarmmanager

import (
	"crypto/rand"
	"encoding/json"
	"math"
	"math/big"
	rand2 "math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	testNumHundred      = 100
	testNumThreeHundred = 300
	testNumSixHundred   = 600

	NodeNums        = 4
	TypesOfSeverity = 3
	Possibility     = 10
	HalfPossibility = 5
)

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcBaseWithDb := &test.TcBaseWithDb{
		Tables: append(tables, &AlarmInfo{}),
	}

	resp := common.RespMsg{
		Status: common.Success,
		Data:   []string{testEdgeSn},
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb).
		ApplyFuncReturn(modulemgr.SendSyncMessage, &model.Message{Content: bytes}, nil).
		ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "GetWithTimeout", bytes, nil)
	test.RunWithPatches(tcBaseWithDb, m, patches)
}

const (
	caseCenterAlarm            = "centerAlarm"
	caseCenterEvent            = "centerEvent"
	caseEdgeAlarm              = "edgeAlarm"
	caseEdgeEvent              = "edgeEvent"
	caseErrSn                  = "errSn"
	caseErrIp                  = "errIp"
	caseErrType                = "errType"
	caseErrAlarmId             = "errAlarmId"
	caseErrAlarmName           = "errAlarmName"
	caseErrResource            = "errResource"
	caseErrPerceivedSeverity   = "errPerceivedSeverity"
	caseErrNotificationType    = "errNotificationType"
	caseErrDetailedInformation = "errDetailedInformation"
	caseErrSuggestion          = "errSuggestion"
	caseErrReason              = "errReason"
	caseErrImpact              = "errImpact"

	testIp      = "10.10.10.10"
	testEdgeSn  = "testEdgeSn"
	testAlarmId = "0x01000003"
)

func newAlarmReq() requests.AlarmReq {
	return requests.AlarmReq{
		Type:                alarms.AlarmType,
		AlarmId:             testAlarmId,
		AlarmName:           "Image Repository Cert Abnormal",
		Resource:            "cert",
		PerceivedSeverity:   alarms.MajorSeverity,
		Timestamp:           "2024-01-01T00:00:00+08:00",
		NotificationType:    alarms.AlarmFlag,
		DetailedInformation: "test alarm detailed information",
		Suggestion:          "test alarm suggestion",
		Reason:              "test alarm reason",
		Impact:              "test alarm impact",
	}
}

func newAlarmsReq(caseType string) requests.AlarmsReq {
	alarmReq := newAlarmReq()
	req := requests.AlarmsReq{
		Alarms: []requests.AlarmReq{newAlarmReq()},
		Sn:     alarms.CenterSn,
		Ip:     testIp,
	}

	switch caseType {
	case caseCenterAlarm:
		return req

	case caseCenterEvent:
		centerEvent := req
		eventReq := alarmReq
		eventReq.Type = alarms.EventType
		centerEvent.Alarms = []requests.AlarmReq{eventReq}
		return centerEvent

	case caseEdgeAlarm:
		edgeAlarm := req
		edgeAlarm.Sn = testEdgeSn
		return edgeAlarm

	case caseEdgeEvent:
		edgeEvent := req
		edgeEvent.Sn = testEdgeSn
		eventReq := alarmReq
		eventReq.Type = alarms.EventType
		edgeEvent.Alarms = []requests.AlarmReq{eventReq}
		return edgeEvent

	case caseErrSn:
		errSn := req
		errSn.Sn = "error sn"
		return errSn

	case caseErrIp:
		errIp := req
		errIp.Ip = "error ip"
		return errIp

	case caseErrType:
		errType := req
		errAlarmReq := alarmReq
		errAlarmReq.Type = "error type"
		errType.Alarms = []requests.AlarmReq{errAlarmReq}
		return errType

	case caseErrAlarmId:
		errAlarmId := req
		errAlarmReq := alarmReq
		errAlarmReq.AlarmId = "error alarm id"
		errAlarmId.Alarms = []requests.AlarmReq{errAlarmReq}
		return errAlarmId

	case caseErrAlarmName:
		errAlarmName := req
		errAlarmReq := alarmReq
		errAlarmReq.AlarmName = generateRandString(testNumHundred)
		errAlarmName.Alarms = []requests.AlarmReq{errAlarmReq}
		return errAlarmName

	case caseErrResource:
		errResource := req
		errAlarmReq := alarmReq
		errAlarmReq.Resource = generateRandString(testNumThreeHundred)
		errResource.Alarms = []requests.AlarmReq{errAlarmReq}
		return errResource

	case caseErrPerceivedSeverity:
		errPerceivedSeverity := req
		errAlarmReq := alarmReq
		errAlarmReq.PerceivedSeverity = "error severity"
		errPerceivedSeverity.Alarms = []requests.AlarmReq{errAlarmReq}
		return errPerceivedSeverity

	case caseErrNotificationType:
		errNotificationType := req
		errAlarmReq := alarmReq
		errAlarmReq.NotificationType = "error notification type"
		errNotificationType.Alarms = []requests.AlarmReq{errAlarmReq}
		return errNotificationType

	case caseErrDetailedInformation:
		errDetailedInformation := req
		errAlarmReq := alarmReq
		errAlarmReq.DetailedInformation = generateRandString(testNumThreeHundred)
		errDetailedInformation.Alarms = []requests.AlarmReq{errAlarmReq}
		return errDetailedInformation

	case caseErrSuggestion:
		errSuggestion := req
		errAlarmReq := alarmReq
		errAlarmReq.Suggestion = generateRandString(testNumSixHundred)
		errSuggestion.Alarms = []requests.AlarmReq{errAlarmReq}
		return errSuggestion

	case caseErrReason:
		errReason := req
		errAlarmReq := alarmReq
		errAlarmReq.Reason = generateRandString(testNumThreeHundred)
		errReason.Alarms = []requests.AlarmReq{errAlarmReq}
		return errReason

	case caseErrImpact:
		errImpact := req
		errAlarmReq := alarmReq
		errAlarmReq.Impact = generateRandString(testNumThreeHundred)
		errImpact.Alarms = []requests.AlarmReq{errAlarmReq}
		return errImpact

	default:
		return requests.AlarmsReq{}
	}
}

func newAlarmInfo(caseType string) AlarmInfo {
	alarmsReq := newAlarmsReq(caseType)
	alarmReq := alarmsReq.Alarms[0]
	dealer := GetAlarmReqDealer(&alarmReq, alarmsReq.Sn, alarmsReq.Ip)
	alarmInfo, err := dealer.getAlarmInfo()
	if err != nil {
		panic(err)
	}
	return *alarmInfo
}

func randIntn(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return -1, err
	}
	randNum := int((*n).Int64())
	return randNum, nil
}

func generateRandString(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	var res []rune
	if length > 0 && length <= testNumSixHundred {
		res = make([]rune, length)
	}
	for i := range res {
		res[i] = letterRunes[rand2.Intn(len(letterRunes))]
	}
	return string(res)
}

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
