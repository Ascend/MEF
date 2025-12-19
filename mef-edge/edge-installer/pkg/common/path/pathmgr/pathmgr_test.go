// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package pathmgr for path manager
package pathmgr

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

func TestNewPathMgr(t *testing.T) {
	convey.Convey("test PathMgr", t, func() {
		pathMgr = NewPathMgr(testInstallRootDir, testInstallationPkgDir, testLogRootDir, testLogBackupRootDir)
		convey.Convey("test SoftwarePathMgr method", testSoftwarePathMgr)
		convey.Convey("test WorkPathMgr method", testWorkPathMgr)
		convey.Convey("test ConfigPathMgr method", testConfigPathMgr)
		convey.Convey("test LogPathMgr method", testLogPathMgr)

		convey.Convey("test WorkAbsPathMgr method", testWorkAbsPathMgr)
	})
}

func testSoftwarePathMgr() {
	if pathMgr == nil {
		panic("path manager is nil")
	}
	softwarePathMgr := pathMgr.SoftwarePathMgr
	softwarePathMgr.GetInstallRootDir()
	softwarePathMgr.GetInstallationPkgDir()
	softwarePathMgr.GetPkgVersionXmlPath()
	softwarePathMgr.GetPkgCompSoftwareDir(constants.EdgeInstaller)
	softwarePathMgr.GetPkgLibDir()
	softwarePathMgr.GetPkgRunShPath()
	softwarePathMgr.GetPkgCompConfigDir(constants.EdgeInstaller)
}

func testWorkPathMgr() {
	if pathMgr == nil {
		panic("path manager is nil")
	}
	workPathMgr := pathMgr.WorkPathMgr
	workPathMgr.GetInstallRootDir()
	workPathMgr.GetMefEdgeDir()
	workPathMgr.GetWorkDir()
	workPathMgr.GetWorkADir()
	workPathMgr.GetWorkBDir()
	workPathMgr.GetCompWorkDir(constants.EdgeMain)
	workPathMgr.GetCompBinaryPath(constants.EdgeMain, constants.EdgeMainFileName)
	workPathMgr.GetCompBinaryPath(constants.EdgeInstaller, constants.MefInitScriptName)
	workPathMgr.GetCompJsonPath(constants.EdgeCore, constants.EdgeCoreJsonName)
	workPathMgr.GetVersionXmlPath()
	workPathMgr.GetDockerIsolationShPath()
	workPathMgr.GetDockerRestoreShPath()
	workPathMgr.GetServicePath(constants.DockerServiceFile)
	workPathMgr.GetUpgradeTempDir()
	workPathMgr.GetUpgradeTempVersionXmlPath()
	workPathMgr.GetUpgradeTempBinaryPath()
	workPathMgr.GetCompLogLinkDir(constants.EdgeCore)
	workPathMgr.GetCompLogLinkPath(constants.EdgeCore, constants.EdgeCoreLogFile)
	workPathMgr.GetCompLogBackupLinkDir(constants.EdgeCore)

	var p1 = gomonkey.ApplyFuncReturn(fileutils.IsExist, true)
	defer p1.Reset()
	workAbsDir, err := workPathMgr.GetWorkAbsDir()
	convey.So(workAbsDir, convey.ShouldResemble, pathMgr.WorkPathMgr.GetUpgradeTempDir())
	convey.So(err, convey.ShouldBeNil)

	var p2 = gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncReturn(GetTargetInstallDir, testInstallRootDir, nil)
	defer p2.Reset()
	workAbsDir, err = workPathMgr.GetWorkAbsDir()
	convey.So(workAbsDir, convey.ShouldResemble, testInstallRootDir)
	convey.So(err, convey.ShouldBeNil)

	var p3 = gomonkey.ApplyFuncReturn(GetTargetInstallDir, "", test.ErrTest)
	defer p3.Reset()
	workAbsDir, err = workPathMgr.GetWorkAbsDir()
	convey.So(workAbsDir, convey.ShouldResemble, "")
	expErr := errors.New("get target software install dir failed")
	convey.So(err, convey.ShouldResemble, expErr)
}

