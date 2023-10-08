// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for testing module
package alarmmanager

import (
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"

	"alarm-manager/pkg/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
)

const (
	normalPageSize = 15
	OverPageSize   = 105
	firstPageNum   = 1
	testSn1        = "testSn1"
	testSn2        = "testSn2"
	groupIdFirst   = 1
)

func TestListAlarmByNodeId(t *testing.T) {
	convey.Convey("normal input for listing alarms by nodeId with Normal Input", t, func() {
		testListAlarmsOrEventsByNodeId(alarms.AlarmType)
	})
	convey.Convey("normal input for listing alarms by nodeId with abnormal input", t, testListAlarmAbNormalInput)
}

func TestListEventsByNodeId(t *testing.T) {
	convey.Convey("normal input for listing events by nodeId with Normal Input", t, func() {
		testListAlarmsOrEventsByNodeId(alarms.EventType)
	})
}

func TestListAlarmsByNodeGroup(t *testing.T) {
	convey.Convey("normal input for listing groupNodes alarms with Normal Input", t, func() {
		testListAlarmsOrEventsOfNodeGroup(alarms.AlarmType)
	})
}

func TestListEventsByNodeGroup(t *testing.T) {
	convey.Convey("normal input for listing groupNodes events with Normal Input", t, func() {
		testListAlarmsOrEventsOfNodeGroup(alarms.EventType)
	})
}

func TestListAlarmsOfCenterNode(t *testing.T) {
	convey.Convey("normal input for listing centerNode alarms with Normal Input", t, func() {
		testListAlarmsOrEventsOfCenter(alarms.AlarmFlag)
	})
}

func TestListEventsOfCenterNode(t *testing.T) {
	convey.Convey("normal input for listing centerNode events with Normal Input", t, func() {
		testListAlarmsOrEventsOfCenter(alarms.EventType)
	})
}

func TestGetAlarm(t *testing.T) {
	convey.Convey("normal input for getting alarm by id with Normal Input", t, func() {
		testGetAlarmOrEventByInfoId(alarms.AlarmType)
	})
	convey.Convey("abnormal input for getting alarm", t, testGetAlarmAbnormalInput)
}

func TestGetEvent(t *testing.T) {
	convey.Convey("normal input for getting alarm by id with Normal Input", t, func() {
		testGetAlarmOrEventByInfoId(alarms.EventType)
	})
}

func testListAlarmsOrEventsByNodeId(queryType string) {
	var reqData = utils.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		Sn:       testSn2,
		IfCenter: "false",
	}
	var (
		resp interface{}
		err  error
	)
	if queryType == alarms.AlarmType {
		resp, err = listAlarms(reqData)
	} else {
		resp, err = listEvents(reqData)
	}
	convey.So(err, convey.ShouldBeNil)
	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	res := true
	if respContent.Data == nil {
		convey.So(respContent.Status, convey.ShouldEqual, common.Success)
		return
	}
	respData, ok := respContent.Data.(utils.ListAlarmsResp)
	if !ok {
		hwlog.RunLog.Error("convert assertion failed")
		return
	}
	for _, alarm := range respData.Records {
		res = res && alarm.Sn == reqData.Sn && (alarm.AlarmType == queryType)
	}
	convey.So(res, convey.ShouldBeTrue)
}

func testListAlarmsOrEventsOfCenter(queryType string) {
	var reqData = utils.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		Sn:       testSn2,
		IfCenter: "true",
	}
	var (
		resp interface{}
		err  error
	)
	if queryType == alarms.AlarmType {
		resp, err = listAlarms(reqData)
	} else {
		resp, err = listEvents(reqData)
	}
	convey.So(err, convey.ShouldBeNil)
	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	res := true
	if respContent.Data == nil {
		convey.So(respContent.Status, convey.ShouldEqual, common.Success)
		return
	}
	convey.So(ok, convey.ShouldBeTrue)
	respData, ok := respContent.Data.(utils.ListAlarmsResp)
	convey.So(ok, convey.ShouldBeTrue)
	for _, alarm := range respData.Records {
		res = res && (alarm.Sn == alarms.CenterSn) && (alarm.AlarmType == queryType)
	}
	convey.So(res, convey.ShouldBeTrue)
}

func testListAlarmsOrEventsOfNodeGroup(queryType string) {
	var reqData = utils.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		GroupId:  groupIdFirst,
		IfCenter: "false",
	}
	var (
		resp interface{}
		err  error
	)
	if queryType == alarms.AlarmType {
		resp, err = listAlarms(reqData)
	} else {
		resp, err = listEvents(reqData)
	}
	convey.So(err, convey.ShouldBeNil)
	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(respContent.Status, convey.ShouldEqual, common.Success)
	res := true
	respData, ok := respContent.Data.(utils.ListAlarmsResp)
	convey.So(ok, convey.ShouldBeTrue)
	for _, alarm := range respData.Records {
		res = res && (groupNodesMap[alarm.Sn]) && (alarm.AlarmType == queryType)
	}
	convey.So(res, convey.ShouldBeTrue)
}

