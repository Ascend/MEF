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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path"
	"edge-installer/pkg/installer/edgectl/common"
)

const dfCmdOut = "Filesystem      Size  Used Avail Use% Mounted on\n" +
	"/dev/mmcblk0p5  974M   42M  865M   5% /home/log\n" +
	"/dev/mmcblk0p7  2.9G  286M  2.5G  11% /home/package\n" +
	"/dev/mmcblk0p6  2.0G  175M  1.7G  10% /usr/local/mindx"

func TestCopyResetScriptCmd(t *testing.T) {
	convey.Convey("test copy reset script cmd methods", t, copyResetScriptCmdMethods)
	convey.Convey("test copy reset script cmd successful", t, copyResetScriptCmdSuccess)
	convey.Convey("test copy reset script cmd failed", t, copyResetScriptCmdFailed)
}

func copyResetScriptCmdSuccess() {
	p := gomonkey.ApplyFuncReturn(envutils.RunCommand, dfCmdOut, nil).
		ApplyFuncReturn(path.GetCompWorkDir, "", nil).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(fileutils.CopyFile, nil)
	defer p.Reset()
	err := CopyResetScriptCmd().Execute(&common.Context{})
	convey.So(err, convey.ShouldBeNil)
	CopyResetScriptCmd().PrintOpLogOk("root", "localhost")
}

func copyResetScriptCmdFailed() {
	convey.Convey("ctx is nil", func() {
		err := CopyResetScriptCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
		CopyResetScriptCmd().PrintOpLogFail("root", "localhost")
	})

	convey.Convey("execute df command failed", func() {
		p := gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p.Reset()
		err := CopyResetScriptCmd().Execute(&common.Context{})
		expectErr := fmt.Errorf("execute [df] command failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get software work dir failed", func() {
		p := gomonkey.ApplyFuncReturn(envutils.RunCommand, dfCmdOut, nil).
			ApplyFuncReturn(path.GetCompWorkDir, "", test.ErrTest)
		defer p.Reset()
		err := CopyResetScriptCmd().Execute(&common.Context{})
		expectErr := fmt.Errorf("get component work dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("firmware path does not exist", func() {
		p := gomonkey.ApplyFuncReturn(envutils.RunCommand, dfCmdOut, nil).
			ApplyFuncReturn(path.GetCompWorkDir, "/", nil).
			ApplyFuncReturn(fileutils.IsExist, false)
		defer p.Reset()
		err := CopyResetScriptCmd().Execute(&common.Context{})
		expectErr := fmt.Errorf("firmware path [/home/package/firmware] does not exist")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("delete existed script failed", func() {
		p := gomonkey.ApplyFuncReturn(envutils.RunCommand, dfCmdOut, nil).
			ApplyFuncReturn(path.GetCompWorkDir, "/", nil).
			ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest)
		defer p.Reset()
		err := CopyResetScriptCmd().Execute(&common.Context{})
		expectErr := fmt.Errorf("delete existed [/home/package/firmware/reset_middleware.sh] failed, error: %v",
			test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("copy reset script failed", func() {
		p := gomonkey.ApplyFuncReturn(envutils.RunCommand, dfCmdOut, nil).
			ApplyFuncReturn(path.GetCompWorkDir, "/", nil).
			ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(fileutils.DeleteFile, nil).
			ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
		defer p.Reset()
		err := CopyResetScriptCmd().Execute(&common.Context{})
		expectErr := fmt.Errorf("copy [/script/reset_middleware.sh] to "+
			"[/home/package/firmware/reset_middleware.sh] failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func copyResetScriptCmdMethods() {
	convey.So(CopyResetScriptCmd().Name(), convey.ShouldEqual, common.CopyResetScriptCmd)
	convey.So(CopyResetScriptCmd().Description(), convey.ShouldEqual, common.CopyResetScriptDesc)
	convey.So(CopyResetScriptCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(CopyResetScriptCmd().LockFlag(), convey.ShouldBeFalse)
}
