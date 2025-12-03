// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package alarmmanager test for alarm_init.go
package alarmmanager

import (
	"context"
	"net/http"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"alarm-manager/pkg/monitors"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

var (
	alarmMgr *alarmManager
	ctx      context.Context
	cancel   context.CancelFunc
)

func TestAlarmManager(t *testing.T) {
	const testDbPath = "./test.db"
	ctx, cancel = context.WithCancel(context.Background())
	modulemgr.ModuleInit()
	if err := modulemgr.Registry(NewAlarmManager(testDbPath, true, ctx)); err != nil {
		panic(err)
	}
	alarmMgr = &alarmManager{
		dbPath: testDbPath,
		enable: true,
		ctx:    ctx,
	}
	convey.Convey("test alarmManager method 'NewAlarmManager', 'Name', 'Enable'", t, testAlarmManager)
	convey.Convey("test alarmManager method 'Start'", t, testAlarmMgrStart)
	convey.Convey("test alarmManager method 'dispatch'", t, testAlarmMgrDispatch)
	convey.Convey("test alarmManager method 'dispatch' failed", t, testAlarmMgrDispatchErr)
	convey.Convey("test alarmManager method 'startMonitoring'", t, testAlarmMgrStartMonitoring)
	convey.Convey("test func clearEdgeAlarms", t, testClearEdgeAlarms)
}

func testAlarmManager() {
	if alarmMgr == nil {
		panic("alarm manager is nil")
	}
	convey.So(alarmMgr.Name(), convey.ShouldEqual, common.AlarmManagerName)
	convey.So(alarmMgr.Enable(), convey.ShouldBeTrue)
}

func testAlarmMgrStart() {
	var p1 = gomonkey.ApplyFuncSeq(modulemgr.ReceiveMessage,
		[]gomonkey.OutputCell{
			{Values: gomonkey.Params{model.Message{}, test.ErrTest}},
			{Values: gomonkey.Params{model.Message{}, nil}},
		}).
		ApplyPrivateMethod(&alarmManager{}, "dispatch", func(*model.Message) { return }).
		ApplyPrivateMethod(&alarmManager{}, "startMonitoring", func() { return }).
		ApplyPrivateMethod(&alarmManager{}, "checkAlarmNum", func() { return })
	defer p1.Reset()
	if alarmMgr == nil {
		panic("alarm manager is nil")
	}
	go alarmMgr.Start()
	cancel()
}

func testAlarmMgrDispatch() {
	var p1 = gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
	defer p1.Reset()

	msg, err := model.NewMessage()
	if err != nil {
		panic(err)
	}
	if alarmMgr == nil {
		panic("alarm manager is nil")
	}
	testCase := []struct {
		method   string
		resource string
	}{
		{
			http.MethodGet,
			listAlarmRouter,
		},
		{
			http.MethodGet,
			getAlarmDetailRouter,
		},
		{
			http.MethodGet,
			listEventsRouter,
		},
		{
			http.MethodGet,
			getEventDetailRouter,
		},
		{
			http.MethodPost,
			requests.ReportAlarmRouter,
		},
		{
			http.MethodGet,
			requests.ClearOneNodeAlarmRouter,
		},
	}
	for _, v := range testCase {
		msg.Router.Option = v.method
		msg.Router.Resource = v.resource
		alarmMgr.dispatch(msg)
	}
}

func testAlarmMgrDispatchErr() {
	var p1 = gomonkey.ApplyFuncSeq(methodSelect,
		[]gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{&model.Message{}}, Times: 3},
		}).ApplyMethodSeq(&model.Message{}, "NewResponse",
		[]gomonkey.OutputCell{
			{Values: gomonkey.Params{&model.Message{}, test.ErrTest}},
			{Values: gomonkey.Params{&model.Message{}, nil}, Times: 2},
		}).ApplyMethodSeq(&model.Message{}, "FillContent",
		[]gomonkey.OutputCell{
			{Values: gomonkey.Params{test.ErrTest}},
			{Values: gomonkey.Params{nil}},
		}).ApplyFuncSeq(modulemgr.SendMessage,
		[]gomonkey.OutputCell{
			{Values: gomonkey.Params{test.ErrTest}},
		})
	defer p1.Reset()

	msg, err := model.NewMessage()
	if err != nil {
		panic(err)
	}
	msg.Router.Option = http.MethodGet
	msg.Router.Resource = listAlarmRouter

	if alarmMgr == nil {
		panic("alarm manager is nil")
	}
	alarmMgr.dispatch(msg)
	alarmMgr.dispatch(msg)
	alarmMgr.dispatch(msg)
	alarmMgr.dispatch(msg)
}

func testAlarmMgrStartMonitoring() {
	var p1 = gomonkey.ApplyFuncReturn(monitors.GetAlarmMonitorList,
		[]monitors.AlarmMonitor{&testTask{name: "test cert task name"}})
	defer p1.Reset()
	if alarmMgr == nil {
		panic("alarm manager is nil")
	}
	alarmMgr.startMonitoring()
}

type testTask struct {
	name string
}

// Monitoring monitor one task and call collectOnce
func (tt *testTask) Monitoring(ctx context.Context) {
}

// CollectOnce call check func and send alarm req
func (tt *testTask) CollectOnce() {
}

func testClearEdgeAlarms() {
	var p1 = gomonkey.ApplyFuncSeq(common.GetItemCount,
		[]gomonkey.OutputCell{
			{Values: gomonkey.Params{1, test.ErrTest}},
			{Values: gomonkey.Params{1, nil}},
			{Values: gomonkey.Params{100001, nil}, Times: 2},
		}).ApplyMethodSeq(&AlarmDbHandler{}, "DeleteEdgeAlarm",
		[]gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{test.ErrTest}},
		})
	defer p1.Reset()

	err := clearEdgeAlarms()
	convey.So(err, convey.ShouldResemble, test.ErrTest)
	err = clearEdgeAlarms()
	convey.So(err, convey.ShouldBeNil)
	err = clearEdgeAlarms()
	convey.So(err, convey.ShouldBeNil)
	err = clearEdgeAlarms()
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}
