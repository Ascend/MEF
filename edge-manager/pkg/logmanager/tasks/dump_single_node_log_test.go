// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
