// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package flows for testing upgrade flow
package flows

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
	commonTasks "edge-installer/pkg/installer/common/tasks"
	"edge-installer/pkg/installer/upgrade/tasks"
)

var (
	testUpgradeDir = "/tmp/test_upgrade_flow_dir"
	flowUpgrade    = NewUpgradeFlow(pathmgr.NewPathMgr(testUpgradeDir, testUpgradeDir, testUpgradeDir, testUpgradeDir))
)

func TestUpgradeFlow(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(&commonTasks.CheckParamTask{}, "Run", nil).
		ApplyMethodReturn(&tasks.CheckUpgradeEnvironmentTask{}, "Run", nil).
		ApplyMethodReturn(&tasks.SetWorkPathTask{}, "Run", nil).
		ApplyMethodReturn(&commonTasks.InstallComponentsTask{}, "Run", nil)
	defer p.Reset()

	convey.Convey("upgrade flow should be success", t, upgradeFlowSuccess)
	convey.Convey("upgrade flow should be failed, check upgrade param failed", t, checkUpgradeParamFailed)
	convey.Convey("upgrade flow should be failed, check upgrade environment failed", t, checkUpgradeEnvFailed)
	convey.Convey("upgrade flow should be failed, set work path failed", t, setWorkPathFailed)
	convey.Convey("upgrade flow should be failed, install components failed", t, installComponentsFailed)
}

func upgradeFlowSuccess() {
	err := flowUpgrade.RunTasks()
	convey.So(err, convey.ShouldBeNil)
}

func checkUpgradeParamFailed() {
	p1 := gomonkey.ApplyMethodReturn(&commonTasks.CheckParamTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := flowUpgrade.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("check upgrade param task failed"))
}

func checkUpgradeEnvFailed() {
	p1 := gomonkey.ApplyMethodReturn(&tasks.CheckUpgradeEnvironmentTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := flowUpgrade.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("check upgrade environment task failed"))
}

func setWorkPathFailed() {
	p1 := gomonkey.ApplyMethodReturn(&tasks.SetWorkPathTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := flowUpgrade.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("set upgrade work path task failed"))
}

func installComponentsFailed() {
	convey.Convey("get work abs dir failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&pathmgr.WorkPathMgr{}, "GetWorkAbsDir", "", test.ErrTest)
		defer p1.Reset()
		err := flowUpgrade.RunTasks()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("install components failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&commonTasks.InstallComponentsTask{}, "Run", test.ErrTest)
		defer p1.Reset()
		err := flowUpgrade.RunTasks()
		convey.So(err, convey.ShouldResemble, errors.New("install components task failed"))
	})
}