func testConfigPathMgr() {
	if pathMgr == nil {
		panic("path manager is nil")
	}
	configPathMgr := pathMgr.ConfigPathMgr
	configPathMgr.GetInstallRootDir()
	configPathMgr.GetMefEdgeDir()
	configPathMgr.GetConfigDir()
	configPathMgr.GetTempCertsDir()
	configPathMgr.GetConfigBackupDir()
	configPathMgr.GetConfigBackupTempDir()
	configPathMgr.GetCompConfigDir(constants.EdgeMain)
	configPathMgr.GetCompKmcDir(constants.EdgeMain)
	configPathMgr.GetCompKmcConfigPath(constants.EdgeMain)
	configPathMgr.GetCompInnerCertsDir(constants.EdgeMain)
	configPathMgr.GetCompInnerRootCertPath(constants.EdgeMain)
	configPathMgr.GetCompInnerSvrCertPath(constants.EdgeMain)
	configPathMgr.GetCompInnerSvrCertPath(constants.EdgeOm)
	configPathMgr.GetCompInnerSvrKeyPath(constants.EdgeMain)
	configPathMgr.GetCompInnerSvrKeyPath(constants.EdgeOm)
	configPathMgr.GetTempRootCertDir()
	configPathMgr.GetTempRootCertPath()
	configPathMgr.GetTempRootCerKeyPath()
	configPathMgr.GetSnPath()
	configPathMgr.GetDockerBackupPath()
	configPathMgr.GetEdgeMainDbPath()
	configPathMgr.GetOMCertDir()
	configPathMgr.GetOMRootCertPath()
	configPathMgr.GetContainerConfigPath()
	configPathMgr.GetPodConfigPath()
	configPathMgr.GetImageCertPath()
	configPathMgr.GetEdgeOmDbPath()
	configPathMgr.GetEdgeCoreDbPath()
	configPathMgr.GetEdgeCoreConfigPath()
}

func testLogPathMgr() {
	if pathMgr == nil {
		panic("path manager is nil")
	}
	logPathMgr := pathMgr.LogPathMgr
	logPathMgr.GetLogRootDir()
	logPathMgr.GetEdgeLogDir()
	logPathMgr.GetComponentLogDir(constants.EdgeCore)
	logPathMgr.GetComponentLogPath(constants.EdgeCore, constants.EdgeCoreLogFile)
	logPathMgr.GetLogBackupRootDir()
	logPathMgr.GetEdgeLogBackupDir()
	logPathMgr.GetComponentLogBackupDir(constants.EdgeCore)
	logPathMgr.GetComponentLogBackupPath(constants.EdgeCore, constants.EdgeCoreLogFile)
}

func testWorkAbsPathMgr() {
	workAbsPathMgr := NewWorkAbsPathMgr(testInstallRootDir)
	workAbsPathMgr.GetCompWorkDir(constants.EdgeMain)
	workAbsPathMgr.GetCompVarDir(constants.EdgeMain)
	workAbsPathMgr.GetCompConfigDir(constants.EdgeMain)
	workAbsPathMgr.GetCompBinDir(constants.EdgeMain)
	workAbsPathMgr.GetCompBinFilePath(constants.EdgeMain, constants.EdgeMainFileName)
	workAbsPathMgr.GetCompScriptDir(constants.EdgeMain)
	workAbsPathMgr.GetCompScriptFilePath(constants.EdgeInstaller, constants.DockerIsolate)
	workAbsPathMgr.GetVersionXmlPath()
	workAbsPathMgr.GetLibDir()
	workAbsPathMgr.GetRunShPath()
	workAbsPathMgr.GetInstallBinaryPath()
	workAbsPathMgr.GetUpgradeBinaryPath()
	workAbsPathMgr.GetUpgradeShPath()
	workAbsPathMgr.GetResetInstallShPath()
	workAbsPathMgr.GetServicePath(constants.ResetService)
}
