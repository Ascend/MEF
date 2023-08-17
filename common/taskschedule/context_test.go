// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

import (
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

const (
	oneHundredMs  = 100 * time.Millisecond
	twoHundredsMs = 200 * time.Millisecond
)

func TestTaskWaitTimeout(t *testing.T) {
	convey.Convey("test task wait timeout", t, func() {
		DefaultScheduler().RegisterGoroutinePool(GoroutinePoolSpec{
			Id:             "TestTaskWaitTimeout",
			MaxConcurrency: 1,
			MaxCapacity:    2,
		})
		DefaultScheduler().RegisterExecutorFactory(
			NewExecutorFactory("TestTaskWaitTimeout", func(context TaskContext) {}))

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:            "TestTaskWaitTimeout-1",
			Command:       "TestTaskWaitTimeout",
			GoroutinePool: "TestTaskWaitTimeout",
		}), convey.ShouldBeNil)
		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:            "TestTaskWaitTimeout-2",
			Command:       "TestTaskWaitTimeout",
			GoroutinePool: "TestTaskWaitTimeout",
			WaitTimeout:   oneHundredMs,
		}), convey.ShouldBeNil)

		ctx, err := DefaultScheduler().GetTaskContext("TestTaskWaitTimeout-2")
		convey.So(err, convey.ShouldBeNil)

		status, err := ctx.GetStatus()
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Phase, convey.ShouldEqual, Waiting)

		time.Sleep(twoHundredsMs)

		status, err = ctx.GetStatus()
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Phase, convey.ShouldEqual, Failed)
	})
}

func TestTaskExecuteTimeout(t *testing.T) {
	convey.Convey("test task execute timeout", t, func() {
		DefaultScheduler().RegisterGoroutinePool(GoroutinePoolSpec{
			Id:             "TestTaskExecuteTimeout",
			MaxConcurrency: 1,
			MaxCapacity:    1,
		})
		DefaultScheduler().RegisterExecutorFactory(
			NewExecutorFactory("TestTaskExecuteTimeout", func(context TaskContext) {}))

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:             "TestTaskExecuteTimeout-1",
			Command:        "TestTaskExecuteTimeout",
			GoroutinePool:  "TestTaskExecuteTimeout",
			ExecuteTimeout: twoHundredsMs,
		}), convey.ShouldBeNil)

		ctx, err := DefaultScheduler().GetTaskContext("TestTaskExecuteTimeout-1")
		convey.So(err, convey.ShouldBeNil)

		time.Sleep(oneHundredMs)

		status, err := ctx.GetStatus()
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Phase, convey.ShouldEqual, Progressing)

		time.Sleep(twoHundredsMs)

		status, err = ctx.GetStatus()
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Phase, convey.ShouldEqual, Failed)
	})
}

func TestTaskGracefulShutdownTimeout(t *testing.T) {
	convey.Convey("test task graceful shutdown timeout", t, func() {
		DefaultScheduler().RegisterGoroutinePool(GoroutinePoolSpec{
			Id:             "TestTaskGracefulShutdownTimeout",
			MaxConcurrency: 1,
			MaxCapacity:    1,
		})
		DefaultScheduler().RegisterExecutorFactory(
			NewExecutorFactory("TestTaskGracefulShutdownTimeout", func(context TaskContext) {}))

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:                      "TestTaskGracefulShutdownTimeout-1",
			Command:                 "TestTaskGracefulShutdownTimeout",
			GoroutinePool:           "TestTaskGracefulShutdownTimeout",
			GracefulShutdownTimeout: oneHundredMs,
		}), convey.ShouldBeNil)

		ctx, err := DefaultScheduler().GetTaskContext("TestTaskGracefulShutdownTimeout-1")
		convey.So(err, convey.ShouldBeNil)
		ctx.Cancel()

		time.Sleep(twoHundredsMs)

		status, err := ctx.GetStatus()
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Phase, convey.ShouldEqual, Failed)
	})
}

func TestTaskHeartbeatTimeout(t *testing.T) {
	convey.Convey("test task heartbeat timeout", t, func() {
		DefaultScheduler().RegisterGoroutinePool(GoroutinePoolSpec{
			Id:             "TestTaskHeartbeatTimeout",
			MaxConcurrency: 1,
			MaxCapacity:    1,
		})
		DefaultScheduler().RegisterExecutorFactory(
			NewExecutorFactory("TestTaskHeartbeatTimeout", func(context TaskContext) {}))

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:               "TestTaskHeartbeatTimeout-1",
			Command:          "TestTaskHeartbeatTimeout",
			GoroutinePool:    "TestTaskHeartbeatTimeout",
			HeartbeatTimeout: oneHundredMs,
		}), convey.ShouldBeNil)

		ctx, err := DefaultScheduler().GetTaskContext("TestTaskHeartbeatTimeout-1")
		convey.So(err, convey.ShouldBeNil)
		ctx.Cancel()

		time.Sleep(twoHundredsMs)

		status, err := ctx.GetStatus()
		convey.So(err, convey.ShouldBeNil)
		convey.So(status.Phase, convey.ShouldEqual, Failed)
	})
}

func TestTaskUpdateStatus(t *testing.T) {
	convey.Convey("test task update status", t, func() {
		DefaultScheduler().RegisterGoroutinePool(GoroutinePoolSpec{
			Id:             "TestTaskUpdateStatus",
			MaxConcurrency: 1,
			MaxCapacity:    1,
		})
		DefaultScheduler().RegisterExecutorFactory(
			NewExecutorFactory("TestTaskUpdateStatus", func(context TaskContext) {}))

		convey.So(DefaultScheduler().SubmitTask(&TaskSpec{
			Id:                      "TestTaskUpdateStatus-1",
			Command:                 "TestTaskUpdateStatus",
			GoroutinePool:           "TestTaskUpdateStatus",
			GracefulShutdownTimeout: oneHundredMs,
		}), convey.ShouldBeNil)

		ctx, err := DefaultScheduler().GetTaskContext("TestTaskUpdateStatus-1")
		convey.So(err, convey.ShouldBeNil)

		time.Sleep(oneHundredMs)
		type testCase struct {
			arg     TaskStatus
			success bool
		}
		testCases := []testCase{
			{arg: TaskStatus{Phase: Progressing}, success: true},
			{arg: TaskStatus{Phase: Failed}, success: true},
			{arg: TaskStatus{Phase: Waiting}, success: false},
			{arg: TaskStatus{Phase: Succeed}, success: false},
		}
		for idx, tc := range testCases {
			assertion := convey.ShouldBeNil
			if !tc.success {
				assertion = convey.ShouldNotBeNil
			}
			fmt.Println(idx)
			convey.So(ctx.UpdateStatus(tc.arg), assertion)
		}
	})
}
