// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package common for test utils function
package common

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

func clearEnv(path string) {
	if err := fileutils.DeleteAllFileWithConfusion(path); err != nil {
		hwlog.RunLog.Errorf("clear env for test failed, error: %v", err)
		return
	}
}

func TestCheckLogDirs(t *testing.T) {
	dirTest := "/tmp/test_check_log_dirs"
	defer clearEnv(dirTest)

	convey.Convey("check log dirs success", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", nil)
		defer p.Reset()
		err := CheckLogDirs(dirTest, dirTest, true)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("check log dirs failed", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p.Reset()
		err := CheckLogDirs(dirTest, dirTest, true)
		expectErr := fmt.Errorf("check dir [%s] failed, error: %v", constants.LogDirName, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func TestCheckDir(t *testing.T) {
	dirTest := "/tmp/test_check_dir"
	defer clearEnv(dirTest)

	convey.Convey("check dir success", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", nil)
		defer p.Reset()
		err := CheckDir(dirTest, "")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("check dir failed, create dir failed", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p.Reset()
		err := CheckDir(dirTest, constants.LogDirName)
		expectErr := fmt.Errorf("create dir [%s] failed, error: %v", constants.LogDirName, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check dir failed, dir does not exist", t, func() {
		p := gomonkey.ApplyFuncReturn(utils.IsFlagSet, true)
		defer p.Reset()
		notExistDir := ""
		err := CheckDir(notExistDir, "")
		expectErr := fmt.Errorf("dir [%s] does not exist", notExistDir)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check dir failed, dir is not absolute path", t, func() {
		notAbsDir := "../"
		err := CheckDir(notAbsDir, constants.LogDirName)
		expectErr := fmt.Errorf("dir [%s] is not absolute path", constants.LogDirName)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check dir failed, dir cannot be in the decompression path", t, func() {
		decompressDir := constants.UnpackPath
		err := CheckDir(decompressDir, constants.LogDirName)
		expectErr := fmt.Errorf("dir [%s] cannot be in the decompression path", constants.LogDirName)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check dir failed, real dir check failed", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p.Reset()
		err := CheckDir(dirTest, constants.LogDirName)
		expectErr := fmt.Errorf("check dir [%s] failed, error: %v", constants.LogDirName, test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func TestCheckInTmpfs(t *testing.T) {
	convey.Convey("check in tmpfs success", t, func() {
		convey.Convey("the path is not in the temporary file system", func() {
			p := gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, false, nil)
			defer p.Reset()
			err := CheckInTmpfs("", false)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("the path is allowed in the temporary file system", func() {
			p := gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, true, nil)
			defer p.Reset()
			err := CheckInTmpfs("", true)
			convey.So(err, convey.ShouldBeNil)
		})
	})

	convey.Convey("check in tmpfs failed", t, func() {
		convey.Convey("IsInTmpfs failed", func() {
			p := gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, false, test.ErrTest)
			defer p.Reset()
			err := CheckInTmpfs("", false)
			convey.So(err, convey.ShouldResemble, test.ErrTest)
		})

		convey.Convey("the path cannot be in the tmpfs filesystem", func() {
			p := gomonkey.ApplyFuncReturn(envutils.IsInTmpfs, true, nil)
			defer p.Reset()
			dirTest := ""
			err := CheckInTmpfs(dirTest, false)
			expectErr := errors.New("the dir cannot be in the tmpfs filesystem")
			convey.So(err, convey.ShouldResemble, expectErr)
		})
	})
}

func TestCopyResetScriptToP7(t *testing.T) {
	output := `Filesystem      Size  Used Avail Use% Mounted on
/dev/mmcblk0p5  974M  154M  753M  17% /home/log
/dev/mmcblk0p7  2.9G  1.5G  1.3G  53% /home/package
/dev/mmcblk0p6  2.0G  248M  1.6G  14% /usr/local/mindx`

	var p = gomonkey.ApplyFuncReturn(envutils.RunCommand, output, nil).
		ApplyFuncReturn(path.GetCompWorkDir, "", nil).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(fileutils.DeleteFile, nil).
		ApplyFuncReturn(fileutils.CopyFile, nil)
	defer p.Reset()

	convey.Convey("copy reset script to p7 should be success", t, testCopyResetScriptToP7)
	convey.Convey("copy reset script to p7 should be failed, run command error", t, testCopyResetScriptToP7ErrRunCommand)
	convey.Convey("copy reset script to p7 should be failed, get comp work dir error", t, testCopyResetScriptToP7ErrGetDir)
	convey.Convey("copy reset script to p7 should be failed, firmware does not exist", t, testCopyResetScriptToP7ErrExist)
	convey.Convey("copy reset script to p7 should be failed, delete error", t, testCopyResetScriptToP7ErrDelete)
	convey.Convey("copy reset script to p7 should be failed, copy error", t, testCopyResetScriptToP7ErrCopy)
}

func testCopyResetScriptToP7() {
	err := CopyResetScriptToP7()
	convey.So(err, convey.ShouldBeNil)
}

func testCopyResetScriptToP7ErrRunCommand() {
	var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", testErr)
	defer p1.Reset()

	err := CopyResetScriptToP7()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("execute [%s] command failed, error: %v", dfCmd, testErr))
}

func testCopyResetScriptToP7ErrGetDir() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetCompWorkDir, "", testErr)
	defer p1.Reset()

	err := CopyResetScriptToP7()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get component work dir failed, error: %v", testErr))
}

func testCopyResetScriptToP7ErrExist() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
	defer p1.Reset()

	firmwarePath := filepath.Join("/home/package", firmwareDir)
	err := CopyResetScriptToP7()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("firmware path [%s] does not exist", firmwarePath))
}

func testCopyResetScriptToP7ErrDelete() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.DeleteFile, testErr)
	defer p1.Reset()

	firmwarePath := filepath.Join("/home/package", firmwareDir)
	resetScriptDst := filepath.Join(firmwarePath, constants.ResetMiddlewareScript)
	err := CopyResetScriptToP7()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("delete existed [%s] failed, error: %v", resetScriptDst, testErr))
}

func testCopyResetScriptToP7ErrCopy() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CopyFile, testErr)
	defer p1.Reset()

	firmwarePath := filepath.Join("/home/package", firmwareDir)
	resetScriptSrc := filepath.Join("", constants.Script, constants.ResetMiddlewareScript)
	resetScriptDst := filepath.Join(firmwarePath, constants.ResetMiddlewareScript)
	err := CopyResetScriptToP7()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("copy [%s] to [%s] failed, "+
		"error: %v", resetScriptSrc, resetScriptDst, testErr))
}
