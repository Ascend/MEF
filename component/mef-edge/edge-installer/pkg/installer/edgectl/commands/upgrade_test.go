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

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/preupgrade/flows"
)

const (
	testInvalidPath = "/home/data/mefedge/unpack/test.tar.gz"
	testTarPath     = "/test/test.tar.gz"
	testCmsPath     = "/test/test.tar.gz.cms"
	testCrlPath     = "/test/test.tar.gz.crl"
	testDelayEffect = false
)

func TestUpgradeCmd(t *testing.T) {
	convey.Convey("test upgrade cmd methods", t, upgradeCmdMethods)
	convey.Convey("test upgrade cmd successful", t, upgradeCmdSuccess)
	convey.Convey("test upgrade cmd failed", t, func() {
		convey.Convey("execute upgrade Failed", executeUpgradeFailed)
		convey.Convey("check param Failed", checkParamFailed)
		convey.Convey("check single path Failed", checkSinglePathFailed)
	})
}

func upgradeCmdMethods() {
	convey.So(UpgradeCmd().Name(), convey.ShouldEqual, common.Upgrade)
	convey.So(UpgradeCmd().Description(), convey.ShouldEqual, common.UpgradeDesc)
	convey.So(UpgradeCmd().BindFlag(), convey.ShouldBeTrue)
	convey.So(UpgradeCmd().LockFlag(), convey.ShouldBeTrue)
}

func upgradeCmdSuccess() {
	param := flows.OfflineUpgradeInstallerParam{TarPath: testTarPath, CmsPath: testCmsPath,
		CrlPath: testCrlPath, EdgeDir: "./", DelayEffect: testDelayEffect}
	flow := flows.OfflineUpgradeInstaller(param)
	p := gomonkey.ApplyFuncReturn(UpgradeCmd, &upgradeCmd{
		tarPath: testTarPath,
		cmsPath: testCmsPath,
		crlPath: testCrlPath,
	}).
		ApplyPrivateMethod(&upgradeCmd{}, "checkParam", func() error { return nil }).
		ApplyMethodReturn(flow, "RunTasks", nil)
	defer p.Reset()

	err := UpgradeCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	UpgradeCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeUpgradeFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := UpgradeCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("run offline upgrade edge-installer flow failed", func() {
		param := flows.OfflineUpgradeInstallerParam{TarPath: testTarPath, CmsPath: testCmsPath,
			CrlPath: testCrlPath, EdgeDir: "./", DelayEffect: testDelayEffect}
		flow := flows.OfflineUpgradeInstaller(param)
		p := gomonkey.ApplyFuncReturn(UpgradeCmd, &upgradeCmd{
			tarPath: testTarPath,
			cmsPath: testCmsPath,
			crlPath: testCrlPath,
		}).
			ApplyPrivateMethod(&upgradeCmd{}, "checkParam", func() error { return nil }).
			ApplyMethodReturn(flow, "RunTasks", test.ErrTest)
		defer p.Reset()
		err := UpgradeCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	UpgradeCmd().PrintOpLogFail(userRoot, ipLocalhost)
}

func checkParamFailed() {
	convey.Convey("tar or cms or crl file not input failed", func() {
		p := gomonkey.ApplyFuncReturn(UpgradeCmd, &upgradeCmd{
			tarPath: "", cmsPath: "", crlPath: ""})
		defer p.Reset()
		err := UpgradeCmd().Execute(ctx)
		expectErr := errors.New("tar or cms or crl file not input")
		convey.So(err, convey.ShouldResemble, fmt.Errorf("check param failed, %v", expectErr))
	})
}

func checkSinglePathFailed() {
	p := gomonkey.ApplyFuncReturn(UpgradeCmd, &upgradeCmd{
		tarPath: testTarPath,
		cmsPath: testCmsPath,
		crlPath: testCrlPath,
	})
	defer p.Reset()

	convey.Convey("path does not exist", func() {
		err := UpgradeCmd().Execute(ctx)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring, "no such file or directory")
	})

	convey.Convey("path cannot be in the decompression path", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, testInvalidPath, nil)
		defer p1.Reset()
		err := UpgradeCmd().Execute(ctx)
		expectErr := errors.New("check tar path failed: the path cannot be in the decompression path")
		convey.So(err, convey.ShouldResemble, fmt.Errorf("check param failed, %v", expectErr))
	})
}
