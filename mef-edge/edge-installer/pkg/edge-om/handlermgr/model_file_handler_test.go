// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr for testing model file handler
package handlermgr

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

var (
	modelFileMsg = &model.Message{}
	modelHandler = modelFileHandler{}
)

func TestModelFileHandler(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", nil)
	defer p.Reset()
	convey.Convey("test model file handler, test delete tmp model file", t, testDeleteTmpModelFile)
	convey.Convey("test model file handler, test delete all model file", t, testDeletePodAllModelFile)
	convey.Convey("test model file handler, test delete part model file", t, testDeletePodPartModelFile)
	convey.Convey("test model file handler should be failed, parse parameters failed", t, parseParamFailed)
}

func getDeleteModelContent(typ string) types.ModelFileInfo {
	return types.ModelFileInfo{
		Operation: "delete",
		Target:    typ,
		Uuid:      "bdf3242b-aec1-4100-af91-afa2b8fde88a",
		ModelFiles: []types.ModelFile{
			{Name: "module.om"},
		},
	}
}

func testDeleteTmpModelFile() {
	err := modelFileMsg.FillContent(getDeleteModelContent(constants.TargetTypeTemp))
	convey.So(err, convey.ShouldBeNil)

	convey.Convey("deleteTmpModelFile success", func() {
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deleteTmpModelFile failed, check dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p1.Reset()
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deleteTmpModelFile failed, delete model file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p1.Reset()
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testDeletePodAllModelFile() {
	err := modelFileMsg.FillContent(getDeleteModelContent(constants.TargetTypeAll))
	convey.So(err, convey.ShouldBeNil)

	convey.Convey("deletePodAllModelFile success", func() {
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deletePodAllModelFile failed, check dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p1.Reset()
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deletePodAllModelFile failed, delete model file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p1.Reset()
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testDeletePodPartModelFile() {
	err := modelFileMsg.FillContent(getDeleteModelContent(""))
	convey.So(err, convey.ShouldBeNil)

	convey.Convey("deletePodPartModelFile success", func() {
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deletePodPartModelFile failed, check dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p1.Reset()
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("deletePodPartModelFile failed, delete model file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p1.Reset()
		err = modelHandler.Handle(modelFileMsg)
		convey.So(err, convey.ShouldBeNil)
	})
}

func parseParamFailed() {
	p1 := gomonkey.ApplyMethodReturn(&model.Message{}, "ParseContent", test.ErrTest)
	defer p1.Reset()
	err := modelHandler.Handle(modelFileMsg)
	convey.So(err, convey.ShouldResemble, errors.New("parse parameters failed"))
}
