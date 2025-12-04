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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/util"
	com "edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/edgectl/common"
)

func TestStopCmd(t *testing.T) {
	convey.Convey("test stop cmd methods", t, stopCmdMethods)
	convey.Convey("test stop cmd successful", t, stopCmdSuccess)
	convey.Convey("test stop cmd failed", t, executeStopFailed)
}

func stopCmdMethods() {
	convey.So(StopCmd().Name(), convey.ShouldEqual, common.Stop)
	convey.So(StopCmd().Description(), convey.ShouldEqual, common.StopDesc)
	convey.So(StopCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(StopCmd().LockFlag(), convey.ShouldBeTrue)
}

func stopCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(com.ComponentMgr{}, "GetComponents", []com.Component{testComponent}).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(util.IsServiceActive, true).
		ApplyMethodReturn(com.ComponentMgr{}, "StopAll", nil)
	defer p.Reset()
	err := StopCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	StopCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeStopFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := StopCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("run stop all failed", func() {
		p := gomonkey.ApplyMethodReturn(com.ComponentMgr{}, "GetComponents", []com.Component{}).
			ApplyMethodReturn(com.ComponentMgr{}, "StopAll", test.ErrTest)
		defer p.Reset()
		err := StopCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	StopCmd().PrintOpLogFail(userRoot, ipLocalhost)
}
