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
