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
