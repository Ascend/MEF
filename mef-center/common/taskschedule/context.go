// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package taskschedule
package taskschedule

import (
	"context"
)

// TaskContext task context
type TaskContext interface {
	context.Context
	// Spec returns task specification
	Spec() TaskSpec
	// GracefulShutdown returns graceful shutdown channel
	GracefulShutdown() <-chan struct{}
	// UpdateLiveness updates task liveness
	UpdateLiveness() error
	// UpdateStatus updates task status
	UpdateStatus(result TaskStatus) error
	// GetStatus gets task status
	GetStatus() (TaskStatus, error)
	// GetSubTaskTree gets sub task tree
	GetSubTaskTree() (TaskTreeNode, error)
	// Cancel cancels task
	Cancel()
}
