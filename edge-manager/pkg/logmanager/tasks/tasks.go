// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package tasks
package tasks

import (
	"time"

	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
)

const (
	dumpMultiNodesLogTaskExecuteTimeout          = 2 * time.Hour
	dumpMultiNodesLogTaskGracefulShutdownTimeout = 5 * time.Minute
	dumpSingleNodeLogTaskConcurrency             = 10
	dumpSingleNodeLogTaskCapacity                = 90

	paramNameNodeSerialNumbers = "nodeSerialNumbers"
)

// InitTasks init goroutine pools and tasks
func InitTasks() error {
	taskschedule.DefaultScheduler().RegisterExecutorFactory(
		taskschedule.NewExecutorFactory(constants.DumpMultiNodesLogTaskName, doDumpMultiNodesLog))
	taskschedule.DefaultScheduler().RegisterExecutorFactory(
		taskschedule.NewExecutorFactory(constants.DumpSingleNodeLogTaskName, doDumpSingleNodeLog))
	taskschedule.DefaultScheduler().RegisterGoroutinePool(taskschedule.GoroutinePoolSpec{
		Id:             constants.DumpMultiNodesLogTaskName,
		MaxConcurrency: 1,
		MaxCapacity:    0,
	})
	taskschedule.DefaultScheduler().RegisterGoroutinePool(taskschedule.GoroutinePoolSpec{
		Id:             constants.DumpSingleNodeLogTaskName,
		MaxConcurrency: dumpSingleNodeLogTaskConcurrency,
		MaxCapacity:    dumpSingleNodeLogTaskCapacity,
	})
	return nil
}

// SubmitLogDumpTask submit log dump task, return task id
func SubmitLogDumpTask(edgeNodes []string) (string, error) {
	masterTask := &taskschedule.TaskSpec{
		Name:                    constants.DumpMultiNodesLogTaskName,
		GoroutinePool:           constants.DumpMultiNodesLogTaskName,
		Command:                 constants.DumpMultiNodesLogTaskName,
		Args:                    map[string]interface{}{paramNameNodeSerialNumbers: edgeNodes},
		ExecuteTimeout:          dumpMultiNodesLogTaskExecuteTimeout,
		GracefulShutdownTimeout: dumpMultiNodesLogTaskGracefulShutdownTimeout,
	}
	if err := taskschedule.DefaultScheduler().SubmitTask(masterTask); err != nil {
		return "", err
	}
	return masterTask.Id, nil
}
