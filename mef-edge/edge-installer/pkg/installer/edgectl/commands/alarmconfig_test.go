// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package commands test for alarm config
package commands

import (
	"errors"
	"flag"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

func TestUpdateAlarmConfig(t *testing.T) {
	var p = gomonkey.ApplyMethodReturn(&config.DbMgr{}, "SetAlarmConfig", nil)
	defer p.Reset()
	convey.Convey("test alarm config cmd methods", t, alarmCfgCmdMethods)
	convey.Convey("test alarm config cmd execute success", t, alarmCfgCmdExecute)
	convey.Convey("test alarm config cmd failed", t, func() {
		convey.Convey("ctx is nil", alarmCfgCmdExecuteErrCtx)
		convey.Convey("does not modify any configuration", alarmCfgCmdNoParam)
		convey.Convey("get alarm config mgr failed", alarmCfgCmdErrGetAlarmCgfMgr)
		convey.Convey("param error", alarmCfgCmdErrCheckParam)
		convey.Convey("update error", alarmCfgCmdErrUpdate)
	})
}

func alarmCfgCmdMethods() {
	convey.So(AlarmConfigCmd().Name(), convey.ShouldEqual, common.AlarmConfig)
	convey.So(AlarmConfigCmd().Description(), convey.ShouldEqual, common.AlarmConfigDesc)
	convey.So(AlarmConfigCmd().LockFlag(), convey.ShouldBeTrue)
	AlarmConfigCmd().PrintOpLogOk(userRoot, ipLocalhost)
	AlarmConfigCmd().PrintOpLogFail(userRoot, ipLocalhost)
}

func alarmCfgCmdExecute() {
	cmd = AlarmConfigCmd()
	cmd.BindFlag()
	if err := flag.Set(common.CertCheckPeriodCmd, "7"); err != nil {
		fmt.Printf("test set alarm config flag failed, error: %v\n", err)
		return
	}
	if err := flag.Set(common.CertOverdueThresholdCmd, "90"); err != nil {
		fmt.Printf("test set alarm config flag failed, error: %v\n", err)
		return
	}
	flag.Parse()

	err := cmd.Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
}

func alarmCfgCmdExecuteErrCtx() {
	err := cmd.Execute(nil)
	expectErr := errors.New("ctx is nil")
	convey.So(err, convey.ShouldResemble, expectErr)
}

func alarmCfgCmdNoParam() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsFlagSet, false)
	defer p1.Reset()
	err := cmd.Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
}

func alarmCfgCmdErrGetAlarmCgfMgr() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, test.ErrTest)
	defer p1.Reset()
	err := cmd.Execute(ctx)
	expectErr := errors.New("get config path manager failed")
	convey.So(err, convey.ShouldResemble, expectErr)
}

func alarmCfgCmdErrCheckParam() {
	alarmCfgCmd := alarmConfigCmd{
		certCheckPeriod:      200,
		certOverdueThreshold: 90,
	}
	err := alarmCfgCmd.checkParam()
	expectErr := fmt.Errorf("param %s is invalid", common.CertOverdueThresholdCmd)
	convey.So(err, convey.ShouldResemble, expectErr)

	alarmCfgCmd.certCheckPeriod = 7
	alarmCfgCmd.certOverdueThreshold = 5
	err = alarmCfgCmd.checkParam()
	expectErr = fmt.Errorf("param %s is invalid", common.CertOverdueThresholdCmd)
	convey.So(err, convey.ShouldResemble, expectErr)

	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{0, test.ErrTest}},

		{Values: gomonkey.Params{7, nil}},
		{Values: gomonkey.Params{0, test.ErrTest}},

		{Values: gomonkey.Params{7, nil}},
		{Values: gomonkey.Params{4, nil}},
	}
	var p1 = gomonkey.ApplyMethodSeq(&config.DbMgr{}, "GetAlarmConfig", outputs).
		ApplyFuncReturn(util.IsFlagSet, false)
	defer p1.Reset()
	err = alarmCfgCmd.checkParam()
	convey.So(err, convey.ShouldResemble, test.ErrTest)
	err = alarmCfgCmd.checkParam()
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func alarmCfgCmdErrUpdate() {
	var p1 = gomonkey.ApplyMethodReturn(&config.DbMgr{}, "SetAlarmConfig", test.ErrTest)
	defer p1.Reset()

	alarmCfgCmd := alarmConfigCmd{
		certCheckPeriod:      7,
		certOverdueThreshold: 90,
	}
	err := alarmCfgCmd.updateConfig()
	convey.So(err.Error(), convey.ShouldNotBeNil)
}
