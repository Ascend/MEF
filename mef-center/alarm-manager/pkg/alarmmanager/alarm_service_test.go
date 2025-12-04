// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package alarmmanager test for alarm_service.go
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
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"alarm-manager/pkg/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/alarms"
	"huawei.com/mindxedge/base/common/requests"
)

const (
	normalPageSize = 15
	errPageSize    = 105
	firstPageNum   = 1
	groupIdFirst   = 1
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
	const testOldEventTotal = 200
	var err error
	for i := 0; i < testOldEventTotal; i++ {
		id, err := randIntn(math.MaxUint32)
		evenData := AlarmInfo{
			Id:                  uint64(id),
			AlarmType:           alarms.EventType,
			CreatedAt:           time.Now(),
			SerialNumber:        testEdgeSn,
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
	event, err := AlarmDbInstance().getNodeOldEvent(testEdgeSn, maxOneNodeEventCount-1)
	convey.So(err, convey.ShouldBeNil)
	convey.So(len(event), convey.ShouldBeLessThan, maxOldEventCount+1)
}

func createAlarms() {
	for _, alarm := range []AlarmInfo{
		newAlarmInfo(caseCenterAlarm),
		newAlarmInfo(caseCenterEvent),
		newAlarmInfo(caseEdgeAlarm),
		newAlarmInfo(caseEdgeEvent),
	} {
		if err := AlarmDbInstance().addAlarmInfo(&alarm); err != nil {
			continue
		}
	}
}

func TestListAlarms(t *testing.T) {
	createAlarms()
	convey.Convey("test func listAlarms: list center alarms", t, func() { testListCenter(alarms.AlarmType) })
	convey.Convey("test func listAlarms: list edge alarms by sn", t, func() { testListEdgeBySn(alarms.AlarmType) })
	convey.Convey("test func listAlarms: list edge alarms by node group id", t, func() {
		testListEdgeByNodeGroupId(alarms.AlarmType)
	})
	convey.Convey("test func listAlarms: list all edge alarms", t, func() { testListAllEdge(alarms.AlarmType) })
	convey.Convey("test func listAlarms: list all alarms", t, func() { testListAll(alarms.AlarmType) })
	convey.Convey("test func listAlarms: abnormal input", t, testListAlarmAbNormalInput)
}

func TestListEvents(t *testing.T) {
	createAlarms()
	convey.Convey("test func listEvents: list center events", t, func() { testListCenter(alarms.EventType) })
	convey.Convey("test func listEvents: list edge events by sn", t, func() { testListEdgeBySn(alarms.EventType) })
	convey.Convey("test func listEvents: list edge events by node group id", t, func() {
		testListEdgeByNodeGroupId(alarms.EventType)
	})
	convey.Convey("test func listEvents: list all edge events", t, func() { testListAllEdge(alarms.EventType) })
	convey.Convey("test func listEvents: list all events", t, func() { testListAll(alarms.EventType) })
}

func testListAll(queryType string) {
	var req = utils.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		Sn:       "",
		IfCenter: "",
	}

	var resp interface{}
	bytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	msg := model.Message{Content: bytes}
	if queryType == alarms.AlarmType {
		resp = listAlarms(&msg)
	} else {
		resp = listEvents(&msg)
	}

	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(respContent.Status, convey.ShouldEqual, common.Success)

	testListAllErrCount(req, queryType)
	testListAllErrList(req, queryType)
	testListAllErrParse(queryType)
}

func testListAllErrCount(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "countAlarmsOrEventsFullNodes",
		func(queryType string) (int64, error) {
			return 1, test.ErrTest
		})
	defer p1.Reset()

	resp := listFullAlarmOrEvents(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListAlarm})
}

func testListAllErrList(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "listAllAlarmsOrEventsDb",
		func(pageNum, pageSize uint64, queryType string) ([]AlarmInfo, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()

	resp := listFullAlarmOrEvents(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListAlarm})
}

func testListAllErrParse(queryType string) {
	msg, err := model.NewMessage()
	convey.So(err, convey.ShouldBeNil)
	dealRequest(msg, queryType)
}

func testListAllEdge(queryType string) {
	var req = utils.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		Sn:       "",
		IfCenter: utils.FalseStr,
	}

	var resp interface{}
	bytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	msg := model.Message{Content: bytes}
	if queryType == alarms.AlarmType {
		resp = listAlarms(&msg)
	} else {
		resp = listEvents(&msg)
	}

	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(respContent.Status, convey.ShouldEqual, common.Success)

	testListAllEdgeErrCount(req, queryType)
	testListAllEdgeErrList(req, queryType)
}

