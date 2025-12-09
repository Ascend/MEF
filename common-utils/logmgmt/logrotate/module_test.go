// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
