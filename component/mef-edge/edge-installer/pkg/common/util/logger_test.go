// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

func TestInitHwLogger(t *testing.T) {
	logConfig := NewLogConfig("nil", "nil")
	logConfig = &hwlog.LogConfig{OnlyToStdout: true}
	convey.Convey("TestInitHwLogger", t, func() {
		err := InitHwLogger(logConfig, logConfig)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func InitHwLogger failed, init run logger failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(hwlog.InitRunLogger, test.ErrTest)
		defer p1.Reset()
		err := InitHwLogger(logConfig, logConfig)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test func InitHwLogger failed, init operate logger failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(hwlog.InitOperateLogger, test.ErrTest)
		defer p1.Reset()
		err := InitHwLogger(logConfig, logConfig)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestMakeSureLogDir(t *testing.T) {
	logDir := CreateLogflie(t)
	defer func() {
		if err := os.RemoveAll(logDir); err != nil {
			t.Fatal(err)
		}
	}()
	convey.Convey("TestMakeSureLogDir", t, func() {
		err := MakeSureLogDir(logDir)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func MakeSureLogDir failed, create edge log dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p1.Reset()
		err := MakeSureLogDir(logDir)
		expErr := fmt.Errorf("create edge log dir [%s] failed, error: %v", filepath.Dir(logDir), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func MakeSureLogDir failed, check edge log dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p1.Reset()
		err := MakeSureLogDir(logDir)
		expErr := fmt.Errorf("check edge log dir [%s] failed, error: %v", filepath.Dir(logDir), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func MakeSureLogDir failed, create component log dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncSeq(fileutils.CreateDir, []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{test.ErrTest}},
		})
		defer p1.Reset()
		err := MakeSureLogDir(logDir)
		expErr := fmt.Errorf("create component log dir [%s] failed, error: %v", logDir, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestInitComponentLog(t *testing.T) {
	const compLogDir = "/tmp/log"
	const compLogBackupDir = "/tmp/log_backup"
	patches := gomonkey.ApplyFuncReturn(path.GetCompLogDirs, compLogDir, compLogBackupDir, nil).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil)
	defer patches.Reset()
	convey.Convey("test func InitComponentLog success", t, func() {
		err := InitComponentLog(constants.EdgeInstaller)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("test func InitComponentLog failed, make sure comp log dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p1.Reset()
		err := InitComponentLog(constants.EdgeInstaller)
		expErr := fmt.Errorf("create edge log dir [%s] failed, error: %v", filepath.Dir(compLogDir), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func InitComponentLog failed, make sure comp backup log dir failed", t, func() {
		var p1 = gomonkey.ApplyFuncSeq(fileutils.RealDirCheck, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", test.ErrTest}},
		})
		defer p1.Reset()
		err := InitComponentLog(constants.EdgeInstaller)
		expErr := fmt.Errorf("check edge log dir [%s] failed, error: %v", filepath.Dir(compLogBackupDir), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
