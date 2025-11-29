// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common
package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

var (
	testPath = "/tmp/test_utils"
	testOpt  = "test"
	testErr  = errors.New("test error")
)

func setup() error {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := util.InitHwLogger(logConfig, logConfig); err != nil {
		return fmt.Errorf("init hw log failed, error: %v", err)
	}

	dbPath := filepath.Join(testPath, constants.Config, constants.EdgeOm)
	if err := fileutils.CreateDir(dbPath, constants.Mode700); err != nil {
		return fmt.Errorf("create db directory failed, error: %v", err)
	}
	return nil
}

func teardown() {
	if err := os.RemoveAll(testPath); err != nil {
		hwlog.RunLog.Errorf("remove test path failed, error: %v", err)
	}
}

// TestMain run test main
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Printf("setup test environment failed: %v\n", err)
		return
	}
	defer teardown()
	exitCode := m.Run()
	fmt.Printf("test complete, exitCode=%d\n", exitCode)
}

func TestInitEdgeOmResource(t *testing.T) {
	patchOmDbPath := filepath.Join(testPath, constants.Config, constants.EdgeOm, constants.DbEdgeOmPath)
	p := gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, &pathmgr.ConfigPathMgr{}, nil).
		ApplyMethodReturn(&pathmgr.ConfigPathMgr{}, "GetEdgeOmDbPath", patchOmDbPath)
	defer p.Reset()

	convey.Convey("get install root dir failed", t, func() {
		p := gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, testErr)
		defer p.Reset()
		err := InitEdgeOmResource()
		expectErr := errors.New("get config path manager failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("init database failed", t, func() {
		p := gomonkey.ApplyFuncReturn(database.InitDB, testErr)
		defer p.Reset()
		err := InitEdgeOmResource()
		expectErr := errors.New("init database failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("table configurations create failed", t, func() {
		p := gomonkey.ApplyFuncReturn(database.InitDB, nil).
			ApplyFuncReturn(database.CreateTableIfNotExist, testErr)
		defer p.Reset()
		err := InitEdgeOmResource()
		expectErr := errors.New("table configurations create failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("init edge om resource success", t, func() {
		err := InitEdgeOmResource()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestLockProcessFlag(t *testing.T) {
	processFlag := util.FlagLockInstance(testPath, constants.ProcessFlag, testOpt)
	convey.Convey("lock process flag failed", t, func() {
		p := gomonkey.ApplyMethodReturn(processFlag, "Lock", testErr)
		defer p.Reset()
		err := LockProcessFlag(testPath, testOpt)
		expectErr := fmt.Errorf("lock control process failed: %v", testErr)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("lock process flag success", t, func() {
		p := gomonkey.ApplyMethodReturn(processFlag, "Lock", nil)
		defer p.Reset()
		err := LockProcessFlag(testPath, testOpt)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestUnlockProcessFlag(t *testing.T) {
	flag := util.FlagLockInstance(testPath, constants.ProcessFlag, testOpt)
	convey.Convey("unlock process flag failed", t, func() {
		p := gomonkey.ApplyMethodReturn(flag, "Unlock", testErr)
		defer p.Reset()
		err := UnlockProcessFlag(testPath, testOpt)
		expectErr := fmt.Errorf("unlock control process failed: %v", testErr)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("unlock process flag success", t, func() {
		p := gomonkey.ApplyMethodReturn(flag, "Unlock", nil)
		defer p.Reset()
		err := UnlockProcessFlag(testPath, testOpt)
		convey.So(err, convey.ShouldBeNil)
	})
}
