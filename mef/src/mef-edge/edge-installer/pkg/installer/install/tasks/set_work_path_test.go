// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for testing set work path task
package tasks

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
)

var setWorkPathTask = SetWorkPathTask{PathMgr: pathMgr}

func clearEnv(path string) {
	if err := fileutils.DeleteAllFileWithConfusion(path); err != nil {
		hwlog.RunLog.Errorf("clear env for test failed, error: %v", err)
		return
	}
}

func TestSetWorkPathTask(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(fileutils.SetParentPathPermission, nil)
	defer p.Reset()

	convey.Convey("set work path success", t, setWorkPathSuccess)
	convey.Convey("set work path failed, set root dir parent permission failed", t, setRootDirsParentPermFailed)
	convey.Convey("set work path failed, prepare install dir failed", t, prepareInstallDirFailed)
	convey.Convey("set work path failed, prepare log dir failed", t, prepareLogDirFailed)
}

func setWorkPathSuccess() {
	defer clearEnv(testDir)
	err := setWorkPathTask.Run()
	convey.So(err, convey.ShouldBeNil)
}

func setRootDirsParentPermFailed() {
	p1 := gomonkey.ApplyFuncReturn(fileutils.SetParentPathPermission, test.ErrTest)
	defer p1.Reset()

	err := setWorkPathTask.Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("set dir [%s] parent permission failed, error: %v",
		setWorkPathTask.PathMgr.GetInstallRootDir(), test.ErrTest))
}

func prepareInstallDirFailed() {
	convey.Convey("get software abs dir failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&pathmgr.WorkPathMgr{}, "GetWorkAbsDir", "", test.ErrTest)
		defer p1.Reset()

		err := setWorkPathTask.Run()
		expectErr := fmt.Errorf("get software abs dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("create dir failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&pathmgr.WorkPathMgr{}, "GetWorkAbsDir", softwareDir, nil).
			ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p1.Reset()

		err := setWorkPathTask.Run()
		expectErr := fmt.Errorf("create dir [%s] failed, error: %v", softwareDir, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func prepareLogDirFailed() {
	dir := setWorkPathTask.PathMgr.GetEdgeLogDir()
	convey.Convey("create log dir failed", func() {
		p1 := gomonkey.ApplyFuncSeq(fileutils.CreateDir, []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}, Times: 2},
			{Values: gomonkey.Params{test.ErrTest}},
		})
		defer p1.Reset()

		err := setWorkPathTask.Run()
		expectErr := fmt.Errorf("create log dir [%s] failed, error: %v", dir, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set log dir permission failed", func() {
		p1 := gomonkey.ApplyFuncSeq(fileutils.CreateDir, []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}, Times: 3},
		}).
			ApplyFuncReturn(fileutils.SetPathPermission, test.ErrTest)
		defer p1.Reset()

		err := setWorkPathTask.Run()
		expectErr := fmt.Errorf("set log dir [%s] permission failed, error: %v", dir, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
