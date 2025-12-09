// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for testing components installation task
package tasks

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common/components"
)

var installComponents = InstallComponentsTask{
	PathMgr:        &pathmgr.PathManager{},
	WorkAbsPathMgr: &pathmgr.WorkAbsPathMgr{},
}

func TestInstallComponentsTask(t *testing.T) {
	convey.Convey("install components should be success", t, installComponentsSuccess)
	convey.Convey("install components should be failed, install component failed", t, installComponentsFailed)
}

func installComponentsSuccess() {
	var p = gomonkey.ApplyMethodReturn(&components.PrepareInstaller{}, "Run", nil).
		ApplyMethodReturn(&components.PrepareEdgeOm{}, "Run", nil).
		ApplyMethodReturn(&components.PrepareEdgeMain{}, "Run", nil).
		ApplyMethodReturn(&components.PrepareEdgeCore{}, "Run", nil).
		ApplyMethodReturn(&components.PrepareDevicePlugin{}, "Run", nil)
	defer p.Reset()

	err := installComponents.Run()
	convey.So(err, convey.ShouldBeNil)
}

func installComponentsFailed() {
	var p = gomonkey.ApplyMethodReturn(&components.PrepareInstaller{}, "Run", test.ErrTest)
	defer p.Reset()

	err := installComponents.Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("install component [%s] failed", constants.EdgeInstaller))
}
