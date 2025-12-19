// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package pathmgr test for utils.go
package pathmgr

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
)

type mockFileInfo struct {
}

func (fi *mockFileInfo) Name() string {
	return ""
}

func (fi *mockFileInfo) Size() int64 {
	return 0
}

func (fi *mockFileInfo) Mode() os.FileMode {
	const fileMode = 0600
	return fileMode
}

func (fi *mockFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fi *mockFileInfo) IsDir() bool {
	return false
}

func (fi *mockFileInfo) Sys() interface{} {
	return 0
}

func TestGetTargetInstallDir(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(os.Lstat, &mockFileInfo{}, nil).
		ApplyFuncReturn(filepath.EvalSymlinks, testInstallRootDir, nil)
	defer patches.Reset()

	convey.Convey("test func GetTargetInstallDir success, return work_A dir", t, func() {
		softwareDir, err := GetTargetInstallDir(testInstallRootDir)
		expRes := NewWorkPathMgr(testInstallRootDir).GetWorkADir()
		convey.So(softwareDir, convey.ShouldResemble, expRes)
		convey.So(err, convey.ShouldResemble, nil)

		var p1 = gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
		defer p1.Reset()
		softwareDir, err = GetTargetInstallDir(testInstallRootDir)
		convey.So(softwareDir, convey.ShouldResemble, expRes)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetTargetInstallDir success, return work_B dir", t, func() {
		workADir := NewWorkPathMgr(testInstallRootDir).GetWorkADir()
		var p1 = gomonkey.ApplyMethodReturn(&mockFileInfo{}, "Mode", os.ModeSymlink).
			ApplyFuncReturn(filepath.EvalSymlinks, workADir, nil)
		defer p1.Reset()

		softwareDir, err := GetTargetInstallDir(testInstallRootDir)
		expRes := NewWorkPathMgr(testInstallRootDir).GetWorkBDir()
		convey.So(softwareDir, convey.ShouldResemble, expRes)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetTargetInstallDir failed, os.Lstat failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Lstat, &mockFileInfo{}, test.ErrTest)
		defer p1.Reset()

		softwareDir, err := GetTargetInstallDir(testInstallRootDir)
		convey.So(softwareDir, convey.ShouldResemble, "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func GetTargetInstallDir failed, filepath.EvalSymlinks failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()

		softwareDir, err := GetTargetInstallDir(testInstallRootDir)
		convey.So(softwareDir, convey.ShouldResemble, "")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
