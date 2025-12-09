// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common test for component and component mgr
package common

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

func TestCheckAllServiceActive(t *testing.T) {
	convey.Convey("check all service active should be failed, component does not exist", t, func() {
		var p = gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
		defer p.Reset()
		componentMgr.CheckAllServiceActive()
	})

	convey.Convey("check all service active should be success", t, func() {
		var p = gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(util.IsServiceActive, true)
		defer p.Reset()
		componentMgr.CheckAllServiceActive()
	})
}

func TestStartComponentMgr(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(util.IsServiceActive, false).
		ApplyFuncReturn(util.CopyServiceFileToSystemd, nil).
		ApplyFuncReturn(util.EnableService, nil).
		ApplyFuncReturn(util.StartService, nil)
	defer p.Reset()

	convey.Convey("test start component mgr should be success", t, testStartComponentMgr)
	convey.Convey("test start component mgr should be failed, get comp name error", t, testStartComponentMgrErrGetName)
	convey.Convey("test start component mgr should be success, svc is active", t, testStartComponentMgrIsSvcActive)
	convey.Convey("test start component mgr should be failed, cp svc file error", t, testStartComponentMgrErrCpSvcFile)
	convey.Convey("test start component mgr should be failed, start svc error", t, testStartComponentMgrErrStartSvc)
}

func testStartComponentMgr() {
	err := componentMgr.Start(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldBeNil)
}

func testStartComponentMgrErrGetName() {
	var errName = ""
	err := componentMgr.Start(errName)
	innerErr := fmt.Errorf("component [%s] does not exist", errName)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get component failed, error: %v", innerErr))
}

func testStartComponentMgrIsSvcActive() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceActive, true)
	defer p1.Reset()

	err := componentMgr.Start(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldBeNil)
}

func testStartComponentMgrErrCpSvcFile() {
	var p1 = gomonkey.ApplyFuncReturn(util.CopyServiceFileToSystemd, test.ErrTest)
	defer p1.Reset()

	svcPath := componentMgr.GetEdgeMain().Service.Path
	err := componentMgr.Start(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("copy service file [%s] failed", svcPath))
}

func testStartComponentMgrErrStartSvc() {
	var p1 = gomonkey.ApplyFuncReturn(util.StartService, test.ErrTest)
	defer p1.Reset()

	err := componentMgr.Start(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func TestStopComponentMgr(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(util.IsServiceInSystemd, true).
		ApplyFuncReturn(util.StopService, nil).
		ApplyFuncReturn(util.IsServiceActive, true)
	defer p.Reset()

	convey.Convey("test stop component mgr should be success", t, testStopComponentMgr)
	convey.Convey("test stop component mgr should be failed, get comp name error", t, testStopComponentMgrErrGetName)
	convey.Convey("test stop component mgr should be success, svc is not in systemd", t, testStopComponentMgrIsSvcSystemd)
	convey.Convey("test stop component mgr should be failed, stop svc error", t, testStopComponentMgrErrStopSvc)
	convey.Convey("test stop component mgr should be success, svc is active", t, testStopComponentMgrIsSvcActive)
}

func testStopComponentMgr() {
	err := componentMgr.Stop(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldBeNil)
}

func testStopComponentMgrErrGetName() {
	var errName = ""
	err := componentMgr.Stop(errName)
	innerErr := fmt.Errorf("component [%s] does not exist", errName)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get component failed, error: %v", innerErr))
}

func testStopComponentMgrIsSvcSystemd() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceInSystemd, false)
	defer p1.Reset()

	err := componentMgr.Stop(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldBeNil)
}

func testStopComponentMgrErrStopSvc() {
	var p1 = gomonkey.ApplyFuncReturn(util.StopService, test.ErrTest)
	defer p1.Reset()

	err := componentMgr.Stop(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldResemble, errors.New("stop service failed"))
}

func testStopComponentMgrIsSvcActive() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceActive, true)
	defer p1.Reset()

	err := componentMgr.Start(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldBeNil)
}

