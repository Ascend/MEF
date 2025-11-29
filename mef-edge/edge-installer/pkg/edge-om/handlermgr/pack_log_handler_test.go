// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package handlermgr for deal every handler
package handlermgr

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

var (
	packLogMsg        *model.Message
	testLogCollectDir = "/tmp/test_log_collect"
)

func setupPackLog() error {
	if err := fileutils.CreateDir(testLogCollectDir, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("create test log collect dir failed, error: %v", err)
		return err
	}
	return nil
}

func teardownPackLog() {
	if err := os.RemoveAll(testLogCollectDir); err != nil {
		hwlog.RunLog.Errorf("clear test log collect dir failed, error: %v", err)
	}
}

func TestPackLogHandler(t *testing.T) {
	if err := setupPackLog(); err != nil {
		return
	}
	defer teardownPackLog()
	convey.Convey("test pack log handler", t, func() {
		convey.Convey("pack log handler is busy", packLogHandlerBusy)
		convey.Convey("test prepare dirs", testPrepareDirs)
		convey.Convey("test do collect", testDoCollect)
		convey.Convey("test do change permission", testDoChangePermission)
		convey.Convey("test do clean", testDoClean)
	})
}

func packLogHandlerBusy() {
	p := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
	defer p.Reset()
	handler := packLogHandler{running: 1}
	err := handler.Handle(packLogMsg)
	convey.So(err, convey.ShouldResemble, errors.New("log pack handler is busy"))
}

func testPrepareDirs() {
	convey.Convey("prepare dirs successful", prepareDirsSuccess)
	convey.Convey("prepare dirs failed", func() {
		convey.Convey("clean temp files failed", cleanTempFilesFailed)
		convey.Convey("create temp dirs failed", createTempDirsFailed)
	})
}

func testDoCollect() {
	convey.Convey("collect log successful", doCollectSuccess)
	convey.Convey("collect log failed", doCollectFailed)
}

func testDoChangePermission() {
	convey.Convey("change permission successful", changePermissionSuccess)
	convey.Convey("change permission failed", changePermissionFailed)
}

func testDoClean() {
	handler := packLogHandler{}
	handler.doClean(false)
	handler.doClean(true)
}

func prepareDirsSuccess() {
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncSeq(fileutils.CreateDir, []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{nil}},
		}).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, nil)
	defer p.Reset()

	err := prepareDirs()
	convey.So(err, convey.ShouldBeNil)
}

func cleanTempFilesFailed() {
	rootDirFile, err := os.Open(testLogCollectDir)
	if err != nil {
		return
	}
	defer func() {
		if err = rootDirFile.Close(); err != nil {
			return
		}
	}()
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(os.Open, rootDirFile, nil)
	defer p.Reset()

	convey.Convey("check root dir failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&fileutils.FileOwnerChecker{}, "Check", testErr)
		defer p1.Reset()
		handler := packLogHandler{}
		err = handler.handle()
		expectErr := fmt.Errorf("failed to prepare temp dirs, failed to clean temp dir: "+
			"failed to check root dir /home/data/mef_logcollect, %v", testErr)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("delete edge om file failed", func() {
		p2 := gomonkey.ApplyMethodSeq(&fileutils.FileOwnerChecker{}, "Check", []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{nil}},
		}).
			ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
			ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil).
			ApplyFuncReturn(fileutils.DeleteFile, testErr)
		defer p2.Reset()
		handler := packLogHandler{}
		err = handler.handle()
		expectErr := fmt.Errorf("failed to prepare temp dirs, failed to clean temp dir: "+
			"failed to delete edge om file /home/data/mef_logcollect/edge_om/edgeNode.tar.gz, %v", testErr)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func createTempDirsFailed() {
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncReturn(fileutils.CreateDir, testErr)
	defer p.Reset()

	handler := packLogHandler{}
	err := handler.handle()
	expectErr := fmt.Errorf("failed to prepare temp dirs, failed to create temp dir: "+
		"failed to create root dir /home/data/mef_logcollect, %v", testErr)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func doCollectSuccess() {
	collector := util.GetLogCollector("", "", "", []string{""})
	p := gomonkey.ApplyFuncReturn(path.GetInstallRootDir, "", nil).
		ApplyFuncReturn(path.GetLogRootDir, "", nil).
		ApplyFuncReturn(path.GetLogBackupRootDir, "", nil).
		ApplyMethodReturn(collector, "Collect", "", nil)
	defer p.Reset()
	handler := packLogHandler{}
	err := handler.doCollect()
	convey.So(err, convey.ShouldBeNil)
}

func doCollectFailed() {
	collector := util.GetLogCollector("", "", "", []string{""})
	p := gomonkey.ApplyFuncReturn(prepareDirs, nil).
		ApplyFuncReturn(path.GetInstallRootDir, "", nil).
		ApplyFuncReturn(path.GetLogRootDir, "", nil).
		ApplyFuncReturn(path.GetLogBackupRootDir, "", nil).
		ApplyMethodReturn(collector, "Collect", "", testErr)
	defer p.Reset()
	handler := packLogHandler{}
	err := handler.handle()
	expectErr := fmt.Errorf("failed to collect log, %v", testErr)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("failed to collect logs, %v", expectErr))
}

func changePermissionSuccess() {
	p := gomonkey.ApplyFuncReturn(utils.SafeChmod, nil).
		ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, nil).
		ApplyFuncReturn(fileutils.RenameFile, nil)
	defer p.Reset()
	handler := packLogHandler{}
	err := handler.doChangePermission()
	convey.So(err, convey.ShouldBeNil)
}

func changePermissionFailed() {
	p := gomonkey.ApplyFuncReturn(utils.SafeChmod, testErr)
	defer p.Reset()
	handler := packLogHandler{}
	err := handler.doChangePermission()
	expectErr := fmt.Errorf("failed to change permission of log, %v", testErr)
	convey.So(err, convey.ShouldResemble, expectErr)
}
