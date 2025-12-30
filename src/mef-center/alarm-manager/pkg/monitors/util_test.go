// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package monitors test for util.go
package monitors

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

func TestGetAlarmMonitorList(t *testing.T) {
	patches := gomonkey.ApplyMethodReturn(&common.DbMgr{}, "GetAlarmConfig", 1, nil)
	defer patches.Reset()
	convey.Convey("test func GetAlarmMonitorList success", t, testGetAlarmMonitorList)
	convey.Convey("test func GetAlarmMonitorList failed", t, testGetAlarmMonitorListErr)
}

func testGetAlarmMonitorList() {
	alarmMonitor := GetAlarmMonitorList("./")
	convey.So(alarmMonitor, convey.ShouldResemble, []AlarmMonitor{certTask})
}

func testGetAlarmMonitorListErr() {
	var p1 = gomonkey.ApplyMethodReturn(&common.DbMgr{}, "GetAlarmConfig", 1, test.ErrTest)
	defer p1.Reset()
	alarmMonitor := GetAlarmMonitorList("./")
	convey.So(alarmMonitor, convey.ShouldResemble, []AlarmMonitor(nil))

	var p2 = gomonkey.ApplyMethodSeq(&common.DbMgr{}, "GetAlarmConfig", []gomonkey.OutputCell{
		{Values: gomonkey.Params{1, nil}},
		{Values: gomonkey.Params{1, test.ErrTest}, Times: 5},
	})
	defer p2.Reset()
	alarmMonitor = GetAlarmMonitorList("./")
	convey.So(alarmMonitor, convey.ShouldResemble, []AlarmMonitor(nil))
}

var testAlarms []*requests.AlarmReq

func TestSendAlarms(t *testing.T) {
	alarm := &requests.AlarmReq{
		Type:                "test alarm type",
		AlarmId:             "test alarm id",
		AlarmName:           "test alarm name",
		Resource:            "test alarm resource",
		PerceivedSeverity:   "test alarm perceived severity",
		Timestamp:           "test alarm time stamp",
		NotificationType:    "test alarm notification type",
		DetailedInformation: "test alarm detailed information",
		Suggestion:          "test alarm suggestion",
		Reason:              "test alarm reason",
		Impact:              "test alarm impact",
	}
	testAlarms = []*requests.AlarmReq{alarm}

	convey.Convey("test func SendAlarms failed, length of alarms is zero", t, testSendAlarmsErrLen)
	convey.Convey("test func SendAlarms failed, nil alarm", t, testSendAlarmsErrAlarm)
	convey.Convey("test func SendAlarms failed, get host ip failed", t, testSendAlarmsErrGetHostIP)
	convey.Convey("test func SendAlarms failed, new message failed", t, testSendAlarmsErrNewMsg)
	convey.Convey("test func SendAlarms failed, new message failed", t, testSendAlarmsErrFillContent)
	convey.Convey("test func SendAlarms failed, send async message failed", t, testSendAlarmsErrSendMsg)
}

func testSendAlarmsErrLen() {
	err := SendAlarms()
	convey.So(err, convey.ShouldResemble, errors.New("alarm is required"))
}

func testSendAlarmsErrAlarm() {
	alarms := []*requests.AlarmReq{nil}
	err := SendAlarms(alarms...)
	convey.So(err, convey.ShouldResemble, errors.New("alarm req can not be nil pointer"))
}

func testSendAlarmsErrGetHostIP() {
	var p1 = gomonkey.ApplyFuncReturn(common.GetHostIP, "", test.ErrTest)
	defer p1.Reset()
	err := SendAlarms(testAlarms...)
	convey.So(err, convey.ShouldResemble, errors.New("get host ip failed"))
}

func testSendAlarmsErrNewMsg() {
	var p1 = gomonkey.ApplyFuncReturn(model.NewMessage, nil, test.ErrTest)
	defer p1.Reset()
	err := SendAlarms(testAlarms...)
	convey.So(err, convey.ShouldResemble, errors.New("new alarm msg failed"))
}

func testSendAlarmsErrFillContent() {
	var p1 = gomonkey.ApplyMethodReturn(&model.Message{}, "FillContent", test.ErrTest)
	defer p1.Reset()
	err := SendAlarms(testAlarms...)
	convey.So(err, convey.ShouldResemble, errors.New("fill content failed"))
}

func testSendAlarmsErrSendMsg() {
	var p1 = gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, test.ErrTest)
	defer p1.Reset()
	err := SendAlarms(testAlarms...)
	convey.So(err, convey.ShouldResemble, errors.New("send async message failed"))
}

func TestUpdateImportedCertsInfo(t *testing.T) {
	var patches = gomonkey.ApplyMethodReturn(&requests.ReqCertParams{}, "GetImportedCertsInfo",
		"test cert info", nil).
		ApplyFuncReturn(json.Unmarshal, nil)
	defer patches.Reset()
	convey.Convey("test func updateImportedCertsInfo success", t, testUpdateCertsInfo)
	convey.Convey("test func updateImportedCertsInfo failed, get info failed", t, testUpdateCertsInfoErrGetInfo)
	convey.Convey("test func updateImportedCertsInfo failed, unmarshal failed", t, testUpdateCertsInfoErrUnmarshal)
}

func testUpdateCertsInfo() {
	err := updateImportedCertsInfo()
	convey.So(err, convey.ShouldBeNil)
}

func testUpdateCertsInfoErrGetInfo() {
	var p1 = gomonkey.ApplyMethodReturn(&requests.ReqCertParams{}, "GetImportedCertsInfo",
		"", test.ErrTest)
	defer p1.Reset()
	err := updateImportedCertsInfo()
	convey.So(err, convey.ShouldResemble, errors.New("get imported certs info from cert-manager failed"))
}

func testUpdateCertsInfoErrUnmarshal() {
	var p1 = gomonkey.ApplyFuncReturn(json.Unmarshal, test.ErrTest)
	defer p1.Reset()
	err := updateImportedCertsInfo()
	convey.So(err, convey.ShouldResemble, errors.New("unmarshal imported certs info failed"))
}
