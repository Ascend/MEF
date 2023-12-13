// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package taskschedule for package main test
package taskschedule

import (
	"context"
	"fmt"
	"testing"

	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcTaskSchedule := &TcTaskSchedule{
		tcBaseWithDb: &test.TcBaseWithDb{
			DbPath: ":memory:?cache=shared",
		},
	}
	test.RunWithPatches(tcTaskSchedule, m, nil)
}

// TcBase struct for test case
type TcTaskSchedule struct {
	tcBaseWithDb *test.TcBaseWithDb
}

// Setup pre-processing
func (tc *TcTaskSchedule) Setup() error {
	if err := tc.tcBaseWithDb.Setup(); err != nil {
		return err
	}
	rawDb, err := test.MockGetDb().DB()
	if err != nil {
		return fmt.Errorf("setup db failed, %v", err)
	}
	rawDb.SetMaxOpenConns(1)

	const thousand = 1000
	if err = InitDefaultScheduler(context.Background(), test.MockGetDb(), SchedulerSpec{
		MaxHistoryMasterTasks: thousand,
		MaxActiveTasks:        thousand,
		AllowedMaxTasksInDb:   thousand,
	}); err != nil {
		return fmt.Errorf("init sheduler failed, %v", err)
	}
	return nil
}

// Teardown post-processing
func (tc *TcTaskSchedule) Teardown() {
	tc.tcBaseWithDb.Teardown()
}
