// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build !MEFEdge_A500

// Package tasks for testing methods that are performed only on non a500 device
package tasks

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
)

func TestPostEffectProcessTaskSmoothConfig(t *testing.T) {
	convey.Convey("smooth config failed, smooth edge_om alarm config to db failed", t, func() {
		p := gomonkey.ApplyPrivateMethod(&PostEffectProcessTask{}, "smoothCommonConfig",
			func(PostEffectProcessTask) error { return nil }).
			ApplyFuncReturn(config.SmoothAlarmConfigDB, test.ErrTest)
		defer p.Reset()
		err := postEffectProcess.smoothConfig()
		convey.So(err, convey.ShouldResemble, errors.New("smooth edge_om alarm config to db failed"))
	})
}
