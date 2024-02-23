// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package restful test for router.go
package restful

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"alarm-manager/pkg/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/restfulmgr"
)

const (
	testParamName     = "id"
	errParamName      = "ip"
	testIdStringValue = "1"
	testIdUintValue   = 1

	testPageNum     = 1
	testPageNumStr  = "1"
	errPageNum      = "error page num"
	testPageSize    = 10
	testPageSizeStr = "10"
	errPageSize     = "error page size"

	testGroupId = 1
	errGroupId  = 0
	testSn      = "test sn"
)

func TestSetRouter(t *testing.T) {
	convey.Convey("test func setRouter", t, func() {
		engine := gin.New()
		setRouter(engine)
	})
}

func TestQueryDispatcherParseData(t *testing.T) {
	convey.Convey("test queryDispatcher method 'ParseData', is string", t, testQueryDispatcherParseDataIsString)
	convey.Convey("test queryDispatcher method 'ParseData', not string", t, testQueryDispatcherParseDataNotString)
}

func testQueryDispatcherParseDataIsString() {
	dispatcher := queryDispatcher{
		GenericDispatcher: restfulmgr.GenericDispatcher{
			RelativePath: "/alarm",
			Method:       http.MethodGet,
			Destination:  common.AlarmManagerName},
		name:     testParamName,
		isString: true,
	}

	// success
	ctx := &gin.Context{}
	rawURL := fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?%s=%s", testParamName, testIdStringValue)
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err := dispatcher.ParseData(ctx)
	convey.So(res, convey.ShouldEqual, testIdStringValue)
	convey.So(err, convey.ShouldBeNil)

	// param name is invalid
	ctx = &gin.Context{}
	rawURL = fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?%s=%s", errParamName, testIdStringValue)
	u, err = url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err = dispatcher.ParseData(ctx)
	convey.So(res, convey.ShouldEqual, "")
	expErr := fmt.Errorf("req string para [%s] is invalid", testParamName)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testQueryDispatcherParseDataNotString() {
	dispatcher := queryDispatcher{
		GenericDispatcher: restfulmgr.GenericDispatcher{
			RelativePath: "/alarm",
			Method:       http.MethodGet,
			Destination:  common.AlarmManagerName},
		name:     testParamName,
		isString: false,
	}

	// success case
	ctx := &gin.Context{}
	rawURL := fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?%s=%d", testParamName, testIdUintValue)
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err := dispatcher.ParseData(ctx)
	convey.So(res, convey.ShouldEqual, testIdUintValue)
	convey.So(err, convey.ShouldBeNil)

	// parse to uint failed
	var p1 = gomonkey.ApplyFuncReturn(strconv.ParseUint, uint64(0), test.ErrTest)
	defer p1.Reset()
	res, err = dispatcher.ParseData(ctx)
	convey.So(res, convey.ShouldEqual, 0)
	expErr := fmt.Errorf("req int para [%s] is invalid", testParamName)
	convey.So(err, convey.ShouldResemble, expErr)
}

var dispatcher listDispatcher

func TestListDispatcherParseData(t *testing.T) {
	dispatcher = listDispatcher{
		GenericDispatcher: restfulmgr.GenericDispatcher{
			RelativePath: "/alarms",
			Method:       http.MethodGet,
			Destination:  common.AlarmManagerName,
		},
	}
	convey.Convey("test listDispatcher method 'ParseData' success, is center", t, testListParseDataIsCenter)
	convey.Convey("test listDispatcher method 'ParseData' failed, param error", t, testListParseDataIsCenterErrParam)
	convey.Convey("test listDispatcher method 'ParseData' success, is edge", t, testListParseDataIsEdge)
	convey.Convey("test listDispatcher method 'ParseData' failed, group id error", t, testListParseDataIsEdgeErrGroupId)
	convey.Convey("test listDispatcher method 'ParseData' failed, uint parse error", t, testListParseDataIsEdgeErrParse)
}

func testListParseDataIsCenter() {
	ctx := &gin.Context{}
	rawURL := fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?pageNum=%d&pageSize=%d&ifCenter=%s",
		testPageNum, testPageSize, utils.TrueStr)
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err := dispatcher.ParseData(ctx)
	expRes := utils.ListAlarmOrEventReq{PageNum: testPageNum, PageSize: testPageSize, IfCenter: utils.TrueStr}
	convey.So(res, convey.ShouldResemble, expRes)
	convey.So(err, convey.ShouldBeNil)
}

func testListParseDataIsCenterErrParam() {
	// pageNum or pageSize is invalid
	ctx := &gin.Context{}
	rawURL := fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?pageNum=%s&pageSize=%s&ifCenter=%s",
		errPageNum, errPageSize, utils.TrueStr)
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err := dispatcher.ParseData(ctx)
	convey.So(res, convey.ShouldBeNil)
	expErr := fmt.Errorf("pageNum[%s] or pageSize[%s] is invalid", errPageNum, errPageSize)
	convey.So(err, convey.ShouldResemble, expErr)

	// ifCenter or groupId or sn is invalid
	ctx = &gin.Context{}
	rawURL = fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?pageNum=%d&pageSize=%d&ifCenter=",
		testPageNum, testPageSize)
	u, err = url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err = dispatcher.ParseData(ctx)
	convey.So(res, convey.ShouldBeNil)
	expErr = fmt.Errorf("params in [%s,%s,%s] cannot be assigned to empty string", ifCenterKey, groupIdKey, snKey)
	convey.So(err, convey.ShouldResemble, expErr)
}

func testListParseDataIsEdge() {
	ctx := &gin.Context{}
	rawURL := fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?pageNum=%d&pageSize=%d&groupId=%d&sn=%s",
		testPageNum, testPageSize, testGroupId, testSn)
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err := dispatcher.ParseData(ctx)
	expRes := utils.ListAlarmOrEventReq{PageNum: testPageNum, PageSize: testPageSize, Sn: testSn, GroupId: testGroupId,
		IfCenter: ""}
	convey.So(res, convey.ShouldResemble, expRes)
	convey.So(err, convey.ShouldBeNil)
}

func testListParseDataIsEdgeErrGroupId() {
	// group id is "0"
	ctx := &gin.Context{}
	rawURL := fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?pageNum=%d&pageSize=%d&groupId=%d&sn=%s",
		testPageNum, testPageSize, errGroupId, testSn)
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err := dispatcher.ParseData(ctx)
	convey.So(res, convey.ShouldBeNil)
	expErr := fmt.Errorf("groupId cannot be assigned to 0")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testListParseDataIsEdgeErrParse() {
	var p1 = gomonkey.ApplyFuncReturn(strconv.ParseUint, uint64(0), test.ErrTest)
	defer p1.Reset()
	ctx := &gin.Context{}
	rawURL := fmt.Sprintf("https://127.0.01:30035/alarmmanager/v1/alarms?pageNum=%d&pageSize=%d&groupId=%d&sn=%s",
		testPageNum, testPageSize, testGroupId, testSn)
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	ctx.Request = &http.Request{URL: u}
	res, err := dispatcher.ParseData(ctx)
	convey.So(res, convey.ShouldBeNil)
	expErr := fmt.Errorf("pageNum[%s] or pageSize[%s] is invalid", testPageNumStr, testPageSizeStr)
	convey.So(err, convey.ShouldResemble, expErr)
}
