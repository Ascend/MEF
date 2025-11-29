// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing check install environment task
package tasks

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

var checkInstallEnvironment = CheckInstallEnvironmentTask{
	InstallRootDir: testDir,
	LogPathMgr:     pathmgr.NewLogPathMgr(testDir, testDir),
}

func TestCheckInstallEnvironmentTask(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(envutils.GetFileDevNum, uint64(0), nil).
		ApplyFuncReturn(envutils.CheckDiskSpace, nil).
		ApplyFuncReturn(exec.LookPath, "", nil).
		ApplyFuncReturn(util.CheckNecessaryCommands, nil)
	defer p.Reset()

	convey.Convey("check environment success", t, checkEnvSuccess)
	convey.Convey("check environment failed, get file dev num failed", t, getFileDevNumFailed)
	convey.Convey("check environment failed, check disk space failed", t, checkDiskSpaceFailed)
	convey.Convey("check environment failed, dev map is nil failed", t, devMapIsNilFailed)
	convey.Convey("check environment failed, check necessary commands failed", t, checkNecessaryCommandsFailed)
}

func checkEnvSuccess() {
	err := checkInstallEnvironment.Run()
	convey.So(err, convey.ShouldBeNil)
}

func getFileDevNumFailed() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{uint64(0), test.ErrTest}},

		{Values: gomonkey.Params{uint64(0), nil}, Times: 2},
		{Values: gomonkey.Params{uint64(0), test.ErrTest}},
	}
	var p1 = gomonkey.ApplyFuncSeq(envutils.GetFileDevNum, outputs)
	defer p1.Reset()

	err := checkInstallEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("check path disk space failed"))

	err = checkInstallEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("check path disk space failed"))
}

func checkDiskSpaceFailed() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{errors.New("check path disk space failed: no enough space")}, Times: 2},
	}
	var p1 = gomonkey.ApplyFuncSeq(envutils.CheckDiskSpace, outputs)
	defer p1.Reset()

	err := checkInstallEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("check path disk space failed"))

	err = checkInstallEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("check path disk space failed"))

	err = checkInstallEnvironment.Run()
	convey.So(err, convey.ShouldBeNil)
}

func devMapIsNilFailed() {
	devMap, err := checkInstallEnvironment.checkDevDiskSpace(testDir, uint64(0), nil)
	convey.So(devMap, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, errors.New("devMap is nil"))
}

func checkNecessaryCommandsFailed() {
	var p1 = gomonkey.ApplyFuncReturn(util.CheckNecessaryCommands, test.ErrTest)
	defer p1.Reset()

	err := checkInstallEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("check necessary commands failed"))
}
