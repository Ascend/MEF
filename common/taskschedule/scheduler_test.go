// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
)

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	m.Run()
}

func setup() error {
	const (
		memoryDsn = ":memory:?cache=shared"
		thousand  = 1000
	)
	if err := hwlog.InitRunLogger(&hwlog.LogConfig{OnlyToStdout: true}, context.Background()); err != nil {
		return fmt.Errorf("init hwlog failed, %v", err)
	}
	db, err := gorm.Open(sqlite.Open(memoryDsn))
	if err != nil {
		return fmt.Errorf("open db failed, %v", err)
	}
	rawDb, err := db.DB()
	if err != nil {
		return fmt.Errorf("setup db failed, %v", err)
	}
	rawDb.SetMaxOpenConns(1)
	if err := InitDefaultScheduler(context.Background(), db, SchedulerSpec{
		MaxHistoryMasterTasks: thousand,
		MaxActiveTasks:        thousand,
	}); err != nil {
		return fmt.Errorf("init sheduler failed, %v", err)
	}
	return nil
}

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
		convey.So(status.Phase, convey.ShouldEqual, Progressing)

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
