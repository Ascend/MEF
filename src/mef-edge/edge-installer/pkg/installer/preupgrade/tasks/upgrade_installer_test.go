// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for testing upgrade installer
package tasks

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

func TestUpgradeInstaller(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(path.GetLogRootDir, testDir, nil).
		ApplyFuncReturn(path.GetLogBackupRootDir, testDir, nil).
		ApplyFuncReturn(util.CheckNecessaryCommands, nil).
		ApplyFuncReturn(os.OpenFile, &os.File{}, nil).
		ApplyMethodReturn(&os.File{}, "Close", nil).
		ApplyMethodReturn(&fileutils.FileOwnerChecker{}, "Check", nil).
		ApplyFuncReturn(envutils.RunCommandWithOsStdout, nil)
	defer p.Reset()

	convey.Convey("upgrade run success", t, upgradeRunSuccess)
	convey.Convey("upgrade run failed, unknown call mode failed", t, unknownCallModeFailed)
	convey.Convey("upgrade run failed, init upgrade parameter failed", t, initUpgradeParaFailed)
	convey.Convey("upgrade run failed, check necessary commands failed", t, checkCommandsFailed)
	convey.Convey("upgrade run failed, upgrade file check invalid", t, upgradeFileInvalid)
	convey.Convey("upgrade run failed, run upgrade command failed", t, runCommandFailed)
}

func upgradeRunSuccess() {
	convey.Convey("default mode success", func() {
		err := UpgradeInstaller(testDir, constants.DefaultMode).Run()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("upgrade mode success", func() {
		err := UpgradeInstaller(testDir, constants.UpgradeMode).Run()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("effect mode success", func() {
		err := UpgradeInstaller(testDir, constants.EffectMode).Run()
		convey.So(err, convey.ShouldBeNil)
	})
}

func unknownCallModeFailed() {
	err := UpgradeInstaller(testDir, "").Run()
	convey.So(err, convey.ShouldResemble, errors.New("unknown call mode for upgrade installer"))
}

func initUpgradeParaFailed() {
	convey.Convey("get log root dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(path.GetLogRootDir, "", test.ErrTest)
		defer p1.Reset()
		err := UpgradeInstaller(testDir, constants.Upgrade).Run()
		expectErr := fmt.Errorf("get log root dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("init upgrade para failed, error: %v", expectErr))
	})

	convey.Convey("get log backup root dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(path.GetLogBackupRootDir, "", test.ErrTest)
		defer p1.Reset()
		err := UpgradeInstaller(testDir, constants.Upgrade).Run()
		expectErr := fmt.Errorf("get log backup root dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("init upgrade para failed, error: %v", expectErr))
	})
}

func checkCommandsFailed() {
	p1 := gomonkey.ApplyFuncReturn(util.CheckNecessaryCommands, test.ErrTest)
	defer p1.Reset()
	err := UpgradeInstaller(testDir, constants.Upgrade).Run()
	expectErr := errors.New("check necessary commands failed")
	convey.So(err, convey.ShouldResemble, fmt.Errorf("call upgrade bin failed, error: %v", expectErr))
}

func upgradeFileInvalid() {
	convey.Convey("open file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(os.OpenFile, &os.File{}, test.ErrTest)
		defer p1.Reset()
		err := UpgradeInstaller(testDir, constants.Upgrade).Run()
		expectErr1 := fmt.Errorf("open file %s failed", constants.UpgradePath)
		expectErr2 := fmt.Errorf("upgrade file check invalid: %v", expectErr1)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("call upgrade bin failed, error: %v", expectErr2))
	})

	convey.Convey("check file failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&fileutils.FileOwnerChecker{}, "Check", test.ErrTest)
		defer p1.Reset()
		err := UpgradeInstaller(testDir, constants.Upgrade).Run()
		expectErr := errors.New("upgrade file check invalid: check file failed")
		convey.So(err, convey.ShouldResemble, fmt.Errorf("call upgrade bin failed, error: %v", expectErr))
	})
}

func runCommandFailed() {
	p1 := gomonkey.ApplyFuncReturn(envutils.RunCommandWithOsStdout, test.ErrTest)
	defer p1.Reset()
	err := UpgradeInstaller(testDir, constants.Upgrade).Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("call upgrade bin failed, error: %v", test.ErrTest))
}
