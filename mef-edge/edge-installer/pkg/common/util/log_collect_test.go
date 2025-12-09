// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/test"
)

func TestGetLogCollector(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Geteuid()), nil)
	defer patches.Reset()

	tempDir, err := os.MkdirTemp("", "temp")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatal(err)
		}
	}()

	convey.Convey("TestGetLogCollector", t, func() {
		GetLogCollector("~", "~", "~", []string{"~"})
		convey.So(checkLogFile(tempDir), convey.ShouldResemble, nil)
	})
}

type mockFileInfoErrSize struct {
}

func (fi *mockFileInfoErrSize) Name() string {
	return ""
}

func (fi *mockFileInfoErrSize) Size() int64 {
	return maxFileSize + 1
}

func (fi *mockFileInfoErrSize) Mode() os.FileMode {
	const fileMode = 0600
	return fileMode
}

func (fi *mockFileInfoErrSize) ModTime() time.Time {
	return time.Time{}
}

func (fi *mockFileInfoErrSize) IsDir() bool {
	return false
}

func (fi *mockFileInfoErrSize) Sys() interface{} {
	return 0
}

func TestCheckLogFile(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(os.Stat, &mockFileInfo{}, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Geteuid()), nil)
	defer patches.Reset()

	convey.Convey("test func checkLogFile failed, os.Stat failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Stat, &mockFileInfo{}, test.ErrTest)
		defer p1.Reset()
		convey.So(checkLogFile(""), convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkLogFile failed, file size error", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Stat, &mockFileInfoErrSize{}, nil)
		defer p1.Reset()
		err := checkLogFile("")
		expErr := errors.New("log file is too large")
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func checkLogFile failed, get uid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), test.ErrTest)
		defer p1.Reset()
		err := checkLogFile("")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkLogFile failed, get gid failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.GetGid, uint32(0), test.ErrTest)
		defer p1.Reset()
		err := checkLogFile("")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkLogFile failed, eval symlink failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()
		err := checkLogFile("")
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func checkLogFile failed, is symlink", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "/tmp", nil)
		defer p1.Reset()
		err := checkLogFile("")
		expErr := errors.New("symlink is not allowed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
