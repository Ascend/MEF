// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
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
