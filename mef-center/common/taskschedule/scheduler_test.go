// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package taskschedule
package taskschedule

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestTaskConcurrentControl(t *testing.T) {
	convey.Convey("test task concurrent control", t, func() {
		DefaultScheduler().RegisterGoroutinePool(GoroutinePoolSpec{
			Id:             "TestTaskConcurrentControl",
			MaxConcurrency: 1,
			MaxCapacity:    1,
		})
		DefaultScheduler().RegisterExecutorFactory(
			NewExecutorFactory("TestTaskConcurrentControl", func(context TaskContext) {}))
		time.Sleep(oneHundredMs)

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:            "TestTaskConcurrentControl-1",
			Command:       "TestTaskConcurrentControl",
			GoroutinePool: "TestTaskConcurrentControl",
		}), convey.ShouldBeNil)

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:            "TestTaskConcurrentControl-2",
			Command:       "TestTaskConcurrentControl",
			GoroutinePool: "TestTaskConcurrentControl",
		}), convey.ShouldBeNil)

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:            "TestTaskConcurrentControl-3",
			Command:       "TestTaskConcurrentControl",
			GoroutinePool: "TestTaskConcurrentControl",
		}), convey.ShouldNotBeNil)

		time.Sleep(oneHundredMs)

		ctx, err := DefaultScheduler().GetTaskContext("TestTaskConcurrentControl-1")
		convey.So(err, convey.ShouldBeNil)
		status, err := ctx.GetStatus()
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Phase, convey.ShouldEqual, Processing)

		ctx, err = DefaultScheduler().GetTaskContext("TestTaskConcurrentControl-2")
		convey.So(err, convey.ShouldBeNil)
		status, err = ctx.GetStatus()
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Phase, convey.ShouldEqual, Waiting)
	})
}

func TestTaskIterate(t *testing.T) {
	convey.Convey("test task iterate", t, func() {
		immediatelySucceedTask := func(context TaskContext) { _ = context.UpdateStatus(TaskStatus{Phase: Succeed}) }

		const (
			capacity3    = 3
			concurrency3 = 3
		)
		DefaultScheduler().RegisterGoroutinePool(GoroutinePoolSpec{
			Id:             "TestTaskIterate",
			MaxConcurrency: concurrency3,
			MaxCapacity:    capacity3,
		})
		DefaultScheduler().RegisterExecutorFactory(
			NewExecutorFactory("TestTaskIterateMaster", func(context TaskContext) {}))
		DefaultScheduler().RegisterExecutorFactory(
			NewExecutorFactory("TestTaskIterateSlave", immediatelySucceedTask))

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:            "TestTaskIterate-1",
			Command:       "TestTaskIterateMaster",
			GoroutinePool: "TestTaskIterate",
		}), convey.ShouldBeNil)

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:            "TestTaskIterate-2",
			Command:       "TestTaskIterateSlave",
			GoroutinePool: "TestTaskIterate",
			ParentId:      "TestTaskIterate-1",
		}), convey.ShouldBeNil)

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:            "TestTaskIterate-3",
			Command:       "TestTaskIterateSlave",
			GoroutinePool: "TestTaskIterate",
			ParentId:      "TestTaskIterate-1",
		}), convey.ShouldBeNil)

		time.Sleep(oneHundredMs)
		subTaskIter := DefaultScheduler().NewSubTaskSelector("TestTaskIterate-1")
		for {
			_, err := subTaskIter.Select()
			if err == ErrNoRunningSubTask {
				break
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		}
	})
}
