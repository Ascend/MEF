// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
