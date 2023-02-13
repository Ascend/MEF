// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logrotate provides log rotation function for third-party software
package logrotate

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestInitializer test initializer
func TestInitializer(t *testing.T) {
	convey.Convey("TestInitializer", t, func() {
		m := Module("my-module", context.Background(), func() (Configs, error) {
			return Configs{CheckIntervalSeconds: 1}, nil
		})

		module, ok := m.(*logRotatorModule)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(module.rotator.configs.CheckIntervalSeconds, convey.ShouldBeZeroValue)
		convey.So(module.Enable(), convey.ShouldBeTrue)
		convey.So(module.rotator.configs.CheckIntervalSeconds, convey.ShouldEqual, 1)
	})
}
