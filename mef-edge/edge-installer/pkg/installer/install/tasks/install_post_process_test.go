// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing post install process task
package tasks

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common/tasks"
)

var postProcess = PostInstallProcessTask{
	PostProcessBaseTask: tasks.PostProcessBaseTask{
		WorkPathMgr: pathMgr.SoftwarePathMgr.WorkPathMgr,
		LogPathMgr:  pathMgr.LogPathMgr,
	},
}

func TestPostInstallProcessTask(t *testing.T) {
	convey.Convey("install post process should be success", t, func() {
		p := gomonkey.ApplyMethodReturn(&tasks.PostProcessBaseTask{}, "CreateSoftwareSymlink", nil).
			ApplyMethodReturn(&tasks.PostProcessBaseTask{}, "UpdateMefServiceInfo", nil).
			ApplyMethodReturn(&tasks.PostProcessBaseTask{}, "SetSoftwareDirImmutable", nil)
		defer p.Reset()
		err := postProcess.Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("install post process should be failed, get software abs dir failed", t, func() {
		p := gomonkey.ApplyMethodReturn(&pathmgr.WorkPathMgr{}, "GetWorkAbsDir", "", test.ErrTest)
		defer p.Reset()
		err := postProcess.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
