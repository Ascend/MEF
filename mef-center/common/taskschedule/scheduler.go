// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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
	"context"
	"sync"

	"gorm.io/gorm"
)

var (
	defaultScheduler     Scheduler
	defaultSchedulerOnce sync.Once
)

// InitDefaultScheduler init the default scheduler
func InitDefaultScheduler(ctx context.Context, db *gorm.DB, spec SchedulerSpec) error {
	var (
		err       error
		scheduler Scheduler
	)
	defaultSchedulerOnce.Do(func() {
		scheduler, err = startScheduler(ctx, db, spec)
		if err == nil {
			defaultScheduler = scheduler
		}
	})
	return err
}

// DefaultScheduler returns the default scheduler
func DefaultScheduler() Scheduler {
	return defaultScheduler
}

// SubTaskSelector selector for
type SubTaskSelector interface {
	Select(cancel ...<-chan struct{}) (TaskContext, error)
}

// Scheduler provides functionality of task schedule and management
type Scheduler interface {
	// RegisterExecutorFactory registers executor factory
	RegisterExecutorFactory(factory TaskExecutorFactory) bool
	// RegisterGoroutinePool registers go routine pool
	RegisterGoroutinePool(pool GoroutinePoolSpec) bool
	// SubmitTask submits master task
	SubmitTask(task *TaskSpec) error
	// GetTaskContext gets task context
	GetTaskContext(taskId string) (TaskContext, error)
	// NewSubTaskSelector returns a selector for subtasks
	NewSubTaskSelector(taskId string) SubTaskSelector
}
