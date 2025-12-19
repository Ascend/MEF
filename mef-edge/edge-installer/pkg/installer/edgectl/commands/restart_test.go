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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	com "edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/edgectl/common"
)

func TestRestartCmd(t *testing.T) {
	convey.Convey("test restart cmd methods", t, restartCmdMethods)
	convey.Convey("test restart cmd successful", t, restartCmdSuccess)
	convey.Convey("test restart cmd failed", t, executeRestartFailed)
}

func restartCmdMethods() {
	convey.So(RestartCmd().Name(), convey.ShouldEqual, common.Restart)
	convey.So(RestartCmd().Description(), convey.ShouldEqual, common.RestartDesc)
	convey.So(RestartCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(RestartCmd().LockFlag(), convey.ShouldBeTrue)
}

func restartCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(com.ComponentMgr{}, "RestartAll", nil)
	defer p.Reset()
	err := RestartCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	RestartCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeRestartFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := RestartCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("run restart all failed", func() {
		p := gomonkey.ApplyMethodReturn(com.ComponentMgr{}, "RestartAll", test.ErrTest)
		defer p.Reset()
		err := RestartCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	RestartCmd().PrintOpLogFail(userRoot, ipLocalhost)
}
