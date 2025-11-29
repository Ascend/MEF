// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package flows for testing effect flow
package flows

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/upgrade/tasks"
)

var (
	testEffectDir = "/tmp/test_effect_flow_dir"
	flowEffect    = NewEffectFlow(pathmgr.NewPathMgr(testEffectDir, testEffectDir, testEffectDir, testEffectDir))
)

func TestEffectFlow(t *testing.T) {
	convey.Convey("effect flow should be success", t, effectFlowSuccess)
	convey.Convey("effect flow should be failed", t, effectFlowFailed)
}

func effectFlowSuccess() {
	p := gomonkey.ApplyMethodReturn(&tasks.PostEffectProcessTask{}, "Run", nil)
	defer p.Reset()
	err := flowEffect.RunTasks()
	convey.So(err, convey.ShouldBeNil)
}

func effectFlowFailed() {
	p := gomonkey.ApplyMethodReturn(&tasks.PostEffectProcessTask{}, "Run", test.ErrTest)
	defer p.Reset()
	err := flowEffect.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("upgrade post process task failed"))
}
