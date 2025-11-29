// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common test for log permission
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

	"edge-installer/pkg/common/constants"
)

var (
	testLogPermDir = "/tmp/test_log_perm_mgr"
	logPermMgr     = LogPermissionMgr{LogPath: testLogPermDir}
)

func setupTestLogPermMgr() error {
	if err := prepareNecessaryLogPath(); err != nil {
		hwlog.RunLog.Errorf("prepare necessary log path failed, error: %v", err)
		return err
	}
	return nil
}

func teardownTestLogPermMgr() {
	if err := fileutils.DeleteAllFileWithConfusion(testLogPermDir); err != nil {
		fmt.Printf("cleanup [%s] failed, error: %v\n", testLogPermDir, err)
	}
}

func TestCheckPermission(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(constants.EdgeUserUid), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(constants.EdgeUserGid), nil)
	defer p.Reset()

	if err := setupTestLogPermMgr(); err != nil {
		panic(err)
	}
	defer teardownTestLogPermMgr()

	convey.Convey("check permission should be success", t, testCheckPermission)
	convey.Convey("check permission should be failed, check path failed", t, testCheckPermissionErrCheckPath)
}

func testCheckPermission() {
	err := logPermMgr.CheckPermission()
	convey.So(err, convey.ShouldBeNil)
}

func testCheckPermissionErrCheckPath() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{"", testErr}},

		{Values: gomonkey.Params{"", nil}, Times: 2},
		{Values: gomonkey.Params{"", testErr}},
	}
	var p1 = gomonkey.ApplyFuncSeq(fileutils.CheckOriginPath, outputs)
	defer p1.Reset()

	logPermList := logPermMgr.GetLogPermList()
	err := logPermMgr.CheckPermission()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("check path [%s] failed, error: %v", logPermList[0].dir, testErr))

	logFile := fmt.Sprintf("%s_%s", filepath.Base(logPermList[1].dir), constants.RunLogFile)
	logFilePath := filepath.Join(logPermList[1].dir, logFile)
	err = logPermMgr.CheckPermission()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("check path [%s] failed, error: %v", logFilePath, testErr))
}

func prepareNecessaryLogPath() error {
	if err := fileutils.DeleteAllFileWithConfusion(testLogPermDir); err != nil {
		hwlog.RunLog.Errorf("cleanup [%s] failed, error: %v", testLogPermDir, err)
		return fmt.Errorf("cleanup [%s] failed", testLogPermDir)
	}

	if err := fileutils.CreateDir(testLogPermDir, constants.Mode755); err != nil {
		hwlog.RunLog.Errorf("create test log dir [%s] failed, error: %v", testLogPermDir, err)
		return fmt.Errorf("create tets log dir [%s] failed", testLogPermDir)
	}

	edgeMainPerm, err := logPermMgr.getEdgeMainPerm()
	if err != nil {
		hwlog.RunLog.Errorf("get edge main perm failed, error: %v", err)
		return errors.New("get edge main perm failed")
	}
	logPermList := []LogPerm{
		logPermMgr.getEdgeInstallerPerm(),
		logPermMgr.getEdgeOmPerm(),
		logPermMgr.getEdgeCorePerm(),
		edgeMainPerm,
	}
	for _, logPerm := range logPermList {
		if err := prepareLogDir(logPerm); err != nil {
			return fmt.Errorf("prepare log dir [%s] failed", logPerm.dir)
		}
	}
	return nil
}

func prepareLogDir(logPerm LogPerm) error {
	if err := fileutils.CreateDir(logPerm.dir, constants.Mode750); err != nil {
		hwlog.RunLog.Errorf("create log dir [%s] failed, error: %v", logPerm.dir, err)
		return fmt.Errorf("create log dir [%s] failed", logPerm.dir)
	}

	logFiles := []string{
		fmt.Sprintf("%s_%s", filepath.Base(logPerm.dir), constants.RunLogFile),
		fmt.Sprintf("%s_%s", filepath.Base(logPerm.dir), constants.OperateLogFile),
	}
	for _, logFile := range logFiles {
		logPath := filepath.Join(logPerm.dir, logFile)
		if err := fileutils.CreateFile(logPath, constants.Mode640); err != nil {
			hwlog.RunLog.Errorf("create log file [%s] failed, error: %v", logPath, err)
			return fmt.Errorf("create log file [%s] failed", logPath)
		}
	}

	param := fileutils.SetOwnerParam{
		Path:       logPerm.dir,
		Uid:        logPerm.userUid,
		Gid:        logPerm.userGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set dir [%s] owner group failed, error: %v", logPerm.dir, err)
		return fmt.Errorf("set dir [%s] owner group failed", logPerm.dir)
	}
	return nil
}
