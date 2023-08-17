// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

import (
	"errors"
)

// common errors
var (
	ErrNoMoreChild           = errors.New("no more child")
	ErrNoRowsAffected        = errors.New("no rows affected")
	ErrNilPointer            = errors.New("nil pointer")
	ErrInvalidPhase          = errors.New("invalid phase")
	ErrTimeout               = errors.New("timeout")
	ErrTypeInvalid           = errors.New("invalid type")
	ErrTaskNotFound          = errors.New("no such task")
	ErrFactoryNotFound       = errors.New("no such factory")
	ErrGoroutinePoolNotFound = errors.New("no such goroutine pool")
	ErrFullQueue             = errors.New("full queue")
	ErrTooManyTask           = errors.New("too many tasks")
	ErrCancelled             = errors.New("cancelled")
)
