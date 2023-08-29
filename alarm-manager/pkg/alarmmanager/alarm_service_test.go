// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for testing module
package alarmmanager

import (
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"

	"alarm-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

const (
	normalPageSize = 15
	OverPageSize   = 105
	UnderPageSize  = 0
	firstPageNum   = 1
	nodeIdZero     = 0
	nodeIdFirst    = 1
	nodeIdSecond   = 2
	groupIdFirst   = 1
)

func TestListAlarmByNodeId(t *testing.T) {
	convey.Convey("normal input for listing alarms by nodeId with Normal Input", t, func() {
		testListAlarmsOrEventsByNodeId(AlarmType)
	})
	convey.Convey("normal input for listing alarms by nodeId with abnormal input", t, testListAlarmAbNormalInput)
}

func TestListEventsByNodeId(t *testing.T) {
	convey.Convey("normal input for listing events by nodeId with Normal Input", t, func() {
		testListAlarmsOrEventsByNodeId(EventType)
	})
}

func TestListAlarmsByNodeGroup(t *testing.T) {
	convey.Convey("normal input for listing groupNodes alarms with Normal Input", t, func() {
		testListAlarmsOrEventsOfNodeGroup(AlarmType)
	})
}

func TestListEventsByNodeGroup(t *testing.T) {
	convey.Convey("normal input for listing groupNodes events with Normal Input", t, func() {
		testListAlarmsOrEventsOfNodeGroup(EventType)
	})
}

func TestListAlarmsOfCenterNode(t *testing.T) {
	convey.Convey("normal input for listing centerNode alarms with Normal Input", t, func() {
		testListAlarmsOrEventsOfCenter(AlarmFlag)
	})
}

func TestListEventsOfCenterNode(t *testing.T) {
	convey.Convey("normal input for listing centerNode events with Normal Input", t, func() {
		testListAlarmsOrEventsOfCenter(EventType)
	})
}

func TestGetAlarm(t *testing.T) {
	convey.Convey("normal input for getting alarm by id with Normal Input", t, func() {
		testGetAlarmOrEventByInfoId(AlarmType)
	})
	convey.Convey("abnormal input for getting alarm", t, testGetAlarmAbnormalInput)
}

func TestGetEvent(t *testing.T) {
	convey.Convey("normal input for getting alarm by id with Normal Input", t, func() {
		testGetAlarmOrEventByInfoId(EventType)
	})
}

func testListAlarmsOrEventsByNodeId(queryType string) {
	var reqData = types.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		NodeId:   nodeIdSecond,
		IfCenter: "false",
	}
	var resp common.RespMsg
	if queryType == AlarmType {
		resp = listAlarms(reqData)
	} else {
		resp = listEvents(reqData)
	}
	res := true
	if resp.Data == nil {
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		return
	}
	respMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		hwlog.RunLog.Error("convert assertion failed")
		return
	}
	respData, ok := respMap[respDataKey].(map[uint64]types.AlarmBriefInfo)
	if !ok {
		hwlog.RunLog.Error("convert assertion failed")
		return
	}
	for _, alarm := range respData {
		res = res && alarm.NodeID == reqData.NodeId && (alarm.AlarmType == queryType)
	}
	convey.So(res, convey.ShouldBeTrue)
}

func testListAlarmsOrEventsOfCenter(queryType string) {
	var reqData = types.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		NodeId:   nodeIdSecond,
		IfCenter: "true",
	}
	var resp common.RespMsg
	if queryType == AlarmType {
		resp = listAlarms(reqData)
	} else {
		resp = listEvents(reqData)
	}
	res := true
	if resp.Data == nil {
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		return
	}
	respMap, ok := resp.Data.(map[string]interface{})
	convey.So(ok, convey.ShouldBeTrue)
	respData, ok := respMap[respDataKey].(map[uint64]types.AlarmBriefInfo)
	convey.So(ok, convey.ShouldBeTrue)
	for _, alarm := range respData {
		res = res && (uint64(alarm.NodeID) == nodeIdZero) && (alarm.AlarmType == queryType)
	}
	convey.So(res, convey.ShouldBeTrue)
}

func testListAlarmsOrEventsOfNodeGroup(queryType string) {
	var reqData = types.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		GroupId:  groupIdFirst,
		IfCenter: "false",
	}
	var resp common.RespMsg
	if queryType == AlarmType {
		resp = listAlarms(reqData)
	} else {
		resp = listEvents(reqData)
	}
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	res := true
	respMap, ok := resp.Data.(map[string]interface{})
	convey.So(ok, convey.ShouldBeTrue)
	respData, ok := respMap[respDataKey].(map[uint64]types.AlarmBriefInfo)
	convey.So(ok, convey.ShouldBeTrue)
	for _, alarm := range respData {
		res = res && (groupNodesMap[alarm.NodeID]) && (alarm.AlarmType == queryType)
	}
	convey.So(res, convey.ShouldBeTrue)
}

