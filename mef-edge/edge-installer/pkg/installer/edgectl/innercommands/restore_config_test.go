// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package innercommands
package innercommands

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/common/tasks"
	com "edge-installer/pkg/installer/edgectl/common"
)

func TestRestoreCfg(t *testing.T) {
	convey.Convey("test restore config cmd methods", t, restoreCfgCmdMethods)
	convey.Convey("test restore config cmd successful", t, restoreCfgCmdSuccess)
	convey.Convey("test restore config cmd failed", t, func() {
		convey.Convey("ctx is nil", restoreCtxIsNilFailed)
		convey.Convey("stop service failed", stopServiceFailed)
		convey.Convey("restore default config dir failed", restoreCfgDirFailed)
	})
}

func restoreCfgCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
		ApplyFuncReturn(copyCfgDir, nil).
		ApplyFuncReturn(path.GetLogRootDir, "", nil).
		ApplyFuncReturn(path.GetLogBackupRootDir, "", nil).
		ApplyMethodReturn(&tasks.SetSystemInfoTask{}, "Run", nil).
		ApplyFuncReturn(fileutils.RenameFile, nil).
		ApplyFuncReturn(util.RemoveContainer, nil)
	defer p.Reset()
	err := RestoreCfgCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	RestoreCfgCmd().PrintOpLogOk("root", "localhost")
}

func restoreCtxIsNilFailed() {
	err := RestoreCfgCmd().Execute(nil)
	expectErr := errors.New("ctx is nil")
	convey.So(err, convey.ShouldResemble, expectErr)
	RestoreCfgCmd().PrintOpLogFail("root", "localhost")
}

func stopServiceFailed() {
	p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", test.ErrTest)
	defer p.Reset()
	err := RestoreCfgCmd().Execute(ctx)
	retErr := fmt.Errorf("stop service failed, error: %v", test.ErrTest)
	convey.So(err, convey.ShouldResemble, retErr)
}

func restoreCfgDirFailed() {
	convey.Convey("restore default config to config temp dir failed", func() {
		p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
			ApplyFuncReturn(copyCfgDir, test.ErrTest)
		defer p.Reset()
		err := RestoreCfgCmd().Execute(ctx)
		retErr := fmt.Errorf("restore default config to config temp dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, retErr)
	})

	convey.Convey("set system info into config files failed", func() {
		p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
			ApplyFuncReturn(copyCfgDir, nil).
			ApplyFuncReturn(path.GetLogRootDir, "", nil).
			ApplyFuncReturn(path.GetLogBackupRootDir, "", nil).
			ApplyMethodReturn(&tasks.SetSystemInfoTask{}, "Run", test.ErrTest)
		defer p.Reset()
		err := RestoreCfgCmd().Execute(ctx)
		retErr := errors.New("set system info into config files failed")
		convey.So(err, convey.ShouldResemble, retErr)
	})

	convey.Convey("backup cur config dir failed", func() {
		p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
			ApplyFuncReturn(copyCfgDir, nil).
			ApplyFuncReturn(path.GetLogRootDir, "", nil).
			ApplyFuncReturn(path.GetLogBackupRootDir, "", nil).
			ApplyMethodReturn(&tasks.SetSystemInfoTask{}, "Run", nil).
			ApplyFuncReturn(fileutils.RenameFile, test.ErrTest)
		defer p.Reset()
		err := RestoreCfgCmd().Execute(ctx)
		retErr := fmt.Errorf("backup cur config dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, retErr)
	})

	convey.Convey("restore default config failed, but recover env successful", func() {
		mockRenameFunOuts := []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{test.ErrTest}},
			{Values: gomonkey.Params{nil}},
		}
		p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
			ApplyFuncReturn(copyCfgDir, nil).
			ApplyFuncReturn(path.GetLogRootDir, "", nil).
			ApplyFuncReturn(path.GetLogBackupRootDir, "", nil).
			ApplyMethodReturn(&tasks.SetSystemInfoTask{}, "Run", nil).
			ApplyFuncSeq(fileutils.RenameFile, mockRenameFunOuts)
		defer p.Reset()
		err := RestoreCfgCmd().Execute(ctx)
		retErr := fmt.Errorf("restore default config failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, retErr)
	})

	convey.Convey("restore default config failed, and recover env failed", func() {
		mockRenameFunOuts := []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{test.ErrTest}},
			{Values: gomonkey.Params{test.ErrTest}},
		}
		p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
			ApplyFuncReturn(copyCfgDir, nil).
			ApplyFuncReturn(path.GetLogRootDir, "", nil).
			ApplyFuncReturn(path.GetLogBackupRootDir, "", nil).
			ApplyMethodReturn(&tasks.SetSystemInfoTask{}, "Run", nil).
			ApplyFuncSeq(fileutils.RenameFile, mockRenameFunOuts)
		defer p.Reset()
		err := RestoreCfgCmd().Execute(ctx)
		retErr := fmt.Errorf("restore default config failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, retErr)
	})
}

