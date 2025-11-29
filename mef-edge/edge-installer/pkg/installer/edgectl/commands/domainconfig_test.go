// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package commands
package commands

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/domainconfig"
)

var domain = "fd.fusiondirector.huawei.com"

func TestDomainConfigCmd(t *testing.T) {
	convey.Convey("test domain config cmd methods", t, domainConfigCmdMethods)
	convey.Convey("test domain config cmd successful", t, domainConfigCmdSuccess)
	convey.Convey("test domain config cmd failed", t, func() {
		convey.Convey("init edge om resource failed", initEdgeOmResourceFailed)
		convey.Convey("run domain config flow failed", domainCfgFlowFailed)
	})
}

func domainConfigCmdMethods() {
	convey.So(DomainConfigCmd().Name(), convey.ShouldEqual, common.DomainConfig)
	convey.So(DomainConfigCmd().Description(), convey.ShouldEqual, common.DomainConfigDesc)
	convey.So(DomainConfigCmd().LockFlag(), convey.ShouldBeTrue)
}

func domainConfigCmdSuccess() {
	p := gomonkey.ApplyFuncReturn(DomainConfigCmd, &domainConfigCmd{domain: domain}).
		ApplyFuncReturn(common.InitEdgeOmResource, nil).
		ApplyMethodReturn(domainconfig.DomainCfgFlow{}, "RunTasks", nil)
	defer p.Reset()
	err := DomainConfigCmd().Execute(&common.Context{})
	convey.So(err, convey.ShouldBeNil)
	DomainConfigCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func initEdgeOmResourceFailed() {
	p := gomonkey.ApplyFuncReturn(DomainConfigCmd, &domainConfigCmd{domain: domain}).
		ApplyFuncReturn(common.InitEdgeOmResource, test.ErrTest)
	defer p.Reset()
	err := DomainConfigCmd().Execute(&common.Context{})
	expectErr := fmt.Errorf("init resource failed, error: %v", test.ErrTest)
	convey.So(err, convey.ShouldResemble, expectErr)
	DomainConfigCmd().PrintOpLogFail(userRoot, ipLocalhost)
}

func domainCfgFlowFailed() {
	p := gomonkey.ApplyFuncReturn(DomainConfigCmd, &domainConfigCmd{domain: domain}).
		ApplyFuncReturn(common.InitEdgeOmResource, nil).
		ApplyMethodReturn(domainconfig.DomainCfgFlow{}, "RunTasks", test.ErrTest)
	defer p.Reset()
	err := DomainConfigCmd().Execute(&common.Context{})
	expectErr := fmt.Errorf("domain mapping config %s failed, error: %v", domain, test.ErrTest)
	convey.So(err, convey.ShouldResemble, expectErr)
}
