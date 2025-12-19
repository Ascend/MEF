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

// Package config test for config

package config

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"
)

var (
	domainCfg = []DomainConfig{
		{
			Domain: "fd",
			IP:     "127.0.0.1",
		},
	}
	domainCfgs = DomainConfigs{Configs: domainCfg}
)

func TestSetDomainCfg(t *testing.T) {
	convey.Convey("set domain config should be success", t, testSetDomainCfg)
	convey.Convey("set domain config should be failed", t, testSetDomainCfgErr)
}

func testSetDomainCfg() {
	var p1 = gomonkey.ApplyFuncReturn(GetComponentDbMgr, &dbMgr, nil)
	defer p1.Reset()
	err := SetDomainCfg(&domainCfgs)
	convey.So(err, convey.ShouldBeNil)
}

func testSetDomainCfgErr() {
	var p1 = gomonkey.ApplyFuncReturn(GetComponentDbMgr, &dbMgr, test.ErrTest)
	defer p1.Reset()

	err := SetDomainCfg(&domainCfgs)
	convey.So(err, convey.ShouldResemble, testErr)
}

func TestGetDomainCfg(t *testing.T) {
	convey.Convey("get domain config should be success", t, testGetDomainCfg)
	convey.Convey("get domain config should be failed", t, testGetDomainCfgErr)
}

func testGetDomainCfg() {
	var p1 = gomonkey.ApplyFuncReturn(GetComponentDbMgr, &dbMgr, nil)
	defer p1.Reset()
	cfgs, err := GetDomainCfg()
	convey.So(cfgs.Configs[0].Domain, convey.ShouldEqual, "fd")
	convey.So(cfgs.Configs[0].IP, convey.ShouldEqual, "127.0.0.1")
	convey.So(err, convey.ShouldBeNil)
}

func testGetDomainCfgErr() {
	var p1 = gomonkey.ApplyFuncReturn(GetComponentDbMgr, &dbMgr, test.ErrTest)
	defer p1.Reset()
	cfgs, err := GetDomainCfg()
	convey.So(cfgs, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, testErr)
}

func TestSetImageCfg(t *testing.T) {
	convey.Convey("set image config should be success", t, testSetImageCfg)
	convey.Convey("set image config should be failed", t, testSetImageCfgErr)
}

func testSetImageCfg() {
	var p1 = gomonkey.ApplyFuncReturn(GetComponentDbMgr, &dbMgr, nil)
	defer p1.Reset()
	imgCfg := ImageConfig{ImageAddress: "127.0.0.1"}
	err := SetImageCfg(&imgCfg)
	convey.So(err, convey.ShouldBeNil)
}

func testSetImageCfgErr() {
	var p1 = gomonkey.ApplyFuncReturn(GetComponentDbMgr, &dbMgr, test.ErrTest)
	defer p1.Reset()
	imgCfg := ImageConfig{ImageAddress: "127.0.0.1"}
	err := SetImageCfg(&imgCfg)
	convey.So(err, convey.ShouldResemble, testErr)
}

func TestGetImageCfg(t *testing.T) {
	convey.Convey("get image config should be success", t, testGetImageCfg)
	convey.Convey("get image config should be failed", t, testGetImageCfgErr)
}

func testGetImageCfg() {
	var p1 = gomonkey.ApplyFuncReturn(GetComponentDbMgr, &dbMgr, nil)
	defer p1.Reset()
	cfgs, err := GetImageCfg()
	convey.So(cfgs.ImageAddress, convey.ShouldEqual, "127.0.0.1")
	convey.So(err, convey.ShouldBeNil)
}

func testGetImageCfgErr() {
	var p1 = gomonkey.ApplyFuncReturn(GetComponentDbMgr, &dbMgr, test.ErrTest)
	defer p1.Reset()
	cfgs, err := GetImageCfg()
	convey.So(cfgs, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, testErr)
}
