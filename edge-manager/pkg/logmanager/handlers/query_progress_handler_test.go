// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers
package handlers

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/testutils"
	"huawei.com/mindxedge/base/common/taskschedule"
)

// TestQueryProgressHandle tests queryProgressHandler
func TestQueryProgressHandle(t *testing.T) {
	convey.Convey("test queryProgressHandler.Handle", t, func() {
		dummyObjs := testutils.DummyTaskSchedule()
		taskTree := taskschedule.TaskTreeNode{Current: &taskschedule.Task{}}
		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, dummyObjs.Scheduler).
			ApplyMethodReturn(dummyObjs.TaskCtx, "GetSubTaskTree", taskTree, nil).
			ApplyFuncReturn(modulemgr.SendMessage, nil)
		defer patch.Reset()

		var handler queryProgressHandler
		err := handler.Handle(&model.Message{Content: constants.DumpMultiNodesLogTaskName + `.abc`})
		convey.So(err, convey.ShouldBeNil)
	})
}
