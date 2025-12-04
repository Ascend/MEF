// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package alarmmanager test for alarm_operate.go
package alarmmanager

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

func TestDealAlarmsReq(t *testing.T) {
	convey.Convey("test func dealAlarmsReq success", t, testDealAlarmsReq)
	convey.Convey("test func dealAlarmsReq failed, param convert failed", t, testDealAlarmsReqErrParse)
	convey.Convey("test func dealAlarmsReq failed, param check failed", t, testDealAlarmsReqErrCheck)
	convey.Convey("test func dealAlarmsReq failed, deal failed", t, testDealAlarmsReqErrDeal)
}

func testDealAlarmsReq() {
	bytes, err := json.Marshal(newAlarmsReq(caseCenterAlarm))
	convey.So(err, convey.ShouldBeNil)

	res := dealAlarmsReq(&model.Message{Content: bytes})
	convey.So(res, convey.ShouldBeNil)
}

func testDealAlarmsReqErrParse() {
	res := dealAlarmsReq(&model.Message{})
	convey.So(res, convey.ShouldBeNil)
}

func testDealAlarmsReqErrCheck() {
	alarmsReqs := []requests.AlarmsReq{
		newAlarmsReq(caseErrSn),
		newAlarmsReq(caseErrIp),
		newAlarmsReq(caseErrType),
		newAlarmsReq(caseErrAlarmId),
		newAlarmsReq(caseErrAlarmName),
		newAlarmsReq(caseErrResource),
		newAlarmsReq(caseErrPerceivedSeverity),
		newAlarmsReq(caseErrNotificationType),
		newAlarmsReq(caseErrDetailedInformation),
		newAlarmsReq(caseErrSuggestion),
		newAlarmsReq(caseErrReason),
		newAlarmsReq(caseErrImpact),
	}

	for _, alarmsReq := range alarmsReqs {
		bytes, err := json.Marshal(alarmsReq)
		convey.So(err, convey.ShouldBeNil)

		res := dealAlarmsReq(&model.Message{Content: bytes})
		convey.So(res, convey.ShouldBeNil)
	}

	// test alarm list length checker
	req := requests.AlarmsReq{
		Sn: alarms.CenterSn,
		Ip: testIp,
	}
	for i := 0; i < maxOneNodeAlarmCount+1; i++ {
		req.Alarms = append(req.Alarms, newAlarmReq())
	}

	bytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)

	res := dealAlarmsReq(&model.Message{Content: bytes})
	convey.So(res, convey.ShouldBeNil)
}

func testDealAlarmsReqErrDeal() {
	var p1 = gomonkey.ApplyFuncReturn(time.Parse, nil, test.ErrTest)
	defer p1.Reset()

	bytes, err := json.Marshal(newAlarmsReq(caseCenterAlarm))
	convey.So(err, convey.ShouldBeNil)

	res := dealAlarmsReq(&model.Message{Content: bytes})
	convey.So(res, convey.ShouldBeNil)
}

func TestAlarmReqDealer(t *testing.T) {
	convey.Convey("test AlarmReqDealer method 'deal' success", t, testArdDeal)
	convey.Convey("test AlarmReqDealer method 'deal' failed, getAlarmInfo failed", t, testArdDealErrGetAlarmInfo)

	convey.Convey("test AlarmReqDealer method 'clearAlarm' success", t, testArdClearAlarm)
	convey.Convey("test AlarmReqDealer method 'clearAlarm' failed, get alarm failed", t, testArdClearAlarmErrGetAlarm)
	convey.Convey("test AlarmReqDealer method 'clearAlarm' failed, delete alarm failed",
		t, testArdClearAlarmErrDeleteAlarm)

	// 'addAlarm' 'dealEvent' success case tested by testArdDeal
	convey.Convey("test AlarmReqDealer method 'addAlarm' failed, get alarm count failed", t, testArdAddAlarmErrGetCount)
	convey.Convey("test AlarmReqDealer method 'addAlarm' failed, alarm count is invalid", t, testArdAddAlarmErrCount)
	convey.Convey("test AlarmReqDealer method 'addAlarm' failed, get alarm failed", t, testArdAddAlarmErrGetAlarm)
	convey.Convey("test AlarmReqDealer method 'addAlarm' failed, add alarm failed", t, testArdAddAlarmErrAddAlarm)

	convey.Convey("test AlarmReqDealer method 'dealEvent' failed, get event count failed", t, testArdDealEventErrGetCount)
	convey.Convey("test AlarmReqDealer method 'dealEvent' failed, event count is large", t, testArdDealEventErrCount)
	convey.Convey("test AlarmReqDealer method 'dealEvent' failed, add event failed", t, testArdDealEventErrAddEvent)
}

func testArdDeal() {
	// test case: add alarm
	alarmReq := newAlarmsReq(caseCenterAlarm).Alarms[0]
	dealer := GetAlarmReqDealer(&alarmReq, alarms.CenterSn, testIp)
	err := dealer.deal()
	convey.So(err, convey.ShouldBeNil)

	// test case: clear alarm
	alarmReq.NotificationType = alarms.ClearFlag
	dealer = GetAlarmReqDealer(&alarmReq, alarms.CenterSn, testIp)
	err = dealer.deal()
	convey.So(err, convey.ShouldBeNil)

	// test case: add event
	alarmReq.Type = alarms.EventType
	dealer = GetAlarmReqDealer(&alarmReq, alarms.CenterSn, testIp)
	err = dealer.deal()
	convey.So(err, convey.ShouldBeNil)
}

