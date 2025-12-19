// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
	paramNameNodeIDs           = "nodeIDs"
	paramNameNodeIps           = "nodeIps"
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
func SubmitLogDumpTask(edgeNodeSNs []string, edgeIps []string, edgeNodeIDs []uint64) (string, error) {
	masterTask := &taskschedule.TaskSpec{
		Name:          constants.DumpMultiNodesLogTaskName,
		GoroutinePool: constants.DumpMultiNodesLogTaskName,
		Command:       constants.DumpMultiNodesLogTaskName,
		Args: map[string]interface{}{
			paramNameNodeSerialNumbers: edgeNodeSNs,
			paramNameNodeIDs:           edgeNodeIDs,
			paramNameNodeIps:           edgeIps,
		},
		ExecuteTimeout:          dumpMultiNodesLogTaskExecuteTimeout,
		GracefulShutdownTimeout: dumpMultiNodesLogTaskGracefulShutdownTimeout,
	}
	if err := taskschedule.DefaultScheduler().SubmitTask(masterTask); err != nil {
		return "", err
	}
	return masterTask.Id, nil
}
