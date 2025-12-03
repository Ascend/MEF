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
	"errors"
)

// common errors
var (
	ErrNoRunningSubTask      = errors.New("no running sub task")
	ErrNoRowsAffected        = errors.New("no rows is affected")
	ErrNilPointer            = errors.New("nil pointer")
	ErrTaskAlreadyFinished   = errors.New("task already finished")
	ErrTimeout               = errors.New("timeout")
	ErrTypeInvalid           = errors.New("invalid type")
	ErrTaskNotFound          = errors.New("no such task")
	ErrFactoryNotFound       = errors.New("no such factory")
	ErrGoroutinePoolNotFound = errors.New("no such goroutine pool")
	ErrFullQueue             = errors.New("task queue is full")
	ErrTooManyTask           = errors.New("too many tasks")
	ErrCancelled             = errors.New("cancelled")
)
