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

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
)

func TestGetNetCfgCmd(t *testing.T) {
	convey.Convey("test get net config cmd methods", t, getNetCfgCmdMethods)
	convey.Convey("test get net config cmd successful", t, getNetCfgCmdSuccess)
	convey.Convey("test get net config cmd failed", t, func() {
		convey.Convey("execute with code failed", executeWithCodeFailed)
		convey.Convey("get net config failed", getNetConfigFailed)
	})
}

func getNetCfgCmdMethods() {
	convey.So(GetNetCfgCmd().Name(), convey.ShouldEqual, common.GetNetCfg)
	convey.So(GetNetCfgCmd().Description(), convey.ShouldEqual, common.GetNetCfgDesc)
	convey.So(GetNetCfgCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(GetNetCfgCmd().LockFlag(), convey.ShouldBeFalse)
	convey.So(GetNetCfgCmd().Execute(&common.Context{}), convey.ShouldBeNil)
}

func getNetCfgCmdSuccess() {
	convey.Convey("current net type is [FD] and with om is [True]", func() {
		p := gomonkey.ApplyFuncReturn(config.GetNetManager, &config.NetManager{NetType: constants.FD, WithOm: true}, nil)
		defer p.Reset()
		code, err := GetNetCfgCmd().(common.CommandWithRetCode).ExecuteWithCode(ctx)
		convey.So(code, convey.ShouldEqual, fdWithOmCode)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("current net type is [FD] and with om is [False]", func() {
		p := gomonkey.ApplyFuncReturn(config.GetNetManager, &config.NetManager{NetType: constants.FD, WithOm: false}, nil)
		defer p.Reset()
		code, err := GetNetCfgCmd().(common.CommandWithRetCode).ExecuteWithCode(ctx)
		convey.So(code, convey.ShouldEqual, fdWithOutOmCode)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("current net type is [MEF]", func() {
		p := gomonkey.ApplyFuncReturn(config.GetNetManager, &config.NetManager{NetType: constants.MEF, WithOm: false}, nil)
		defer p.Reset()
		code, err := GetNetCfgCmd().(common.CommandWithRetCode).ExecuteWithCode(ctx)
		convey.So(code, convey.ShouldEqual, otherNetCode)
		convey.So(err, convey.ShouldBeNil)
	})
	GetNetCfgCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeWithCodeFailed() {
	convey.Convey("ctx is nil failed", func() {
		code, err := GetNetCfgCmd().(common.CommandWithRetCode).ExecuteWithCode(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(code, convey.ShouldEqual, defaultErrorCode)
		convey.So(err, convey.ShouldResemble, expectErr)
		GetNetCfgCmd().PrintOpLogFail(userRoot, ipLocalhost)
	})
}

func getNetConfigFailed() {
	convey.Convey("get install root dir failed", func() {
		p := gomonkey.ApplyFuncReturn(config.GetComponentDbMgr, nil, test.ErrTest)
		defer p.Reset()
		code, err := GetNetCfgCmd().(common.CommandWithRetCode).ExecuteWithCode(ctx)
		expectErr := errors.New("get component db manager failed")
		convey.So(code, convey.ShouldEqual, defaultErrorCode)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get net config failed", func() {
		p := gomonkey.ApplyFuncReturn(config.GetNetManager, nil, test.ErrTest)
		defer p.Reset()
		code, err := GetNetCfgCmd().(common.CommandWithRetCode).ExecuteWithCode(ctx)
		convey.So(code, convey.ShouldEqual, defaultErrorCode)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
