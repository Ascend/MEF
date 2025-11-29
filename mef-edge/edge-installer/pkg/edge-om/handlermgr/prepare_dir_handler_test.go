// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package handlermgr for deal every handler
package handlermgr

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

var (
	prepareHandler prepareDirHandler
	prepareDirMsg  = &model.Message{}
	testDir        = "/tmp/test_prepare_dir"
	expectErr      = errors.New("prepare directory for software failed")
)

func setupPrepareDir() error {
	var err error
	prepareDirMsg, err = util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:                constants.DownloadManagerName,
		Destination:           constants.ModEdgeOm,
		Operation:             constants.OptUpdate,
		Resource:              constants.InnerPrepareDir,
		TransferStructIntoStr: true,
	})
	if err != nil {
		hwlog.RunLog.Errorf("new test prepare dir message failed, error: %v", err)
		return err
	}
	if err = fileutils.CreateDir(testDir, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("create test prepare dir failed, error: %v", err)
		return err
	}
	return nil
}

func teardownPrepareDir() {
	if err := os.RemoveAll(testDir); err != nil {
		hwlog.RunLog.Errorf("clear test prepare dir failed, error: %v", err)
	}
}

func TestPrepareDirHandler(t *testing.T) {
	if err := setupPrepareDir(); err != nil {
		panic(err)
	}
	defer teardownPrepareDir()
	p := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
	defer p.Reset()
	convey.Convey("test prepare dir handler, test process create dir request", t, testProcessCreateDirRequest)
	convey.Convey("test prepare dir handler, test process delete dir request", t, testProcessDeleteDirRequest)
	convey.Convey("test prepare dir handler failed, parse dir request failed", t, parseReqFailed)
}

func testProcessCreateDirRequest() {
	p1 := gomonkey.ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(fileutils.CreateDir, nil)
	defer p1.Reset()

	err := prepareDirMsg.FillContent(config.DirReq{Path: testDir, ToDelete: false})
	convey.So(err, convey.ShouldBeNil)

	convey.Convey("process create dir request success", processCreateDirRequestSuccess)
	convey.Convey("process create dir request failed", processCreateDirRequestFailed)
}

func processCreateDirRequestSuccess() {
	err := prepareHandler.Handle(prepareDirMsg)
	convey.So(err, convey.ShouldBeNil)
}

func processCreateDirRequestFailed() {
	convey.Convey("delete path failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, testErr)
		defer p2.Reset()
		err := prepareHandler.Handle(prepareDirMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set permission of root dir failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.SetPathPermission, testErr)
		defer p2.Reset()
		err := prepareHandler.Handle(prepareDirMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("create dir failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.CreateDir, testErr)
		defer p2.Reset()
		err := prepareHandler.Handle(prepareDirMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set path owner failed", func() {
		p2 := gomonkey.ApplyFuncReturn(util.SetPathOwnerGroupToMEFEdge, testErr)
		defer p2.Reset()
		err := prepareHandler.Handle(prepareDirMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func testProcessDeleteDirRequest() {
	err := prepareDirMsg.FillContent(config.DirReq{Path: testDir, ToDelete: true})
	convey.So(err, convey.ShouldBeNil)

	convey.Convey("process delete dir request success", processDeleteDirRequestSuccess)
	convey.Convey("process delete dir request failed", processDeleteDirRequestFailed)
}

func processDeleteDirRequestSuccess() {
	err := prepareHandler.Handle(prepareDirMsg)
	convey.So(err, convey.ShouldBeNil)
}

func processDeleteDirRequestFailed() {
	p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, testErr)
	defer p1.Reset()
	err := prepareHandler.Handle(prepareDirMsg)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func parseReqFailed() {
	testInvalidMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new test invalid message failed, error: %v", err)
		return
	}
	err = testInvalidMsg.FillContent(model.RawMessage{})
	convey.So(err, convey.ShouldBeNil)
	err = prepareHandler.Handle(testInvalidMsg)
	convey.So(err, convey.ShouldResemble, errors.New("parse request parameter failed"))
}