func testListAllEdgeErrCount(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "countEdgeAlarmsOrEvents",
		func(queryType string) (int64, error) {
			return 1, test.ErrTest
		})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEvents(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListAlarm})
}

func testListAllEdgeErrList(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "listAllEdgeAlarmsOrEventsDb",
		func(pageNum, pageSize uint64, queryType string) ([]AlarmInfo, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEvents(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListAlarm})
}

func TestGetAlarmDetail(t *testing.T) {
	convey.Convey("normal input for getting alarm by id with Normal Input", t, func() {
		testGetAlarmOrEventByInfoId(alarms.AlarmType)
	})
	convey.Convey("abnormal input for getting alarm", t, testGetAlarmAbnormalInput)
	convey.Convey("test func getAlarmOrEventDbDetail failed, parse param failed", t, testGetAlarmDetailErrParse)
	convey.Convey("test func getAlarmOrEventDbDetail failed, getAlarmOrEventInfoByAlarmInfoId failed", t,
		testGetAlarmDetailErrGetInfo)
}

func TestGetEventDetail(t *testing.T) {
	convey.Convey("normal input for getting alarm by id with Normal Input", t, func() {
		testGetAlarmOrEventByInfoId(alarms.EventType)
	})
}

func testGetAlarmDetailErrParse() {
	msg, err := model.NewMessage()
	convey.So(err, convey.ShouldBeNil)
	resp := getAlarmOrEventDbDetail(msg, alarms.AlarmType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"})
}

func testGetAlarmDetailErrGetInfo() {
	queryType := alarms.AlarmType
	inputId := 1

	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "getAlarmOrEventInfoByAlarmInfoId",
		func(Id uint64) (*AlarmInfo, error) {
			return &AlarmInfo{}, gorm.ErrRecordNotFound
		})
	defer p1.Reset()

	msg, err := model.NewMessage()
	convey.So(err, convey.ShouldBeNil)
	err = msg.FillContent(inputId)
	convey.So(err, convey.ShouldBeNil)
	resp := getAlarmOrEventDbDetail(msg, queryType)
	expResp := &common.RespMsg{Status: common.ErrorGetAlarmDetail, Msg: fmt.Sprintf("id [%d] not found", inputId)}
	convey.So(resp, convey.ShouldResemble, expResp)

	var p2 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "getAlarmOrEventInfoByAlarmInfoId",
		func(Id uint64) (*AlarmInfo, error) {
			return &AlarmInfo{}, test.ErrTest
		})
	defer p2.Reset()
	resp = getAlarmOrEventDbDetail(msg, queryType)
	expResp = &common.RespMsg{Status: common.ErrorGetAlarmDetail}
	convey.So(resp, convey.ShouldResemble, expResp)

	var p3 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "getAlarmOrEventInfoByAlarmInfoId",
		func(Id uint64) (*AlarmInfo, error) {
			return &AlarmInfo{AlarmType: alarms.EventType}, nil
		})
	defer p3.Reset()
	resp = getAlarmOrEventDbDetail(msg, queryType)
	expResp = &common.RespMsg{Status: common.ErrorParamInvalid,
		Msg: fmt.Sprintf("the inputID[%d] is not an ID of %s", inputId, queryType)}
	convey.So(resp, convey.ShouldResemble, expResp)
}

func testListEdgeBySn(queryType string) {
	var req = utils.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		Sn:       testEdgeSn,
		IfCenter: utils.FalseStr,
	}

	var resp interface{}
	bytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	msg := model.Message{Content: bytes}
	if queryType == alarms.AlarmType {
		resp = listAlarms(&msg)
	} else {
		resp = listEvents(&msg)
	}

	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(respContent.Status, convey.ShouldEqual, common.Success)

	respData, ok := respContent.Data.(utils.ListAlarmsResp)
	convey.So(ok, convey.ShouldBeTrue)

	res := true
	for _, alarm := range respData.Records {
		res = res && alarm.Sn == req.Sn && (alarm.AlarmType == queryType)
	}
	convey.So(res, convey.ShouldBeTrue)

	testListEdgeBySnErrCount(req, queryType)
	testListEdgeBySnErrList(req, queryType)
}

func testListEdgeBySnErrCount(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "countAlarmsOrEventsBySn",
		func(sn string, queryType string) (int64, error) {
			return 1, test.ErrTest
		})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEventsBySn(req, queryType)
	expResp := &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm,
		Msg: fmt.Sprintf("failed to count %s", queryType)}
	convey.So(resp, convey.ShouldResemble, expResp)
}

