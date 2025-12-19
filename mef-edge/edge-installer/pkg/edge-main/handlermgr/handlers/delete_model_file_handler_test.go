// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlers for testing delete model file handler
package handlers

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
)

var deleteModelHandler = deleteModelFileHandler{}

func TestDeleteModelFileHandler(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(modeltask.GetModelMgr, &modeltask.ModelMgr{}).
		ApplyMethodReturn(&modeltask.ModelMgr{}, "LockGlobal", true).
		ApplyMethod(&modeltask.ModelMgr{}, "UnLockGlobal", func(*modeltask.ModelMgr) {}).
		ApplyMethod(&modeltask.ModelMgr{}, "CancelTasks", func(*modeltask.ModelMgr) {}).
		ApplyMethod(&modeltask.ModelMgr{}, "Clear", func(*modeltask.ModelMgr) {}).
		ApplyFuncReturn(util.SendSyncMsg, constants.Success, nil).
		ApplyFuncReturn(modulemgr.SendAsyncMessage, nil)
	defer p.Reset()
	convey.Convey("test delete model file handler successful", t, deleteModelFileHandlerSuccess)
	convey.Convey("test delete model file handler failed", t, deleteModelFileHandlerFailed)
}

func deleteModelFileHandlerSuccess() {
	err := deleteModelHandler.Handle(&model.Message{})
	convey.So(err, convey.ShouldBeNil)
}

func deleteModelFileHandlerFailed() {
	convey.Convey("operation is locked", func() {
		p1 := gomonkey.ApplyMethodReturn(&modeltask.ModelMgr{}, "LockGlobal", false)
		defer p1.Reset()
		err := deleteModelHandler.Handle(&model.Message{})
		convey.So(err, convey.ShouldResemble, errors.New("cannot perform this operation, other operation is working"))
	})

	convey.Convey("send pods_data message to edge om failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, constants.Failed, test.ErrTest)
		defer p1.Reset()
		err := deleteModelHandler.Handle(&model.Message{})
		convey.So(err, convey.ShouldResemble, errors.New("send pods_data message to edge om failed"))
	})

	convey.Convey("delete all model file failed by edge om", func() {
		p1 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, constants.Failed, nil)
		defer p1.Reset()
		err := deleteModelHandler.Handle(&model.Message{})
		convey.So(err, convey.ShouldResemble, errors.New("delete all model file failed by edge om"))
	})
}