func testGetAlarmOrEventByInfoId(queryType string) {
	var (
		resp interface{}
		err  error
	)
	if queryType == alarms.AlarmType {
		resp, err = getAlarmDetail(DefaultAlarmID)
	} else {
		resp, err = getEventDetail(DefaultEventID)
	}
	convey.So(err, convey.ShouldBeNil)
	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(respContent.Status, convey.ShouldEqual, common.Success)
	alarm, ok := respContent.Data.(*AlarmInfo)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(alarm.AlarmType, convey.ShouldEqual, queryType)

}

func testListAlarmAbNormalInput() {

	inputCase1 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize,
		Sn: testSn1, GroupId: groupIdFirst, IfCenter: "false"}
	listAlarmsWithInput(inputCase1, false, "sn and groupId can't exist at the same time when"+
		" ifCenter is not true", false,
		defaultTestCaseCallback)
	// with IfCenter == true sn and groupId should be ignored
	inputCase2 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, Sn: testSn1,
		GroupId: 1, IfCenter: "true"}
	listAlarmsWithInput(inputCase2, true, "", false, CallbackAllCenterNodes)
	inputCase3 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: OverPageSize,
		Sn: testSn1, IfCenter: "false"}
	listAlarmsWithInput(inputCase3, false, "", true, CallBackStringsContains)
	// with IfCenter == true sn and groupId should be ignored
	inputCase4 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize,
		Sn: testSn1, IfCenter: "true"}
	listAlarmsWithInput(inputCase4, true, "", false, CallbackAllCenterNodes)
	inputCase5 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: "true"}
	listAlarmsWithInput(inputCase5, true, "", false, CallbackAllCenterNodes)
	inputCase6 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: "false"}
	listAlarmsWithInput(inputCase6, true, "", true, CallbackAllAlarms)
	inputCase7 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize}
	listAlarmsWithInput(inputCase7, true, "", true, CallbackAllAlarms)
	inputCase8 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, GroupId: groupIdFirst,
		IfCenter: "true"}
	listAlarmsWithInput(inputCase8, true, "", true, CallbackAllCenterNodes)
	inputCase9 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: OverPageSize, IfCenter: "true"}
	listAlarmsWithInput(inputCase9, false, "", true, defaultTestCaseCallback)
	inputCase10 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: OverPageSize, GroupId: groupIdFirst,
		IfCenter: "false"}
	listAlarmsWithInput(inputCase10, false, "", true, defaultTestCaseCallback)
	inputCase11 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: ""}
	listAlarmsWithInput(inputCase11, true, "", true, CallbackAllAlarms)
	inputCase12 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: ""}
	listAlarmsWithInput(inputCase12, true, "", true, CallbackAllAlarms)
}

func testGetAlarmAbnormalInput() {
	// use an event id[2] to look for alarm
	getAlarmWithInput(DefaultEventID, false)
	getAlarmWithInput(0, false)
}

func getAlarmWithInput(id uint64, expectRes bool) {
	resp, err := getAlarmDetail(id)
	convey.So(err, convey.ShouldBeNil)
	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	if expectRes {
		convey.So(respContent.Status, convey.ShouldEqual, common.Success)
	} else {
		convey.So(respContent.Status, convey.ShouldNotEqual, common.Success)
	}
}

func listAlarmsWithInput(input interface{}, expectRes bool, expectMsg string, ignoredMsg bool,
	callback func(msg common.RespMsg) bool) {
	resp, err := listAlarms(input)
	convey.So(err, convey.ShouldBeNil)
	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	if expectRes {
		convey.So(respContent.Status, convey.ShouldEqual, common.Success)
	} else {
		convey.So(respContent.Status, convey.ShouldNotEqual, common.Success)
	}
	if !ignoredMsg {
		convey.So(respContent.Msg, convey.ShouldEqual, expectMsg)
	}
	if expectRes {
		convey.So(callback(*respContent), convey.ShouldEqual, true)
	}
}

func defaultTestCaseCallback(resp common.RespMsg) bool {
	return true
}

func CallBackStringsContains(resp common.RespMsg) bool {
	return strings.Contains(resp.Msg, "Uint checker Check [PageSize] failed")
}

func CallbackAllAlarms(resp common.RespMsg) bool {
	respData, ok := resp.Data.(utils.ListAlarmsResp)
	if !ok {
		hwlog.RunLog.Error("failed to marshal alarm info")
		return false
	}
	res := true
	for _, alarm := range respData.Records {
		res = res && alarm.AlarmType == alarms.AlarmType
	}
	return res
}

func CallbackAllCenterNodes(resp common.RespMsg) bool {
	// empty list result return nil data
	if resp.Data == nil {
		return true
	}
	respData, ok := resp.Data.(utils.ListAlarmsResp)
	if !ok {
		hwlog.RunLog.Error("failed to marshal alarm info")
		return false
	}
	res := true
	for _, alarm := range respData.Records {
		res = res && alarm.Sn == alarms.CenterSn
	}
	return res
}
