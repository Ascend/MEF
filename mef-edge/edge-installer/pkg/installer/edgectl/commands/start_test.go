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

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	com "edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/edgectl/common"
)

var testComponent = com.Component{
	Name: constants.MefInitServiceName,
	Dir:  "./componentDir",
	Service: com.FileInfo{
		Name:      constants.EdgeInitServiceFile,
		Path:      "./servicePath",
		ModeUmask: constants.ModeUmask077,
		UserName:  constants.RootUserName,
	},
	Bin: com.FileInfo{
		Name:      constants.MefInitScriptName,
		Path:      "./binPath",
		ModeUmask: constants.ModeUmask077,
		UserName:  constants.RootUserName,
	},
}

func TestStartCmd(t *testing.T) {
	convey.Convey("test start cmd methods", t, startCmdMethods)
	convey.Convey("test start cmd successful", t, startCmdSuccess)
	convey.Convey("test start cmd failed", t, executeStartFailed)
}

func startCmdMethods() {
	convey.So(StartCmd().Name(), convey.ShouldEqual, common.Start)
	convey.So(StartCmd().Description(), convey.ShouldEqual, common.StartDesc)
	convey.So(StartCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(StartCmd().LockFlag(), convey.ShouldBeTrue)
}

func startCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(com.ComponentMgr{}, "GetComponents", []com.Component{testComponent}).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(util.IsServiceActive, true).
		ApplyMethodReturn(com.ComponentMgr{}, "StartAll", nil)
	defer p.Reset()
	err := StartCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	StartCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeStartFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := StartCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("run start all failed", func() {
		p := gomonkey.ApplyMethodReturn(com.ComponentMgr{}, "GetComponents", []com.Component{}).
			ApplyMethodReturn(com.ComponentMgr{}, "StartAll", test.ErrTest)
		defer p.Reset()
		err := StartCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	StartCmd().PrintOpLogFail(userRoot, ipLocalhost)
}
