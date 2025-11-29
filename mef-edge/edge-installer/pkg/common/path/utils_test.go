// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package path test for utils.go
package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

const testInstallRootDir = "./"

func TestGetInstallRootDir(t *testing.T) {
	convey.Convey("test func GetInstallRootDir success", t, func() {
		installRootDir, err := GetInstallRootDir()
		convey.So(installRootDir, convey.ShouldNotResemble, "")
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetInstallRootDir failed, get executable path failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()
		installRootDir, err := GetInstallRootDir()
		convey.So(installRootDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("get install dir failed, get current path failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetInstallDir(t *testing.T) {
	convey.Convey("test func GetInstallDir success", t, func() {
		installDir, err := GetInstallDir()
		convey.So(installDir, convey.ShouldNotResemble, "")
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetInstallDir failed, get executable path failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()
		installRootDir, err := GetInstallDir()
		convey.So(installRootDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("get current path failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetInstallDir failed, eval symlink failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()
		installRootDir, err := GetInstallDir()
		convey.So(installRootDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("eval install dir symlink failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetWorkPathMgr(t *testing.T) {
	convey.Convey("test func GetWorkPathMgr success", t, func() {
		workPathMgr, err := GetWorkPathMgr()
		convey.So(workPathMgr, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetWorkPathMgr failed, get executable path failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()
		workPathMgr, err := GetWorkPathMgr()
		convey.So(workPathMgr, convey.ShouldBeNil)
		expErr := fmt.Errorf("get install dir failed, get current path failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetConfigPathMgr(t *testing.T) {
	convey.Convey("test func GetConfigPathMgr success", t, func() {
		configPathMgr, err := GetConfigPathMgr()
		convey.So(configPathMgr, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetConfigPathMgr failed, get executable path failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()
		configPathMgr, err := GetConfigPathMgr()
		convey.So(configPathMgr, convey.ShouldBeNil)
		expErr := fmt.Errorf("get install dir failed, get current path failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetLogPathMgr(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(strings.LastIndex, 1)
	defer patches.Reset()

	convey.Convey("test func GetLogPathMgr success", t, func() {
		logPathMgr, err := GetLogPathMgr()
		convey.So(logPathMgr, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetLogPathMgr failed, GetInstallRootDir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()
		logPathMgr, err := GetLogPathMgr()
		convey.So(logPathMgr, convey.ShouldBeNil)
		expErr := fmt.Errorf("get install dir failed, get current path failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetLogPathMgr failed, GetLogRootDir or GetLogBackupRootDir failed", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{testInstallRootDir, nil}},
			{Values: gomonkey.Params{"", test.ErrTest}},

			{Values: gomonkey.Params{testInstallRootDir, nil}, Times: 2},
			{Values: gomonkey.Params{"", test.ErrTest}},
		}
		var p1 = gomonkey.ApplyFuncSeq(filepath.EvalSymlinks, outputs)
		defer p1.Reset()

		logPathMgr, err := GetLogPathMgr()
		convey.So(logPathMgr, convey.ShouldBeNil)
		expErr := errors.New("get log root dir failed")
		convey.So(err, convey.ShouldResemble, expErr)

		logPathMgr, err = GetLogPathMgr()
		convey.So(logPathMgr, convey.ShouldBeNil)
		expErr = errors.New("get log backup root dir failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetEdgeLogDirs(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(pathmgr.NewLogPathMgr,
		pathmgr.NewLogPathMgr(testInstallRootDir, testInstallRootDir)).
		ApplyFuncReturn(fileutils.IsSoftLink, nil).
		ApplyFuncReturn(strings.LastIndex, 1)
	defer patches.Reset()

	convey.Convey("test func GetEdgeLogDirs success", t, func() {
		edgeLogDir, edgeLogBackupDir, err := GetEdgeLogDirs()
		convey.So(edgeLogDir, convey.ShouldResemble, constants.MEFEdgeLogName)
		convey.So(edgeLogBackupDir, convey.ShouldResemble, constants.MEFEdgeLogBackupName)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetEdgeLogDirs failed, GetLogPathMgr failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()

		edgeLogDir, edgeLogBackupDir, err := GetEdgeLogDirs()
		convey.So(edgeLogDir, convey.ShouldResemble, "")
		convey.So(edgeLogBackupDir, convey.ShouldResemble, "")
		expErr := errors.New("get log path manager failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetEdgeLogDirs failed, GetEdgeLogDir or GetEdgeLogBackupDir failed", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{test.ErrTest}},

			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{test.ErrTest}},
		}
		var p1 = gomonkey.ApplyFuncSeq(fileutils.IsSoftLink, outputs).
			ApplyFuncReturn(pathmgr.NewLogPathMgr, pathmgr.NewLogPathMgr(testInstallRootDir, testInstallRootDir))
		defer p1.Reset()

		edgeLogDir, edgeLogBackupDir, err := GetEdgeLogDirs()
		convey.So(edgeLogDir, convey.ShouldResemble, "")
		convey.So(edgeLogBackupDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("edge log dir link check failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)

		edgeLogDir, edgeLogBackupDir, err = GetEdgeLogDirs()
		convey.So(edgeLogDir, convey.ShouldResemble, "")
		convey.So(edgeLogBackupDir, convey.ShouldResemble, "")
		expErr = fmt.Errorf("edge log backup dir link check failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetLogRootDir(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(strings.LastIndex, 1)
	defer patches.Reset()

	convey.Convey("test func GetLogRootDir success", t, func() {
		logRootDir, err := GetLogRootDir(testInstallRootDir)
		convey.So(logRootDir, convey.ShouldResemble, ".")
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetLogRootDir failed, filepath.EvalSymlinks failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()

		logRootDir, err := GetLogRootDir(testInstallRootDir)
		convey.So(logRootDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("eval log link dir symlink failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetLogRootDir failed, strings.LastIndex error", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(strings.LastIndex, -1)
		defer p1.Reset()

		logRootDir, err := GetLogRootDir(testInstallRootDir)
		convey.So(logRootDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("[%s] not found in path [%s]", constants.MEFEdgeLogName, testInstallRootDir)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetLogBackupRootDir(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(strings.LastIndex, 1)
	defer patches.Reset()

	convey.Convey("test func GetLogBackupRootDir success", t, func() {
		logBackupRootDir, err := GetLogBackupRootDir(testInstallRootDir)
		convey.So(logBackupRootDir, convey.ShouldResemble, ".")
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetLogBackupRootDir failed, filepath.EvalSymlinks failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()

		logBackupRootDir, err := GetLogBackupRootDir(testInstallRootDir)
		convey.So(logBackupRootDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("eval log backup link dir symlink failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetLogBackupRootDir failed, strings.LastIndex error", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(strings.LastIndex, -1)
		defer p1.Reset()

		logBackupRootDir, err := GetLogBackupRootDir(testInstallRootDir)
		convey.So(logBackupRootDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("[%s] not found in path [%s]", constants.MEFEdgeLogBackupName, testInstallRootDir)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetCompLogDirs(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(fileutils.ReadLink, testInstallRootDir, nil)
	defer patches.Reset()

	convey.Convey("test func GetCompLogDirs success", t, func() {
		installerLogDir, installerLogBackupDir, err := GetCompLogDirs(constants.EdgeInstaller)
		convey.So(installerLogDir, convey.ShouldResemble, testInstallRootDir)
		convey.So(installerLogBackupDir, convey.ShouldResemble, testInstallRootDir)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetCompLogDirs failed, filepath.EvalSymlinks failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()

		installerLogDir, installerLogBackupDir, err := GetCompLogDirs(constants.EdgeInstaller)
		convey.So(installerLogDir, convey.ShouldResemble, "")
		convey.So(installerLogBackupDir, convey.ShouldResemble, "")
		expErr := errors.New("get work path manager failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetCompLogDirs failed, readLink failed", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", test.ErrTest}},

			{Values: gomonkey.Params{testInstallRootDir, nil}},
			{Values: gomonkey.Params{"", test.ErrTest}},
		}
		var p1 = gomonkey.ApplyFuncSeq(fileutils.ReadLink, outputs)
		defer p1.Reset()

		installerLogDir, installerLogBackupDir, err := GetCompLogDirs(constants.EdgeInstaller)
		convey.So(installerLogDir, convey.ShouldResemble, "")
		convey.So(installerLogBackupDir, convey.ShouldResemble, "")
		expErr := fmt.Errorf("read compnonent log link failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)

		installerLogDir, installerLogBackupDir, err = GetCompLogDirs(constants.EdgeInstaller)
		convey.So(installerLogDir, convey.ShouldResemble, "")
		convey.So(installerLogBackupDir, convey.ShouldResemble, "")
		expErr = fmt.Errorf("read compnonent log backup link failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
