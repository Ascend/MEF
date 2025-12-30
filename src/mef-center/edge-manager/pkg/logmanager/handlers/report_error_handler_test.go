// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
		err := handler.Handle(&model.Message{Content: []byte(`{"id":"abc"}`)})
		convey.So(err, convey.ShouldBeNil)
	})
}
