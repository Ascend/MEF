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
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/logmanager/testutils"
	"edge-manager/pkg/logmanager/utils"
)

// TestDumpEdgeLogs tests dumpEdgeLogs
func TestDumpEdgeLogs(t *testing.T) {
	convey.Convey("test dumpEdgeLogs", t, func() {
		dummyObjs := testutils.DummyTaskSchedule()
		outputs := []gomonkey.OutputCell{
			{Values: []interface{}{dummyObjs.TaskCtx, nil}, Times: 1},
			{Values: []interface{}{nil, taskschedule.ErrNoRunningSubTask}, Times: 1},
		}
		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, dummyObjs.Scheduler).
			ApplyMethodReturn(dummyObjs.Scheduler, "SubmitTask", nil).
			ApplyMethodReturn(dummyObjs.TaskCtx, "Spec", taskschedule.TaskSpec{Id: "1"}).
			ApplyMethodReturn(dummyObjs.TaskCtx, "GracefulShutdown", nil).
			ApplyMethodReturn(dummyObjs.TaskCtx, "GetStatus", succeedStatus, nil).
			ApplyMethodReturn(dummyObjs.TaskCtx, "UpdateStatus", nil).
			ApplyMethodSeq(dummyObjs.SubTaskSelector, "Select", outputs)
		defer patch.Reset()

		tasks, err := dumpEdgeLogs(dummyObjs.TaskCtx, []string{"1"}, []string{"1"}, []uint64{1})
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(tasks), convey.ShouldEqual, 1)
	})
}

// TestDumpMultiNodesLog tests dumpMultiNodesLog
func TestDumpMultiNodesLog(t *testing.T) {
	mockCreateTarGz := func(ctx taskschedule.TaskContext, subTasks []taskschedule.Task) error {
		if err := os.MkdirAll(filepath.Dir(edgeNodesLogTempPath), common.Mode600); err != nil {
			return err
		}
		file, err := os.OpenFile(edgeNodesLogTempPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, common.Mode600)
		if err != nil {
			return err
		}
		return file.Close()
	}

	convey.Convey("test dump multiNodesLog", t, func() {

		taskSpec := taskschedule.TaskSpec{Args: map[string]interface{}{
			paramNameNodeSerialNumbers: []string{"1"},
			paramNameNodeIps:           []string{"1"},
			paramNameNodeIDs:           []uint64{1}}}
		dummyObjs := testutils.DummyTaskSchedule()
		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, dummyObjs.Scheduler).
			ApplyMethodReturn(dummyObjs.TaskCtx, "UpdateStatus", nil).
			ApplyMethodReturn(dummyObjs.TaskCtx, "Spec", taskSpec).
			ApplyFuncReturn(dumpEdgeLogs, []taskschedule.Task{{}}, nil).
			ApplyFuncReturn(utils.CheckDiskSpace, nil).
			ApplyFuncReturn(utils.CleanTempFiles, false, nil).
			ApplyFuncReturn(fileutils.CreateDir, nil).
			ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
			ApplyFuncReturn(fileutils.RenameFile, nil).
			ApplyFunc(createTarGz, mockCreateTarGz)
		defer patch.Reset()

		err := dumpMultiNodesLog(dummyObjs.TaskCtx)
		convey.So(err, convey.ShouldBeNil)
	})
}