func TestRestartComponentMgr(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(util.CopyServiceFileToSystemd, nil).
		ApplyFuncReturn(util.EnableService, nil).
		ApplyFuncReturn(util.RestartService, nil).
		ApplyFuncReturn(util.IsServiceActive, true)
	defer p.Reset()

	convey.Convey("test restart component mgr should be success", t, testRestartComponentMgr)
	convey.Convey("test restart component mgr should be failed, get comp name error", t, testRestartCompMgrErrGetName)
	convey.Convey("test restart component mgr should be failed, cp svc file error", t, testRestartCompMgrErrCpSvcFile)
	convey.Convey("test restart component mgr should be failed, restart error", t, testRestartComponentMgrErrRestartSvc)
	convey.Convey("test restart component mgr should be success, svc is active", t, testRestartComponentMgrIsSvcActive)
}

func testRestartComponentMgr() {
	err := componentMgr.Restart(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldBeNil)
}

func testRestartCompMgrErrGetName() {
	var errName = ""
	err := componentMgr.Restart(errName)
	innerErr := fmt.Errorf("component [%s] does not exist", errName)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get component failed, error: %v", innerErr))
}

func testRestartCompMgrErrCpSvcFile() {
	var p1 = gomonkey.ApplyFuncReturn(util.CopyServiceFileToSystemd, test.ErrTest)
	defer p1.Reset()

	svcName := constants.EdgeMainServiceFile
	err := componentMgr.Restart(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("register service [%s] failed", svcName))
}

func testRestartComponentMgrErrRestartSvc() {
	var p1 = gomonkey.ApplyFuncReturn(util.RestartService, test.ErrTest)
	defer p1.Reset()

	svcName := constants.EdgeMainServiceFile
	err := componentMgr.Restart(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("restart service [%s] failed", svcName))
}

func testRestartComponentMgrIsSvcActive() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceActive, false)
	defer p1.Reset()

	err := componentMgr.Restart(constants.EdgeMainFileName)
	convey.So(err, convey.ShouldBeNil)
}

func TestStartAllComponentMgr(t *testing.T) {
	// ci environment is x86 sys, uid of /usr/bin/docker is 1000, so stub related function util.CheckNecessaryCommands
	var p = gomonkey.ApplyFuncReturn(SetNodeIPToEdgeCore, nil).
		ApplyFuncReturn(util.GetMefId, uint32(0), uint32(0), nil).
		ApplyMethodReturn(&GenerateCertsTask{}, "MakeSureEdgeCerts", nil).
		ApplyFuncReturn(config.CheckIsA500, false).
		ApplyFuncReturn(util.CopyServiceFileToSystemd, nil).
		ApplyFuncReturn(util.StartService, nil).
		ApplyFuncReturn(util.RestartService, nil).
		ApplyFuncReturn(util.CheckNecessaryCommands, nil)
	defer p.Reset()

	convey.Convey("test start and restart all comp mgr should be success", t, testStartAllComponentMgr)
	convey.Convey("test start and restart all comp mgr should be failed, get dir error", t, testStartAllCompMgrErrGetDir)
	convey.Convey("test start and restart all comp mgr should be failed, cp file error", t, testStartAllCompMgrErrCpFile)
	convey.Convey("test start and restart all comp mgr should be failed, start error", t, testStartAllCompMgrErrStart)
}

func testStartAllComponentMgr() {
	err := componentMgr.StartAll()
	convey.So(err, convey.ShouldBeNil)

	err = componentMgr.RestartAll()
	convey.So(err, convey.ShouldBeNil)
}

func testStartAllCompMgrErrGetDir() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetInstallRootDir, "", test.ErrTest)
	defer p1.Reset()

	err := componentMgr.StartAll()
	convey.So(err, convey.ShouldResemble, errors.New("make sure certs exist failed"))

	err = componentMgr.RestartAll()
	convey.So(err, convey.ShouldResemble, errors.New("make sure certs exist failed"))
}

func testStartAllCompMgrErrCpFile() {
	var p1 = gomonkey.ApplyFuncReturn(util.CopyServiceFileToSystemd, test.ErrTest)
	defer p1.Reset()

	err := componentMgr.StartAll()
	convey.So(err, convey.ShouldResemble, errors.New("register all service failed"))

	err = componentMgr.RestartAll()
	convey.So(err, convey.ShouldResemble, errors.New("register all service failed"))
}

