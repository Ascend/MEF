// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/uninstall"
)

func TestUninstallCmd(t *testing.T) {
	convey.Convey("test uninstall cmd methods", t, uninstallCmdMethods)
	convey.Convey("test uninstall cmd successful", t, uninstallCmdSuccess)
	convey.Convey("test uninstall cmd failed", t, executeUninstallFailed)
}

func uninstallCmdMethods() {
	convey.So(UninstallCmd().Name(), convey.ShouldEqual, common.Uninstall)
	convey.So(UninstallCmd().Description(), convey.ShouldEqual, common.UninstallDesc)
	convey.So(UninstallCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(UninstallCmd().LockFlag(), convey.ShouldBeTrue)
}

func uninstallCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(uninstall.FlowUninstall{}, "RunTasks", nil)
	defer p.Reset()
	err := UninstallCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	UninstallCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeUninstallFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := UninstallCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("run uninstall flow failed", func() {
		p := gomonkey.ApplyMethodReturn(uninstall.FlowUninstall{}, "RunTasks", test.ErrTest)
		defer p.Reset()
		err := UninstallCmd().Execute(ctx)
		expectErr := fmt.Errorf("uninstall %s failed", constants.MEFEdgeName)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	UninstallCmd().PrintOpLogFail(userRoot, ipLocalhost)
}
