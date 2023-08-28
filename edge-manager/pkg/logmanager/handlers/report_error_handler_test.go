// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers
package handlers

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/logmanager/testutils"
	"huawei.com/mindxedge/base/common/taskschedule"
)

// TestReportErrorHandle tests the reportErrorHandler
func TestReportErrorHandle(t *testing.T) {
	convey.Convey("test reportErrorHandler.Handle", t, func() {
		dummyObjs := testutils.DummyTaskSchedule()
		patch := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil).
			ApplyFuncReturn(taskschedule.DefaultScheduler, dummyObjs.Scheduler).
			ApplyMethodReturn(dummyObjs.TaskCtx, "UpdateStatus", nil)
		defer patch.Reset()

		var handler reportErrorHandler
		err := handler.Handle(&model.Message{Content: `{"id":"abc"}`})
		convey.So(err, convey.ShouldBeNil)
	})
}
