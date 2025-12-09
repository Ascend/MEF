// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build TESTCODE
// +build TESTCODE

// Package testutils
package testutils

import (
	"edge-manager/pkg/constants"

	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common/taskschedule"
)

type dummyScheduler struct {
	taskschedule.Scheduler
	taskCtx         taskschedule.TaskContext
	subTaskSelector taskschedule.SubTaskSelector
}

type dummyTaskCtx struct {
	taskschedule.TaskContext
}

type dummySubTaskSelector struct {
	taskschedule.SubTaskSelector
}

func (s *dummyScheduler) GetTaskContext(string) (taskschedule.TaskContext, error) {
	return s.taskCtx, nil
}

func (s *dummyScheduler) NewSubTaskSelector(string) taskschedule.SubTaskSelector {
	return s.subTaskSelector
}

func (c *dummyTaskCtx) Spec() taskschedule.TaskSpec {
	return taskschedule.TaskSpec{
		Args: map[string]interface{}{
			constants.PeerInfo: model.MsgPeerInfo{
				Ip: "test",
				Sn: "test",
			},
		},
	}
}

// DummyTaskScheduleObjects dummy schedule objects
type DummyTaskScheduleObjects struct {
	Scheduler       taskschedule.Scheduler
	TaskCtx         taskschedule.TaskContext
	SubTaskSelector taskschedule.SubTaskSelector
}

// DummyTaskSchedule dummy task schedule
func DummyTaskSchedule() *DummyTaskScheduleObjects {
	objects := &DummyTaskScheduleObjects{
		TaskCtx:         &dummyTaskCtx{},
		SubTaskSelector: &dummySubTaskSelector{},
	}
	objects.Scheduler = &dummyScheduler{
		taskCtx:         objects.TaskCtx,
		subTaskSelector: objects.SubTaskSelector,
	}
	return objects
}
