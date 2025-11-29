// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package flows for testing online upgrade
package flows

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/installer/preupgrade/tasks"
)

func TestOnlineUpgradeInstaller(t *testing.T) {
	convey.Convey("test get online upgrade flow", t, func() {
		upgradeFlow := OnlineUpgradeInstaller(testDir)
		convey.So(upgradeFlow, convey.ShouldNotBeNil)
	})

	convey.Convey("test online upgrade methods", t, func() {
		upgrade := onlineUpgradeInstaller{}
		convey.Convey("test print operator log", func() {
			p := gomonkey.ApplyGlobalVar(&config.NetMgr, config.NetManager{NetType: "MEF", IP: "127.0.0.1"})
			defer p.Reset()
			upgrade.onlineUpgradeOpLogOk()
			upgrade.onlineUpgradeOpLogFailed()
		})

		convey.Convey("test effect", func() {
			p := gomonkey.ApplyMethodReturn(tasks.EffectInstaller(testDir), "Run", test.ErrTest)
			defer p.Reset()
			upgrade.effect()
		})
	})
}
