// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
	return nil
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
