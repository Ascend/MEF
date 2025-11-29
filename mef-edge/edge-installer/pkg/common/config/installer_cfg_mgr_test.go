// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package config test for installer config manager
package config

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

func TestSetNetManager(t *testing.T) {
	netConfig := NetManager{
		NetType: constants.FD,
		WithOm:  true,
	}

	convey.Convey("set net manager should be success", t, func() {
		err := SetNetManager(&dbMgr, &netConfig)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("set net manager should be failed, net config is nil", t, func() {
		err := SetNetManager(&dbMgr, nil)
		convey.So(err, convey.ShouldResemble, errors.New("set net manager failed, input is nil"))
	})
}

func TestGetNetManager(t *testing.T) {
	convey.Convey("get net manager should be success", t, func() {
		netMgr, err := GetNetManager(&dbMgr)
		convey.So(netMgr.NetType, convey.ShouldEqual, constants.FD)
		convey.So(netMgr.WithOm, convey.ShouldEqual, true)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("set net manager should be failed, net config is nil", t, func() {
		var c *DbMgr
		var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetConfig",
			func(*DbMgr, string, interface{}) error {
				return testErr
			})
		defer p1.Reset()
		netCfg, err := GetNetManager(&dbMgr)
		convey.So(netCfg, convey.ShouldBeNil)
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func TestSetInstall(t *testing.T) {
	installCfg := InstallerConfig{
		InstallDir:    "",
		LogPath:       "",
		LogBackupPath: "",
		SerialNumber:  "abcde",
	}

	convey.Convey("set install config should be success", t, func() {
		err := SetInstall(&dbMgr, &installCfg)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestGetInstall(t *testing.T) {
	convey.Convey("set install config should be success", t, func() {
		installCfg, err := GetInstall(&dbMgr)
		convey.So(installCfg.SerialNumber, convey.ShouldEqual, "abcde")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("set install config should be failed", t, func() {
		var c *DbMgr
		var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetConfig",
			func(*DbMgr, string, interface{}) error {
				return testErr
			})
		defer p1.Reset()
		installCfg, err := GetInstall(&dbMgr)
		convey.So(installCfg, convey.ShouldBeNil)
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func TestGetNodeId(t *testing.T) {
	convey.Convey("get node id should be success", t, func() {
		sn := GetNodeId(&dbMgr)
		convey.So(sn, convey.ShouldEqual, "abcde")
	})

	convey.Convey("get node id should be failed", t, func() {
		var c *DbMgr
		var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetConfig",
			func(*DbMgr, string, interface{}) error {
				return testErr
			})
		defer p1.Reset()
		sn := GetNodeId(&dbMgr)
		convey.So(sn, convey.ShouldEqual, "")
	})
}

func TestCheckIsA500(t *testing.T) {
	convey.Convey("test func CheckIsA500 success", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, constants.A500Name, nil)
		defer p1.Reset()
		convey.So(CheckIsA500(), convey.ShouldBeTrue)
	})

	convey.Convey("test func CheckIsA500 failed, run command failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
		defer p1.Reset()
		convey.So(CheckIsA500(), convey.ShouldBeFalse)
	})
}
