// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package commands test for get alarm config
package commands

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/installer/edgectl/common"
)

func TestGetAlarmConfig(t *testing.T) {
	var p = gomonkey.ApplyMethodReturn(&config.DbMgr{}, "GetAlarmConfig", 0, nil)
	defer p.Reset()
	convey.Convey("test get alarm config cmd methods", t, getAlarmCfgCmdMethods)
	convey.Convey("test get alarm config cmd execute success", t, getAlarmCfgCmdExecute)
	convey.Convey("test get alarm config cmd failed", t, func() {
		convey.Convey("ctx is nil", getAlarmCfgCmdExecuteErrCtx)
		convey.Convey("get alarm config mgr failed", getAlarmCfgCmdExecuteErrGetDbMgr)
		convey.Convey("get alarm config failed", getAlarmCfgCmdExecuteErrGetCfg)
	})
}

func getAlarmCfgCmdMethods() {
	convey.So(GetAlarmCfgCmd().Name(), convey.ShouldEqual, common.GetAlarmCfg)
	convey.So(GetAlarmCfgCmd().Description(), convey.ShouldEqual, common.GetAlarmCfgDesc)
	convey.So(GetAlarmCfgCmd().LockFlag(), convey.ShouldBeTrue)
	GetAlarmCfgCmd().PrintOpLogOk(userRoot, ipLocalhost)
	GetAlarmCfgCmd().PrintOpLogFail(userRoot, ipLocalhost)
}

func getAlarmCfgCmdExecute() {
	cmd = GetAlarmCfgCmd()
	cmd.BindFlag()
	err := cmd.Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
}

func getAlarmCfgCmdExecuteErrCtx() {
	cmd = GetAlarmCfgCmd()
	err := cmd.Execute(nil)
	expectErr := errors.New("ctx is nil")
	convey.So(err, convey.ShouldResemble, expectErr)
}

func getAlarmCfgCmdExecuteErrGetDbMgr() {
	var p1 = gomonkey.ApplyFuncReturn(config.GetComponentDbMgr, nil, test.ErrTest)
	defer p1.Reset()
	cmd = GetAlarmCfgCmd()
	err := cmd.Execute(ctx)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func getAlarmCfgCmdExecuteErrGetCfg() {
	var p1 = gomonkey.ApplyMethodReturn(&config.DbMgr{}, "GetAlarmConfig", 0, test.ErrTest)
	defer p1.Reset()
	cmd = GetAlarmCfgCmd()
	err := cmd.Execute(ctx)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}
