// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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

		_, err := SubmitLogDumpTask(nil, nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}