func testStartAllCompMgrErrStart() {
	var p1 = gomonkey.ApplyFuncReturn(util.StartService, test.ErrTest)
	defer p1.Reset()
	err := componentMgr.StartAll()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("start target [%s] failed", constants.MefEdgeTargetFile))

	var p2 = gomonkey.ApplyFuncReturn(util.RestartService, test.ErrTest)
	defer p2.Reset()
	err = componentMgr.RestartAll()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("restart target [%s] failed", constants.MefEdgeTargetFile))
}

func TestStopAllComponentMgr(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(util.IsServiceInSystemd, true).
		ApplyFuncReturn(util.StopService, nil).
		ApplyFuncReturn(util.IsServiceActive, true).
		ApplyFuncReturn(filepath.EvalSymlinks, "", nil).
		ApplyFuncReturn(envutils.RunCommand, "", nil)
	defer p.Reset()

	convey.Convey("test stop all comp mgr should be success", t, testStopAllComponentMgr)
	convey.Convey("test stop all comp mgr should be success, svc is not in systemd", t, testStopAllCompMgrIsSvcSystemd)
	convey.Convey("test stop all comp mgr should be failed, stop svc error", t, testStopAllComponentMgrErrStopSvc)
	convey.Convey("test stop all comp mgr should be success, svc is active", t, testStopAllComponentMgrIsSvcActive)

	convey.Convey("test rm limit port rule should be failed, eval symlink error", t, testRmLimitPortRuleErrEvalSymlink)
	convey.Convey("test rm limit port rule should be success, run command error", t, testRmLimitPortRuleErrRunCommand)
}

func testStopAllComponentMgr() {
	err := componentMgr.StopAll()
	convey.So(err, convey.ShouldBeNil)
}

func testStopAllCompMgrIsSvcSystemd() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceInSystemd, false)
	defer p1.Reset()

	err := componentMgr.StopAll()
	convey.So(err, convey.ShouldBeNil)
}

func testStopAllComponentMgrErrStopSvc() {
	var p1 = gomonkey.ApplyFuncReturn(util.StopService, test.ErrTest)
	defer p1.Reset()

	err := componentMgr.StopAll()
	convey.So(err, convey.ShouldBeNil)
}

func testStopAllComponentMgrIsSvcActive() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceActive, false)
	defer p1.Reset()

	err := componentMgr.StopAll()
	convey.So(err, convey.ShouldBeNil)
}

func testRmLimitPortRuleErrEvalSymlink() {
	var p1 = gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "", test.ErrTest)
	defer p1.Reset()

	err := componentMgr.StopAll()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("delete port limit rule failed when stop components,"+
		" %v", test.ErrTest))
}

func testRmLimitPortRuleErrRunCommand() {
	var p1 = gomonkey.ApplyFuncReturn(envutils.RunCommand, "", test.ErrTest)
	defer p1.Reset()

	err := componentMgr.StopAll()
	convey.So(err, convey.ShouldBeNil)
}

func TestUnregisterAllServices(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(util.ResetFailedService, nil).
		ApplyFuncReturn(util.IsServiceInSystemd, true).
		ApplyFuncReturn(util.IsServiceActive, false).
		ApplyFuncReturn(util.IsServiceEnabled, true).
		ApplyFuncReturn(util.DisableService, nil).
		ApplyFuncReturn(util.RemoveServiceFileInSystemd, nil)
	defer p.Reset()

	convey.Convey("test unregister all svc should be success", t, testUnregisterAllSvc)
	convey.Convey("test unregister all svc should be failed, rst failed svc error", t, testUnregisterAllSvcErrRstFailedSvc)
	convey.Convey("test unregister all svc should be success, svc is not in systemd", t, testUnregisterAllSvcIsInSystemd)
	convey.Convey("test unregister all svc should be failed, svc is active", t, testUnregisterAllSvcErrActive)
	convey.Convey("test unregister all svc should be failed, disable svc error", t, testUnregisterAllSvcErrDisable)
	convey.Convey("test unregister all svc should be failed, remove svc file error", t, testUnregisterAllSvcErrRemoveFile)

	convey.Convey("test unregister target should be failed, svc is active", t, testUnregisterTargetErrActive)
	convey.Convey("test unregister target should be failed, remove svc file error", t, testUnregisterTargetErrRemoveFile)
}

