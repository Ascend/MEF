// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

// TaskExecutor task execution func
type TaskExecutor func(TaskContext)

// TaskExecutorFactory task executor factory
type TaskExecutorFactory interface {
	// GetID gets id
	GetID() string
	// CreateExecutor creates task executor
	CreateExecutor() TaskExecutor
}

type taskExecutorFactoryImpl struct {
	id       string
	executor TaskExecutor
}

func (c taskExecutorFactoryImpl) GetID() string {
	return c.id
}

func (c taskExecutorFactoryImpl) CreateExecutor() TaskExecutor {
	return c.executor
}

// NewExecutorFactory creates a basic executor factory
func NewExecutorFactory(id string, executor TaskExecutor) TaskExecutorFactory {
	return taskExecutorFactoryImpl{id: id, executor: executor}
}