func restoreCfgCmdMethods() {
	convey.So(RestoreCfgCmd().Name(), convey.ShouldEqual, com.RestoreCfgCmd)
	convey.So(RestoreCfgCmd().Description(), convey.ShouldEqual, com.RestoreCfgDesc)
	convey.So(RestoreCfgCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(RestoreCfgCmd().LockFlag(), convey.ShouldBeTrue)
}

func TestCopyCfgDir(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).
		ApplyFuncReturn(fileutils.CopyDir, nil)
	defer p.Reset()

	convey.Convey("copy config dir should be success", t, testCopyCfgDir)
	convey.Convey("copy config dir should be failed", t, func() {
		convey.Convey("create dir failed", testCopyCfgDirErrCreateDir)
		convey.Convey("copy dir failed", testCopyCfgDirErrCopyDir)
		convey.Convey("get uid or gid failed", testCopyCfgDirErrGetUidOrGid)
		convey.Convey("set path owner group failed", testCopyCfgDirErrSetPathOwnerGroup)
		convey.Convey("run command failed", testCopyCfgDirErrRunCommand)
	})
}

func testCopyCfgDir() {
	p := gomonkey.ApplyFuncReturn(util.GetMefId, uint32(constants.EdgeUserUid), uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(fileutils.SetPathOwnerGroup, nil).
		ApplyFuncReturn(envutils.RunCommandWithUser, "", nil)
	defer p.Reset()

	err := copyCfgDir("", "")
	convey.So(err, convey.ShouldBeNil)
}

func testCopyCfgDirErrCreateDir() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},
	}

	var p1 = gomonkey.ApplyFuncSeq(fileutils.CreateDir, outputs).
		ApplyFuncReturn(util.GetMefId, uint32(constants.EdgeUserUid), uint32(constants.EdgeUserGid), nil)
	defer p1.Reset()

	err := copyCfgDir("dirSrc", "dirDst")
	convey.So(err, convey.ShouldResemble, fmt.Errorf("create dir [%s] failed, error: %v", "dirDst", test.ErrTest))

	err = copyCfgDir("dirSrc", "dirDst")
	edgeMainDirDst := filepath.Join("dirDst", constants.EdgeMain)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("create dir [%s] failed, error: %v", edgeMainDirDst, test.ErrTest))
}

func testCopyCfgDirErrCopyDir() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CopyDir, test.ErrTest)
	defer p1.Reset()

	err := copyCfgDir("dirSrc", "dirDst")
	subDirSrc := filepath.Join("dirSrc", constants.EdgeInstaller)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("copy dir [%s] failed: %v", subDirSrc, test.ErrTest))
}

func testCopyCfgDirErrGetUidOrGid() {
	var p1 = gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(constants.EdgeUserUid), test.ErrTest)
	defer p1.Reset()

	err := copyCfgDir("", "")
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get uid or gid failed: %v", test.ErrTest))
}

func testCopyCfgDirErrSetPathOwnerGroup() {
	var p1 = gomonkey.ApplyFuncReturn(util.GetMefId, uint32(constants.EdgeUserUid), uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(fileutils.SetPathOwnerGroup, test.ErrTest)
	defer p1.Reset()

	err := copyCfgDir("", "")
	convey.So(err, convey.ShouldResemble, fmt.Errorf("set dir [%s] owner and group failed,"+
		" error: %v", constants.EdgeMain, test.ErrTest))
}

func testCopyCfgDirErrRunCommand() {
	var p1 = gomonkey.ApplyFuncReturn(util.GetMefId, uint32(constants.EdgeUserUid), uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(fileutils.SetPathOwnerGroup, nil).
		ApplyFuncReturn(envutils.RunCommandWithUser, "", test.ErrTest)
	defer p1.Reset()

	err := copyCfgDir("", "")
	convey.So(err, convey.ShouldResemble, fmt.Errorf("copy dir [%s] failed: %v", constants.EdgeMain, test.ErrTest))
}
