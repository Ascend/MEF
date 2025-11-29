// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing config components task
package tasks

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/common/components"
)

var configComponents = ConfigComponentsTask{PathMgr: pathMgr, WorkAbsPathMgr: workAbsPathMgr}

func TestConfigComponentsTask(t *testing.T) {
	convey.Convey("config components should be success", t, configComponentsSuccess)
	convey.Convey("config components should be failed", t, configComponentsFailed)
}

func configComponentsSuccess() {
	var p = gomonkey.ApplyMethodReturn(&components.PrepareInstaller{}, "PrepareCfgDir", nil).
		ApplyMethodReturn(&components.PrepareEdgeOm{}, "PrepareCfgDir", nil).
		ApplyMethodReturn(&components.PrepareEdgeMain{}, "PrepareCfgDir", nil).
		ApplyMethodReturn(&components.PrepareEdgeCore{}, "PrepareCfgDir", nil)
	defer p.Reset()

	err := configComponents.Run()
	convey.So(err, convey.ShouldBeNil)
}

func configComponentsFailed() {
	var p = gomonkey.ApplyMethodReturn(&components.PrepareInstaller{}, "PrepareCfgDir", test.ErrTest)
	defer p.Reset()

	err := configComponents.Run()
	expectErr := fmt.Errorf("prepare [%s] config directories failed", constants.EdgeInstaller)
	convey.So(err, convey.ShouldResemble, expectErr)
}
