// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
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