func testArdDealErrGetAlarmInfo() {
	var p1 = gomonkey.ApplyFuncReturn(time.Parse, time.Time{}, test.ErrTest)
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	err := dealer.deal()
	expErr := errors.New("get alarm info failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdClearAlarm() {
	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	alarmInfo, err := dealer.getAlarmInfo()
	if err != nil {
		panic(err)
	}
	dealer.alarmInfo = alarmInfo
	err = dealer.clearAlarm()
	convey.So(err, convey.ShouldBeNil)
}

func testArdClearAlarmErrGetAlarm() {
	var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Find", &gorm.DB{Error: test.ErrTest})
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	err := dealer.clearAlarm()
	expErr := errors.New("get alarm info from db failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdClearAlarmErrDeleteAlarm() {
	var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Delete", &gorm.DB{Error: test.ErrTest})
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	alarmInfo, err := dealer.getAlarmInfo()
	if err != nil {
		panic(err)
	}
	dealer.alarmInfo = alarmInfo
	err = dealer.clearAlarm()
	convey.So(err, convey.ShouldBeNil)

	var p2 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "getAlarmInfo",
		func(string, string) ([]AlarmInfo, error) {
			return []AlarmInfo{{Ip: testIp}}, nil
		})
	defer p2.Reset()
	err = dealer.clearAlarm()
	expErr := errors.New("delete alarm data failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdAddAlarmErrGetCount() {
	var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Count", &gorm.DB{Error: test.ErrTest})
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	err := dealer.addAlarm()
	expErr := errors.New("get node alarm count failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdAddAlarmErrCount() {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "getNodeAlarmCount",
		func(string, string) (int, error) {
			return testNumHundred, nil
		})
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	err := dealer.addAlarm()
	expErr := errors.New("node's alarm count have reached the max counts")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdAddAlarmErrGetAlarm() {
	var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Find", &gorm.DB{Error: test.ErrTest})
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	err := dealer.addAlarm()
	expErr := errors.New("get alarm info from db failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdAddAlarmErrAddAlarm() {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "getAlarmInfo",
		func(string, string) ([]AlarmInfo, error) {
			return []AlarmInfo{}, nil
		})
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	alarmInfo, err := dealer.getAlarmInfo()
	if err != nil {
		panic(err)
	}
	dealer.alarmInfo = alarmInfo
	err = dealer.addAlarm()
	convey.So(err, convey.ShouldBeNil)

	var p2 = gomonkey.ApplyFuncReturn(rand.Int, nil, test.ErrTest)
	defer p2.Reset()
	err = dealer.addAlarm()
	expErr := errors.New("add alarm into db failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdDealEventErrGetCount() {
	var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Count", &gorm.DB{Error: test.ErrTest})
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	err := dealer.dealEvent()
	expErr := errors.New("get node event count failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdDealEventErrCount() {
	outputs1 := []gomonkey.OutputCell{
		{Values: gomonkey.Params{&gorm.DB{Error: test.ErrTest}}},
		{Values: gomonkey.Params{&gorm.DB{}}, Times: 2},
	}
	outputs2 := []gomonkey.OutputCell{
		{Values: gomonkey.Params{&gorm.DB{Error: test.ErrTest}}},
		{Values: gomonkey.Params{&gorm.DB{}}},
	}
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "getNodeEventCount",
		func(string, string) (int, error) {
			return testNumHundred, nil
		}).
		ApplyMethodSeq(&gorm.DB{}, "Find", outputs1).
		ApplyMethodSeq(&gorm.DB{}, "Delete", outputs2)
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)

	err := dealer.dealEvent()
	expErr := errors.New("get node oldest event failed")
	convey.So(err, convey.ShouldResemble, expErr)

	err = dealer.dealEvent()
	expErr = errors.New("delete oldest event failed")
	convey.So(err, convey.ShouldResemble, expErr)

	err = dealer.dealEvent()
	dealer.alarmInfo = nil
	expErr = errors.New("alarm info is nil")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testArdDealEventErrAddEvent() {
	var p1 = gomonkey.ApplyFuncReturn(rand.Int, nil, test.ErrTest)
	defer p1.Reset()

	dealer := GetAlarmReqDealer(&newAlarmsReq(caseCenterAlarm).Alarms[0], alarms.CenterSn, testIp)
	alarmInfo, err := dealer.getAlarmInfo()
	if err != nil {
		panic(err)
	}
	dealer.alarmInfo = alarmInfo
	err = dealer.dealEvent()
	expErr := errors.New("add new event into db failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func TestDealNodeClearReq(t *testing.T) {
	convey.Convey("test func dealNodeClearReq success", t, testDealNodeClearReq)
	convey.Convey("test func dealNodeClearReq failed, param convert failed", t, testDealNodeClearReqErrParse)
	convey.Convey("test func dealNodeClearReq failed, delete failed", t, testDealNodeClearReqErrDelete)
}

func testDealNodeClearReq() {
	req := requests.ClearNodeAlarmReq{Sn: alarms.CenterSn}
	bytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	res := dealNodeClearReq(&model.Message{Content: bytes})
	convey.So(res, convey.ShouldEqual, common.OK)
}

func testDealNodeClearReqErrParse() {
	res := dealNodeClearReq(&model.Message{})
	convey.So(res, convey.ShouldEqual, common.FAIL)
}

func testDealNodeClearReqErrDelete() {
	var p1 = gomonkey.ApplyMethodReturn(&gorm.DB{}, "Delete", &gorm.DB{Error: test.ErrTest})
	defer p1.Reset()

	req := requests.ClearNodeAlarmReq{Sn: alarms.CenterSn}
	bytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	res := dealNodeClearReq(&model.Message{Content: bytes})
	convey.So(res, convey.ShouldEqual, common.FAIL)
}