func testListEdgeBySnErrList(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "listEdgeAlarmsOrEventsDb",
		func(pageNum, pageSize uint64, queryType string) ([]AlarmInfo, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEventsBySn(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm})
}

func testListCenter(queryType string) {
	var req = utils.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		Sn:       testEdgeSn,
		IfCenter: utils.TrueStr,
	}

	var resp interface{}
	bytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	msg := model.Message{Content: bytes}
	if queryType == alarms.AlarmType {
		resp = listAlarms(&msg)
	} else {
		resp = listEvents(&msg)
	}

	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(respContent.Status, convey.ShouldEqual, common.Success)
	respData, ok := respContent.Data.(utils.ListAlarmsResp)
	convey.So(ok, convey.ShouldBeTrue)

	res := true
	for _, alarm := range respData.Records {
		res = res && (alarm.Sn == alarms.CenterSn) && (alarm.AlarmType == queryType)
	}
	convey.So(res, convey.ShouldBeTrue)

	testListCenterErrCount(req, queryType)
	testListCenterErrList(req, queryType)
}

func testListCenterErrCount(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "countAlarmsOrEventsBySn",
		func(sn string, queryType string) (int64, error) {
			return 1, test.ErrTest
		})
	defer p1.Reset()

	resp := listCenterAlarmOrEvents(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListCenterNodeAlarm})
}

func testListCenterErrList(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "listCenterAlarmsOrEventsDb",
		func(pageNum, pageSize uint64, queryType string) ([]AlarmInfo, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()

	resp := listCenterAlarmOrEvents(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListCenterNodeAlarm})
}

func testListEdgeByNodeGroupId(queryType string) {
	var req = utils.ListAlarmOrEventReq{
		PageNum:  firstPageNum,
		PageSize: normalPageSize,
		GroupId:  groupIdFirst,
		IfCenter: utils.FalseStr,
	}

	var resp interface{}
	bytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	msg := model.Message{Content: bytes}
	if queryType == alarms.AlarmType {
		resp = listAlarms(&msg)
	} else {
		resp = listEvents(&msg)
	}

	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(respContent.Status, convey.ShouldEqual, common.Success)

	testListEdgeByNodeGroupIdErrMarshal(req, queryType)
	testListEdgeByNodeGroupIdErrNotFound(req, queryType)
	testListEdgeByNodeGroupIdErrResp(req, queryType)
	testListEdgeByNodeGroupIdErrRespUnmarshal(req, queryType)
	testListEdgeByNodeGroupIdErrCount(req, queryType)
	testListEdgeByNodeGroupIdErrList(req, queryType)
}

func testListEdgeByNodeGroupIdErrMarshal(req utils.ListAlarmOrEventReq, queryType string) {
	getSnsReq := requests.GetSnsReq{GroupId: req.GroupId}
	bytes, err := json.Marshal(getSnsReq)
	convey.So(err, convey.ShouldBeNil)

	var p1 = gomonkey.ApplyFuncSeq(json.Marshal, []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, test.ErrTest}},
		{Values: gomonkey.Params{bytes, nil}},
		{Values: gomonkey.Params{nil, test.ErrTest}},
	})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEventsByGroupId(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorParamInvalid})

	resp = listEdgeAlarmsOrEventsByGroupId(req, queryType)
	expResp := &common.RespMsg{Status: common.ErrorDecodeRespFromEdgeMgr, Msg: "marshal sns information failed"}
	convey.So(resp, convey.ShouldResemble, expResp)
}

func testListEdgeByNodeGroupIdErrNotFound(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyFuncReturn(common.SendSyncMessageByRestful,
		common.RespMsg{Status: common.ErrorNodeGroupNotFound, Msg: "", Data: nil})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEventsByGroupId(req, queryType)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testListEdgeByNodeGroupIdErrResp(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyFuncReturn(common.SendSyncMessageByRestful,
		common.RespMsg{Status: common.ErrorDecodeRespFromEdgeMgr, Msg: "", Data: nil})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEventsByGroupId(req, queryType)
	errMsg, _ := common.ErrorMap[resp.Status]
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorDecodeRespFromEdgeMgr, Msg: errMsg})
}

func testListEdgeByNodeGroupIdErrRespUnmarshal(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyFuncReturn(common.SendSyncMessageByRestful,
		common.RespMsg{Status: common.Success, Msg: "", Data: "test response"}).
		ApplyFuncReturn(json.Unmarshal, test.ErrTest)
	defer p1.Reset()

	resp := listEdgeAlarmsOrEventsByGroupId(req, queryType)
	expResp := &common.RespMsg{Status: common.ErrorDecodeRespFromEdgeMgr, Msg: "unmarshal sns information failed"}
	convey.So(resp, convey.ShouldResemble, expResp)
}

