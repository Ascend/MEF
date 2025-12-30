// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
