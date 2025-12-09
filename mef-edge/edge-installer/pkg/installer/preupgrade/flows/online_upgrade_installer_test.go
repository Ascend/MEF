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
