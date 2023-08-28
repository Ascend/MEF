// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package tasks
package tasks

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/testutils"
	"edge-manager/pkg/logmanager/utils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"
)

// TestMain sets up the environment
func TestMain(m *testing.M) {
	if err := testutils.PrepareHwlog(); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
	}
	if err := testutils.PrepareTempDirs(); err != nil {
		fmt.Printf("prepare dirs failed, %v\n", err)
	}
	m.Run()
	if err := testutils.CleanupTempDirs(); err != nil {
		fmt.Printf("cleanup dirs failed, %v\n", err)
	}
}

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

		tasks, err := dumpEdgeLogs(dummyObjs.TaskCtx, []string{"1"})
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(tasks), convey.ShouldEqual, 1)
	})
}

// TestCreateTarGz tests createTarGz
func TestCreateTarGz(t *testing.T) {
	convey.Convey("test createTarGz", t, func() {
		dummyObjs := testutils.DummyTaskSchedule()
		tasks := []taskschedule.Task{
			{Spec: taskschedule.TaskSpec{Id: "1", Args: map[string]interface{}{constants.NodeSerialNumber: "1"}}},
			{Spec: taskschedule.TaskSpec{Id: "2", Args: map[string]interface{}{constants.NodeSerialNumber: "2"}}},
		}
		for _, task := range tasks {
			filePath := filepath.Join(constants.LogDumpTempDir, task.Spec.Id+common.TarGzSuffix)
			pkgFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, common.Mode600)
			convey.So(err, convey.ShouldBeNil)
			err = pkgFile.Close()
			convey.So(err, convey.ShouldBeNil)
		}

		patch := gomonkey.
			ApplyFuncReturn(taskschedule.DefaultScheduler, dummyObjs.Scheduler).
			ApplyMethodReturn(dummyObjs.TaskCtx, "UpdateStatus", nil).
			ApplyFunc(utils.WithDiskPressureProtect, testutils.WithoutDiskPressureProtect)
		defer patch.Reset()
		err := createTarGz(dummyObjs.TaskCtx, tasks)
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestDumpMultiNodesLog tests dumpMultiNodesLog
func TestDumpMultiNodesLog(t *testing.T) {
	mockCreateTarGz := func(ctx taskschedule.TaskContext, subTasks []taskschedule.Task) error {
		file, err := os.OpenFile(edgeNodesLogTempPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, common.Mode600)
		if err != nil {
			return err
		}
		return file.Close()
	}

	convey.Convey("test dump multiNodesLog", t, func() {
		taskSpec := taskschedule.TaskSpec{Args: map[string]interface{}{paramNameNodeSerialNumbers: []string{"1"}}}
		dummyObjs := testutils.DummyTaskSchedule()
		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, dummyObjs.Scheduler).
			ApplyMethodReturn(dummyObjs.TaskCtx, "UpdateStatus", nil).
			ApplyMethodReturn(dummyObjs.TaskCtx, "Spec", taskSpec).
			ApplyFuncReturn(dumpEdgeLogs, []taskschedule.Task{{}}, nil).
			ApplyFuncReturn(envutils.CheckDiskSpace, nil).
			ApplyFunc(createTarGz, mockCreateTarGz)
		defer patch.Reset()

		err := dumpMultiNodesLog(dummyObjs.TaskCtx)
		convey.So(err, convey.ShouldBeNil)
	})
}