func testUnregisterAllSvc() {
	err := componentMgr.UnregisterAllServices()
	convey.So(err, convey.ShouldBeNil)
}

func testUnregisterAllSvcErrRstFailedSvc() {
	var p1 = gomonkey.ApplyFuncReturn(util.ResetFailedService, test.ErrTest)
	defer p1.Reset()

	err := componentMgr.UnregisterAllServices()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("system service reset-failed failed, error: %v", test.ErrTest))
}

func testUnregisterAllSvcIsInSystemd() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceInSystemd, false)
	defer p1.Reset()

	err := componentMgr.UnregisterAllServices()
	convey.So(err, convey.ShouldBeNil)
}

func testUnregisterAllSvcErrActive() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceActive, true)
	defer p1.Reset()

	svcFileName := constants.EdgeInitServiceFile
	err := componentMgr.UnregisterAllServices()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("unregister service [%s] failed, error: "+
		"please stop service [%s] first", svcFileName, svcFileName))
}

func testUnregisterAllSvcErrDisable() {
	var p1 = gomonkey.ApplyFuncReturn(util.DisableService, test.ErrTest)
	defer p1.Reset()

	svcFileName := constants.EdgeInitServiceFile
	err := componentMgr.UnregisterAllServices()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("unregister service [%s] failed, error: "+
		"disable service [%s] failed", svcFileName, svcFileName))
}

func testUnregisterAllSvcErrRemoveFile() {
	var p1 = gomonkey.ApplyFuncReturn(util.RemoveServiceFileInSystemd, test.ErrTest)
	defer p1.Reset()

	svcFileName := constants.EdgeInitServiceFile
	err := componentMgr.UnregisterAllServices()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("unregister service [%s] failed, error: "+
		"remove service [%s] failed", svcFileName, svcFileName))
}

func testUnregisterTargetErrActive() {
	var p1 = gomonkey.ApplyFuncReturn(util.IsServiceActive, true).
		ApplyFuncReturn(util.StopService, test.ErrTest)
	defer p1.Reset()

	err := unregisterTarget()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("stop target [%s] failed", constants.MefEdgeTargetFile))
}

func testUnregisterTargetErrRemoveFile() {
	var p1 = gomonkey.ApplyFuncReturn(util.RemoveServiceFileInSystemd, test.ErrTest)
	defer p1.Reset()

	err := unregisterTarget()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("remove target [%s] failed", constants.MefEdgeTargetFile))
}

func TestUpdateServiceFiles(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(fileutils.EvalSymlinks, "", nil).
		ApplyFuncReturn(util.ReplaceValueInService, nil)
	defer p.Reset()

	convey.Convey("test update service files should be success", t, updateServiceFilesSuccess)
	convey.Convey("test update service files should be failed", t, updateServiceFilesFailed)
}

func updateServiceFilesSuccess() {
	err := componentMgr.UpdateServiceFiles(logDir, logDir)
	convey.So(err, convey.ShouldBeNil)
}

func updateServiceFilesFailed() {
	convey.Convey("get software abs dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.EvalSymlinks, "", test.ErrTest)
		defer p1.Reset()
		workDir := componentMgr.workPathMgr.GetWorkDir()
		err := componentMgr.UpdateServiceFiles(logDir, logDir)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("get work [%s] abs dir failed, error: %v",
			workDir, test.ErrTest))
	})

	convey.Convey("replace value in service failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.ReplaceValueInService, test.ErrTest)
		defer p1.Reset()
		servicePath := componentMgr.GetEdgeInit().Service.Path
		err := componentMgr.UpdateServiceFiles(logDir, logDir)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("replace marks in service file [%s] failed", servicePath))
	})
}
