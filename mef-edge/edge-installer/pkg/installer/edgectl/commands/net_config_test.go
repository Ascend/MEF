// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package commands
package commands

import (
	"errors"
	"flag"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
)

func TestNetConfigCmd(t *testing.T) {
	convey.Convey("test net config cmd methods", t, netConfigCmdMethods)
	convey.Convey("test net config cmd successful", t, netConfigCmdSuccess)
	convey.Convey("test net config cmd failed", t, func() {
		convey.Convey("execute net config failed", executeNetCfgFailed)
		convey.Convey("necessary param check failed", necessaryParamCheckFailed)
		convey.Convey("set net type failed", setNetTypeFailed)
	})
}

func netConfigCmdMethods() {
	convey.So(NetConfigCmd().Name(), convey.ShouldEqual, common.NetConfig)
	convey.So(NetConfigCmd().Description(), convey.ShouldEqual, common.NetConfigDesc)
	convey.So(NetConfigCmd().LockFlag(), convey.ShouldBeTrue)
}

func netConfigCmdSuccess() {
	cmd := NetConfigCmd()
	cmd.BindFlag()
	if err := flag.Set("net_type", constants.FDWithOM); err != nil {
		hwlog.RunLog.Errorf("test set flag net_type failed, error: %v", err)
		return
	}
	flag.Parse()
	p := gomonkey.ApplyFuncReturn(common.InitEdgeOmResource, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
		ApplyFuncReturn(envutils.RunCommandWithUser, nil, nil)
	defer p.Reset()
	err := cmd.Execute(ctx)
	expectErr := fmt.Errorf("net config failed, error: type of net manager only support [%s]", constants.MEF)
	convey.So(err, convey.ShouldResemble, expectErr)
	NetConfigCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeNetCfgFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := NetConfigCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("net config failed", func() {
		if err := flag.Set("net_type", constants.FDWithOM); err != nil {
			hwlog.RunLog.Errorf("test set flag net_type failed, error: %v", err)
			return
		}
		p := gomonkey.ApplyFuncReturn(NetConfigCmd, &netConfigCmd{netType: constants.FDWithOM}).
			ApplyFuncReturn(common.InitEdgeOmResource, nil)
		defer p.Reset()
		err := NetConfigCmd().Execute(ctx)
		expectErr := fmt.Errorf("net config failed, error: type of net manager only support [%s]", constants.MEF)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	NetConfigCmd().PrintOpLogFail(userRoot, ipLocalhost)
}

func necessaryParamCheckFailed() {
	convey.Convey("type of net manager not support failed", func() {
		if err := flag.Set("net_type", constants.FD); err != nil {
			hwlog.RunLog.Errorf("test set flag net_type failed, error: %v", err)
			return
		}
		err := NetConfigCmd().Execute(ctx)
		expectErr := fmt.Errorf("type of net manager only support [%s]", constants.MEF)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("net config failed, error: %v", expectErr))
	})

	convey.Convey("param ip is empty failed", func() {
		if err := flag.Set("net_type", constants.MEF); err != nil {
			hwlog.RunLog.Errorf("test set flag net_type failed, error: %v", err)
			return
		}
		err := NetConfigCmd().Execute(ctx)
		expectErr := errors.New("param ip is necessary, can not be empty")
		convey.So(err, convey.ShouldResemble, fmt.Errorf("net config failed, error: %v", expectErr))
	})

	convey.Convey("param root_ca is empty failed", func() {
		if err := flag.Set("net_type", constants.MEF); err != nil {
			hwlog.RunLog.Errorf("test set flag net_type failed, error: %v", err)
			return
		}
		if err := flag.Set("ip", "127.0.0.1"); err != nil {
			hwlog.RunLog.Errorf("test set flag ip failed, error: %v", err)
			return
		}
		err := NetConfigCmd().Execute(ctx)
		expectErr := errors.New("param root_ca is necessary, can not be empty")
		convey.So(err, convey.ShouldResemble, fmt.Errorf("net config failed, error: %v", expectErr))
	})
}

func setNetTypeFailed() {
	if err := flag.Set("net_type", constants.MEF); err != nil {
		hwlog.RunLog.Errorf("test set flag net_type failed, error: %v", err)
		return
	}
	if err := flag.Set("ip", ""); err != nil {
		hwlog.RunLog.Errorf("test set flag ip failed, error: %v", err)
		return
	}
	p := gomonkey.ApplyFuncReturn(NetConfigCmd, &netConfigCmd{netType: constants.MEF}).
		ApplyFuncReturn(common.InitEdgeOmResource, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
		ApplyFuncReturn(envutils.RunCommandWithUser, nil, test.ErrTest)
	defer p.Reset()
	err := NetConfigCmd().Execute(ctx)
	expectErr := errors.New("net config failed, error: param ip is necessary, can not be empty")
	convey.So(err, convey.ShouldResemble, expectErr)
}
