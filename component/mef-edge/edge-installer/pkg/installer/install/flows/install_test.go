// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package flows for testing installation flow
package flows

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
	commonTasks "edge-installer/pkg/installer/common/tasks"
	"edge-installer/pkg/installer/install/tasks"
)

var (
	testDir        = "/tmp/test_install_flow_dir"
	pathMgr        = pathmgr.NewPathMgr(testDir, testDir, testDir, testDir)
	workAbsPathMgr = pathmgr.NewWorkAbsPathMgr(testDir)
	installFlow    = NewInstallFlow(pathMgr, workAbsPathMgr, true)
)

func TestFlowInstall(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(&commonTasks.CheckParamTask{}, "Run", nil).
		ApplyMethodReturn(&tasks.CheckInstallEnvironmentTask{}, "Run", nil).
		ApplyMethodReturn(&tasks.AddUserAccountTask{}, "Run", nil).
		ApplyMethodReturn(&tasks.SetWorkPathTask{}, "Run", nil).
		ApplyMethodReturn(&tasks.ConfigComponentsTask{}, "Run", nil).
		ApplyMethodReturn(&commonTasks.InstallComponentsTask{}, "Run", nil).
		ApplyMethodReturn(&common.GenerateCertsTask{}, "Run", nil).
		ApplyMethodReturn(&commonTasks.SetSystemInfoTask{}, "Run", nil).
		ApplyMethodReturn(&tasks.PostInstallProcessTask{}, "Run", nil).
		ApplyFuncReturn(util.GetMefId, uint32(constants.EdgeUserUid), uint32(constants.EdgeUserGid), nil)
	defer p.Reset()

	convey.Convey("install flow should be success", t, installFlowSuccess)
	convey.Convey("install flow should be failed, check install param failed", t, checkInstallParamFailed)
	convey.Convey("install flow should be failed, check install environment failed", t, checkInstallEnvFailed)
	convey.Convey("install flow should be failed, add user account failed", t, addUserAccountFailed)
	convey.Convey("install flow should be failed, set work path failed", t, setWorkPathFailed)
	convey.Convey("install flow should be failed, config components failed", t, configComponentsFailed)
	convey.Convey("install flow should be failed, install components failed", t, installComponentsFailed)
	convey.Convey("install flow should be failed, generate certs failed", t, generateCertsFailed)
}

func installFlowSuccess() {
	err := installFlow.RunTasks()
	convey.So(err, convey.ShouldBeNil)
}

func checkInstallParamFailed() {
	p1 := gomonkey.ApplyMethodReturn(&commonTasks.CheckParamTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := installFlow.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("check install param task failed"))
}

func checkInstallEnvFailed() {
	p1 := gomonkey.ApplyMethodReturn(&tasks.CheckInstallEnvironmentTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := installFlow.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("check install environment task failed"))
}

func addUserAccountFailed() {
	p1 := gomonkey.ApplyMethodReturn(&tasks.AddUserAccountTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := installFlow.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("add user account task failed"))
}

func setWorkPathFailed() {
	p1 := gomonkey.ApplyMethodReturn(&tasks.SetWorkPathTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := installFlow.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("set work path task failed"))
}

func configComponentsFailed() {
	p1 := gomonkey.ApplyMethodReturn(&tasks.ConfigComponentsTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := installFlow.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("config components task failed"))
}

func installComponentsFailed() {
	p1 := gomonkey.ApplyMethodReturn(&commonTasks.InstallComponentsTask{}, "Run", test.ErrTest)
	defer p1.Reset()
	err := installFlow.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("install components task failed"))
}

func generateCertsFailed() {
	p1 := gomonkey.ApplyFuncReturn(util.GetMefId, uint32(0), uint32(0), test.ErrTest)
	defer p1.Reset()
	err := installFlow.RunTasks()
	convey.So(err, convey.ShouldResemble, errors.New("get generate certs task failed"))
}
