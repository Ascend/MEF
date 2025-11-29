// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package flows for testing offline upgrade
package flows

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestOfflineUpgradeInstaller(t *testing.T) {
	convey.Convey("test get offline upgrade flow", t, func() {
		upgradeFlow := OfflineUpgradeInstaller(OfflineUpgradeInstallerParam{
			EdgeDir:     testDir,
			DelayEffect: false,
		})
		convey.So(upgradeFlow, convey.ShouldNotBeNil)
	})
}
