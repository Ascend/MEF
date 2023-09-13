// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package tasks
package tasks

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/logmanager/testutils"
)

func TestInitTasks(t *testing.T) {
	convey.Convey("test init tasks", t, func() {
		dummyObjs := testutils.DummyTaskSchedule()
		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, dummyObjs.Scheduler).
			ApplyMethodReturn(dummyObjs.Scheduler, "RegisterExecutorFactory", false).
			ApplyMethodReturn(dummyObjs.Scheduler, "RegisterGoroutinePool", false)
		defer patch.Reset()

		err := InitTasks()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestSubmitLogDumpTask(t *testing.T) {
	convey.Convey("test init tasks", t, func() {
		dummyObjs := testutils.DummyTaskSchedule()
		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, dummyObjs.Scheduler).
			ApplyMethodReturn(dummyObjs.Scheduler, "SubmitTask", nil)
		defer patch.Reset()

		_, err := SubmitLogDumpTask(nil)
		convey.So(err, convey.ShouldBeNil)
	})
}
