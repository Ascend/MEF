// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package uninstall
package uninstall

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
)

var (
	testPath      = "/tmp/test_uninstall/MEFEdge"
	workPathMgr   = pathmgr.NewWorkPathMgr(filepath.Dir(testPath))
	configPathMgr = pathmgr.NewConfigPathMgr(filepath.Dir(testPath))
	testErr       = errors.New("test error")
)

func setup() error {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := util.InitHwLogger(logConfig, logConfig); err != nil {
		return err
	}

	if err := fileutils.CreateDir(testPath, constants.Mode700); err != nil {
		return fmt.Errorf("create test path failed, error: %v", err)
	}
	return nil
}

func teardown() {
	if err := os.RemoveAll(testPath); err != nil {
		hwlog.RunLog.Errorf("remove test path failed, error: %v", err)
	}
}

// TestMain run test main
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Printf("setup test environment failed: %v\n", err)
		return
	}
	defer teardown()
	exitCode := m.Run()
	fmt.Printf("test complete, exitCode=%d\n", exitCode)
}

func TestFlowUninstall(t *testing.T) {
	convey.Convey("test uninstall successful", t, doUninstallSuccess)
	convey.Convey("test uninstall failed", t, func() {
		convey.Convey("remove service failed", removeServiceFailed)
		convey.Convey("remove external files failed", removeExternalFilesFailed)
		convey.Convey("remove container failed", removeContainerFailed)
		convey.Convey("remove install dir failed", removeInstallDirFailed)
	})
}

func doUninstallSuccess() {
	p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
		ApplyMethodReturn(common.ComponentMgr{}, "UnregisterAllServices", nil).
		ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncReturn(util.RemoveContainer, nil)
	defer p.Reset()
	uninstallFlow := NewFlowUninstall(workPathMgr, configPathMgr)
	err := uninstallFlow.RunTasks()
	convey.So(err, convey.ShouldBeNil)
}

func removeServiceFailed() {
	convey.Convey("stop all services failed", func() {
		p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", testErr)
		defer p.Reset()
		uninstallFlow := NewFlowUninstall(workPathMgr, configPathMgr)
		err := uninstallFlow.RunTasks()
		expectErr := errors.New("stop all services failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("remove all services failed", func() {
		p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
			ApplyMethodReturn(common.ComponentMgr{}, "UnregisterAllServices", testErr)
		defer p.Reset()
		uninstallFlow := NewFlowUninstall(workPathMgr, configPathMgr)
		err := uninstallFlow.RunTasks()
		expectErr := errors.New("remove all services failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func removeExternalFilesFailed() {
	p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
		ApplyMethodReturn(common.ComponentMgr{}, "UnregisterAllServices", nil).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, testErr)
	defer p.Reset()
	uninstallFlow := NewFlowUninstall(workPathMgr, configPathMgr)
	err := uninstallFlow.RunTasks()
	expectErr := fmt.Errorf("remove [%s] failed", constants.PreUpgradePath)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func removeContainerFailed() {
	p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
		ApplyMethodReturn(common.ComponentMgr{}, "UnregisterAllServices", nil).
		ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncReturn(util.RemoveContainer, testErr)
	defer p.Reset()
	uninstallFlow := NewFlowUninstall(workPathMgr, configPathMgr)
	err := uninstallFlow.RunTasks()
	convey.So(err, convey.ShouldResemble, testErr)
}

func removeInstallDirFailed() {
	p := gomonkey.ApplyMethodReturn(common.ComponentMgr{}, "StopAll", nil).
		ApplyMethodReturn(common.ComponentMgr{}, "UnregisterAllServices", nil).
		ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncReturn(util.RemoveContainer, nil).
		ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, testErr)
	defer p.Reset()
	uninstallFlow := NewFlowUninstall(workPathMgr, configPathMgr)
	err := uninstallFlow.RunTasks()
	expectErr := fmt.Errorf("remove software dir [%s] failed", testPath)
	convey.So(err, convey.ShouldResemble, expectErr)
}
