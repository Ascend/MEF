// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package commands
package commands

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

func TestEffectCmd(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
		ApplyMethodReturn(config.VersionXmlMgr{}, "GetVersion", "5.0.RC1", nil).
		ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "./upgrade", nil).
		ApplyFuncReturn(path.GetLogRootDir, "/var/alog", nil).
		ApplyFuncReturn(path.GetLogBackupRootDir, "/home/log", nil).
		ApplyFuncReturn(pathmgr.GetTargetInstallDir, "./software", nil).
		ApplyFuncReturn(util.UnSetImmutable, nil).
		ApplyFuncReturn(config.EffectToOldestVersionSmooth, nil).
		ApplyFuncReturn(envutils.RunCommandWithOsStdout, nil)
	defer p.Reset()

	convey.Convey("test effect cmd methods", t, effectCmdMethods)
	convey.Convey("test effect cmd successful", t, effectCmdSuccess)
	convey.Convey("test effect cmd failed", t, func() {
		convey.Convey("ctx is nil failed", effectCtxIsNilFailed)
		convey.Convey("execute effect cmd failed", executeCmdFailed)
		convey.Convey("call upgrade bin failed", callUpgradeBinFailed)
		convey.Convey("effect post process failed", effectPostProcFailed)
	})
}

func effectCmdMethods() {
	convey.So(EffectCmd().Name(), convey.ShouldEqual, common.Effect)
	convey.So(EffectCmd().Description(), convey.ShouldEqual, common.EffectDesc)
	convey.So(EffectCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(EffectCmd().LockFlag(), convey.ShouldBeTrue)
}

func effectCmdSuccess() {
	err := EffectCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	EffectCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func effectCtxIsNilFailed() {
	err := EffectCmd().Execute(nil)
	expectErr := errors.New("ctx is nil")
	convey.So(err, convey.ShouldResemble, expectErr)
	EffectCmd().PrintOpLogFail(userRoot, ipLocalhost)
}

func executeCmdFailed() {
	convey.Convey("no software need to effect", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
		defer p.Reset()
		err := EffectCmd().Execute(ctx)
		expectErr := errors.New("no software is to effect")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get old inner version failed", func() {
		p := gomonkey.ApplyMethodReturn(config.VersionXmlMgr{}, "GetVersion", "", test.ErrTest)
		defer p.Reset()
		err := EffectCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("check effect bin file failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CheckOwnerAndPermission, "", test.ErrTest)
		defer p.Reset()
		err := EffectCmd().Execute(ctx)
		expectErr := errors.New("check effect bin file failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get log root dir failed", func() {
		p := gomonkey.ApplyFuncReturn(path.GetLogRootDir, "", test.ErrTest)
		defer p.Reset()
		err := EffectCmd().Execute(ctx)
		expectErr := fmt.Errorf("get log root dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get log backup root dir failed", func() {
		p := gomonkey.ApplyFuncReturn(path.GetLogBackupRootDir, "", test.ErrTest)
		defer p.Reset()
		err := EffectCmd().Execute(ctx)
		expectErr := fmt.Errorf("get log backup root dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func callUpgradeBinFailed() {
	p := gomonkey.ApplyFuncReturn(util.UnSetImmutable, test.ErrTest).
		ApplyFuncReturn(envutils.RunCommandWithOsStdout, test.ErrTest)
	defer p.Reset()
	err := EffectCmd().Execute(ctx)
	expectErr := fmt.Errorf("call upgrade bin failed, error: %v", test.ErrTest)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func effectPostProcFailed() {
	p := gomonkey.ApplyFuncReturn(config.EffectToOldestVersionSmooth, test.ErrTest)
	defer p.Reset()
	err := EffectCmd().Execute(ctx)
	convey.So(err, convey.ShouldResemble, errors.New("smooth config file for old version failed"))
}