func testGetAlarmOrEventByInfoId(queryType string) {
	var resp common.RespMsg
	if queryType == AlarmType {
		resp = getAlarmDetail(DefaultAlarmID)
	} else {
		resp = getEventDetail(DefaultEventID)
	}
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	alarms, ok := resp.Data.(map[uint64]AlarmInfo)
	convey.So(ok, convey.ShouldBeTrue)
	for _, alarm := range alarms {
		convey.So(alarm.AlarmType, convey.ShouldEqual, queryType)
	}
}

func testListAlarmAbNormalInput() {

	inputCase1 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize,
		NodeId: nodeIdFirst, GroupId: groupIdFirst, IfCenter: "false"}
	listAlarmsWithInput(inputCase1, false, "nodeId and groupId can't exist at the same time "+
		"when ifCenter is not true", false,
		defaultTestCaseCallback)
	// with IfCenter == true nodeid and groupId should be ignored
	inputCase2 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, NodeId: nodeIdFirst,
		GroupId: 1, IfCenter: "true"}
	listAlarmsWithInput(inputCase2, true, "", false, CallbackAllCenterNodes)
	inputCase3 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: OverPageSize,
		NodeId: nodeIdFirst, IfCenter: "false"}
	listAlarmsWithInput(inputCase3, false, "", true, CallBackStringsContains)
	// with IfCenter == true nodeid and groupId should be ignored
	inputCase4 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize,
		NodeId: nodeIdFirst, IfCenter: "true"}
	listAlarmsWithInput(inputCase4, true, "", false, CallbackAllCenterNodes)
	inputCase5 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: "true"}
	listAlarmsWithInput(inputCase5, true, "", false, CallbackAllCenterNodes)
	inputCase6 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: "false"}
	listAlarmsWithInput(inputCase6, true, "", true, CallbackAllAlarms)
	inputCase7 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize}
	listAlarmsWithInput(inputCase7, true, "", true, CallbackAllAlarms)
	inputCase8 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, GroupId: groupIdFirst,
		IfCenter: "true"}
	listAlarmsWithInput(inputCase8, true, "", true, CallbackAllCenterNodes)
	inputCase9 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: OverPageSize, IfCenter: "true"}
	listAlarmsWithInput(inputCase9, false, "", true, defaultTestCaseCallback)
	inputCase10 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: OverPageSize, GroupId: groupIdFirst,
		IfCenter: "false"}
	listAlarmsWithInput(inputCase10, false, "", true, defaultTestCaseCallback)
	inputCase11 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: ""}
	listAlarmsWithInput(inputCase11, true, "", true, CallbackAllAlarms)
	inputCase12 := types.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: ""}
	listAlarmsWithInput(inputCase12, true, "", true, CallbackAllAlarms)
}

func testGetAlarmAbnormalInput() {
	// use an event id[2] to look for alarm
	getAlarmWithInput(DefaultEventID, false)
	getAlarmWithInput(nodeIdZero, false)
}

func getAlarmWithInput(id uint64, expectRes bool) {
	resp := getAlarmDetail(id)
	if expectRes {
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	} else {
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}
func listAlarmsWithInput(input interface{}, expectRes bool, expectMsg string, ignoredMsg bool,
	callback func(msg common.RespMsg) bool) {
	resp := listAlarms(input)
	if expectRes {
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	} else {
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
	if !ignoredMsg {
		convey.So(resp.Msg, convey.ShouldEqual, expectMsg)
	}
	if expectRes {
		convey.So(callback(resp), convey.ShouldEqual, true)
	}
}

func defaultTestCaseCallback(resp common.RespMsg) bool {
	return true
}

func CallBackStringsContains(resp common.RespMsg) bool {
	return strings.Contains(resp.Msg, "Uint checker Check [PageSize] failed")
}

func CallbackAllAlarms(resp common.RespMsg) bool {
	respMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		hwlog.RunLog.Error("convert assertion failed")
		return false
	}
	respData, ok := respMap[respDataKey].(map[uint64]types.AlarmBriefInfo)
	if !ok {
		hwlog.RunLog.Error("failed to marshal alarm info")
		return false
	}
	res := true
	for _, alarm := range respData {
		res = res && alarm.AlarmType == AlarmType
	}
	return res
}

func CallbackAllCenterNodes(resp common.RespMsg) bool {
	respMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		hwlog.RunLog.Error("convert assertion failed")
		return false
	}
	respData, ok := respMap[respDataKey].(map[uint64]types.AlarmBriefInfo)
	if !ok {
		hwlog.RunLog.Error("failed to marshal alarm info")
		return false
	}
	res := true
	for _, alarm := range respData {
		res = res && alarm.NodeID == CenterNodeID
	}
	return res
}
