// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package alarmmanager for testing module
package alarmmanager

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindx/common/hwlog"

	"alarm-manager/pkg/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
)

const (
	normalPageSize    = 15
	OverPageSize      = 105
	firstPageNum      = 1
	firstPageSize     = 10
	testSn1           = "testSn1"
	testSn2           = "testSn2"
	groupIdFirst      = 1
	testOldEventTotal = 200
)

func TestAddAlarm(t *testing.T) {
	convey.Convey("normal input for add alarm", t, func() {
		testAddAlarm()
	})
}

func testAddAlarm() {
	for i := 0; i < 200; i++ {
		var alarm AlarmInfo
		randSetOneAlarm(&alarm)
		sn, _ := randIntn(math.MaxUint32)
		alarm.SerialNumber = strconv.Itoa(sn)
		err := AlarmDbInstance().addAlarmInfo(&alarm)
		convey.So(err, convey.ShouldBeNil)
	}
}

func TestGetNodeOldEvent(t *testing.T) {
	convey.Convey("test limit oldest events ", t, func() {
		testGetNodeOldEvent()
	})
}

func testGetNodeOldEvent() {
	var err error
	for i := 0; i < testOldEventTotal; i++ {
		id, err := randIntn(math.MaxUint32)
		evenData := AlarmInfo{
			Id:                  uint64(id),
			AlarmType:           alarms.EventType,
			CreatedAt:           time.Now(),
			SerialNumber:        testSn1,
			Ip:                  "10.10.10.10",
			AlarmId:             "AlarmId",
			AlarmName:           "AlarmName",
			PerceivedSeverity:   "PerceivedSeverity",
			DetailedInformation: "DetailedInformation",
			Suggestion:          "Suggestion",
			Reason:              "Reason",
			Impact:              "Impact",
			Resource:            "Resource",
		}
		err = AlarmDbInstance().addAlarmInfo(&evenData)
		if err != nil {
			break
		}
	}
	convey.So(err, convey.ShouldBeNil)
	event, err := AlarmDbInstance().getNodeOldEvent(testSn1, maxOneNodeEventCount-1)
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(event), convey.ShouldBeLessThan, maxOldEventCount+1)
}

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
	bytes, err := json.Marshal(reqData)
	convey.So(err, convey.ShouldBeNil)
	msg := model.Message{Content: bytes}
	if queryType == alarms.AlarmType {
		resp, err = listAlarms(&msg)
	} else {
		resp, err = listEvents(&msg)
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
	bytes, err := json.Marshal(reqData)
	convey.So(err, convey.ShouldBeNil)
	msg := model.Message{Content: bytes}
	if queryType == alarms.AlarmType {
		resp, err = listAlarms(&msg)
	} else {
		resp, err = listEvents(&msg)
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
	bytes, err := json.Marshal(reqData)
	convey.So(err, convey.ShouldBeNil)
	msg := model.Message{Content: bytes}
	if queryType == alarms.AlarmType {
		resp, err = listAlarms(&msg)
	} else {
		resp, err = listEvents(&msg)
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
		msg := model.Message{Content: []byte(fmt.Sprintf("%d", defaultAlarmID))}
		resp, err = getAlarmDetail(&msg)
	} else {
		msg := model.Message{Content: []byte(fmt.Sprintf("%d", defaultEventID))}
		resp, err = getEventDetail(&msg)
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
	getAlarmWithInput(defaultEventID, false)
	getAlarmWithInput(0, false)
}

func getAlarmWithInput(id uint64, expectRes bool) {
	msg := model.Message{Content: []byte(fmt.Sprintf("%d", id))}
	resp, err := getAlarmDetail(&msg)
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
	bytes, err := json.Marshal(input)
	convey.So(err, convey.ShouldBeNil)
	resp, err := listAlarms(&model.Message{Content: bytes})
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

func TestDispatcher(t *testing.T) {
	var resp *model.Message
	patch := gomonkey.ApplyFunc(modulemgr.SendMessage, func(m *model.Message) error {
		resp = m
		return nil
	})
	defer patch.Reset()
	convey.Convey("test dispatcher", t, func() {
		content := utils.ListAlarmOrEventReq{
			PageNum:  1,
			PageSize: 10,
		}
		msg := model.Message{}
		msg.SetRouter("", "", http.MethodGet, "/alarmmanager/v1/alarms")
		err := msg.FillContent(content, true)
		convey.So(err, convey.ShouldBeNil)
		alarmMgr := &alarmManager{}
		alarmMgr.dispatch(&msg)
		convey.So(resp, convey.ShouldNotBeNil)

		var listRet common.RespMsg
		if resp == nil {
			return
		}
		err = json.Unmarshal(resp.Content, &listRet)
		convey.So(err, convey.ShouldBeNil)
		convey.So(listRet.Status, convey.ShouldEqual, common.Success)
		bytes, err := json.Marshal(listRet.Data)
		convey.So(err, convey.ShouldBeNil)
		var alarmList utils.ListAlarmsResp
		err = json.Unmarshal(bytes, &alarmList)
		convey.So(err, convey.ShouldBeNil)
	})
}