func testListEdgeByNodeGroupIdErrCount(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "countAlarmsOrEventsBySns",
		func(sns []string, queryType string) (int64, error) {
			return 1, test.ErrTest
		})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEventsByGroupId(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm})
}

func testListEdgeByNodeGroupIdErrList(req utils.ListAlarmOrEventReq, queryType string) {
	var p1 = gomonkey.ApplyPrivateMethod(&AlarmDbHandler{}, "countAlarmsOrEventsBySns",
		func(sns []string, queryType string) (int64, error) {
			return 1, nil
		}).
		ApplyPrivateMethod(&AlarmDbHandler{}, "listAlarmsOrEventsOfGroup",
			func(pageNum, pageSize uint64, sns []string, queryType string) ([]AlarmInfo, error) {
				return nil, test.ErrTest
			})
	defer p1.Reset()

	resp := listEdgeAlarmsOrEventsByGroupId(req, queryType)
	convey.So(resp, convey.ShouldResemble, &common.RespMsg{Status: common.ErrorListEdgeNodeAlarm})
}

func testGetAlarmOrEventByInfoId(queryType string) {
	centerAlarm, err := AlarmDbInstance().getAlarmInfo(testAlarmId, alarms.CenterSn)
	if err != nil || len(centerAlarm) < 1 {
		panic("get center alarm info failed")
	}
	var centerAlarmInfo AlarmInfo
	var centerEventInfo AlarmInfo
	for _, v := range centerAlarm {
		switch v.AlarmType {
		case alarms.AlarmType:
			centerAlarmInfo = v
		case alarms.EventType:
			centerEventInfo = v
		default:
			continue
		}
	}

	var resp interface{}
	if queryType == alarms.AlarmType {
		msg := model.Message{Content: []byte(fmt.Sprintf("%d", centerAlarmInfo.Id))}
		resp = getAlarmDetail(&msg)
	} else {
		msg := model.Message{Content: []byte(fmt.Sprintf("%d", centerEventInfo.Id))}
		resp = getEventDetail(&msg)
	}

	respContent, ok := resp.(*common.RespMsg)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(respContent.Status, convey.ShouldEqual, common.Success)
	alarm, ok := respContent.Data.(*AlarmInfo)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(alarm.AlarmType, convey.ShouldEqual, queryType)
}

func testListAlarmAbNormalInput() {
	inputCase1 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize,
		Sn: testEdgeSn, GroupId: groupIdFirst, IfCenter: "false"}
	listAlarmsWithInput(inputCase1, false, "sn and groupId can't exist at the same time when"+
		" ifCenter is not true", false,
		defaultTestCaseCallback)
	// with IfCenter == true sn and groupId should be ignored
	inputCase2 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, Sn: testEdgeSn,
		GroupId: 1, IfCenter: "true"}
	listAlarmsWithInput(inputCase2, true, "", false, CallbackAllCenterNodes)
	inputCase3 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: errPageSize,
		Sn: testEdgeSn, IfCenter: "false"}
	listAlarmsWithInput(inputCase3, false, "", true, CallBackStringsContains)
	// with IfCenter == true sn and groupId should be ignored
	inputCase4 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize,
		Sn: testEdgeSn, IfCenter: "true"}
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
	inputCase9 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: errPageSize, IfCenter: "true"}
	listAlarmsWithInput(inputCase9, false, "", true, defaultTestCaseCallback)
	inputCase10 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: errPageSize, GroupId: groupIdFirst,
		IfCenter: "false"}
	listAlarmsWithInput(inputCase10, false, "", true, defaultTestCaseCallback)
	inputCase11 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: ""}
	listAlarmsWithInput(inputCase11, true, "", true, CallbackAllAlarms)
	inputCase12 := utils.ListAlarmOrEventReq{PageNum: firstPageNum, PageSize: normalPageSize, IfCenter: ""}
	listAlarmsWithInput(inputCase12, true, "", true, CallbackAllAlarms)
}

func testGetAlarmAbnormalInput() {
	var defaultEventID uint64
	// use an event id[2] to look for alarm
	getAlarmWithInput(defaultEventID, false)
	getAlarmWithInput(0, false)
}

func getAlarmWithInput(id uint64, expectRes bool) {
	msg := model.Message{Content: []byte(fmt.Sprintf("%d", id))}
	resp := getAlarmDetail(&msg)
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
	resp := listAlarms(&model.Message{Content: bytes})
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
