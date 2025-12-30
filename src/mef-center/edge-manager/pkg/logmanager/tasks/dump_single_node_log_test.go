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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/testutils"
	"edge-manager/pkg/logmanager/utils"

	"huawei.com/mindxedge/base/common/taskschedule"
)

func TestDoDumpSingleNodeLog(t *testing.T) {
	taskSpec := taskschedule.TaskSpec{Args: map[string]interface{}{
		constants.NodeSnAndIp: "1", constants.NodeID: 1}}
	dummyValues := testutils.DummyTaskSchedule()

	convey.Convey("test dump edge logs", t, func() {
		okMsg := &model.Message{}
		err := okMsg.FillContent(common.OK)
		convey.So(err, convey.ShouldBeNil)

		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, dummyValues.Scheduler).
			ApplyMethodReturn(dummyValues.TaskCtx, "Spec", taskSpec).
			ApplyFunc(utils.FeedbackTaskError, func(taskCtx taskschedule.TaskContext, err error) {
				convey.So(err, convey.ShouldBeNil)
			}).
			ApplyFuncReturn(modulemgr.SendSyncMessage, okMsg, nil)
		defer patch.Reset()
		doDumpSingleNodeLog(dummyValues.TaskCtx)
	})
}
