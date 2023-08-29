// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
