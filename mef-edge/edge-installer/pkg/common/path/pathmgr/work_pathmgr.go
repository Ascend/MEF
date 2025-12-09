// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package pathmgr work path manager
package pathmgr

import (
	"errors"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

// WorkPathMgr work path manager for installing the software in softlink
type WorkPathMgr struct {
	installRootDir string
}

// NewWorkPathMgr new work path manager in softlink
func NewWorkPathMgr(installRootDir string) *WorkPathMgr {
	return &WorkPathMgr{
		installRootDir: installRootDir,
	}
}

// GetInstallRootDir get install root dir. default: /usr/local/mindx
func (wpm *WorkPathMgr) GetInstallRootDir() string {
	return wpm.installRootDir
}

// GetMefEdgeDir get mef edge dir. default: /usr/local/mindx/MEFEdge
func (wpm *WorkPathMgr) GetMefEdgeDir() string {
	return filepath.Join(wpm.GetInstallRootDir(), constants.MEFEdgeName)
}

// GetWorkDir get work dir path in softlink. default: /usr/local/mindx/MEFEdge/software
func (wpm *WorkPathMgr) GetWorkDir() string {
	return filepath.Join(wpm.GetMefEdgeDir(), constants.SoftwareDir)
}

// GetWorkADir get workA dir path in softlink. default: /usr/local/mindx/MEFEdge/software_A
func (wpm *WorkPathMgr) GetWorkADir() string {
	return filepath.Join(wpm.GetMefEdgeDir(), constants.SoftwareDirA)
}

// GetWorkBDir get workB dir path in softlink. default: /usr/local/mindx/MEFEdge/software_B
func (wpm *WorkPathMgr) GetWorkBDir() string {
	return filepath.Join(wpm.GetMefEdgeDir(), constants.SoftwareDirB)
}

// GetCompWorkDir get component work dir in softlink. e.g. /usr/local/mindx/MEFEdge/software/edge_main
func (wpm *WorkPathMgr) GetCompWorkDir(component string) string {
	return filepath.Join(wpm.GetWorkDir(), component)
}

// GetCompBinaryPath get component binary path in softlink.
// e.g. /usr/local/mindx/MEFEdge/software/edge_main/bin/edge-main
func (wpm *WorkPathMgr) GetCompBinaryPath(component, binaryName string) string {
	if component == constants.EdgeInstaller {
		return filepath.Join(wpm.GetCompWorkDir(component), constants.Script, binaryName)
	}
	return filepath.Join(wpm.GetCompWorkDir(component), constants.Bin, binaryName)
}

// GetCompJsonPath get component json file path in softlink.
// e.g. /usr/local/mindx/MEFEdge/software/edge_core/config/edgecore.json
func (wpm *WorkPathMgr) GetCompJsonPath(component, jsonFileName string) string {
	return filepath.Join(wpm.GetCompWorkDir(component), constants.Config, jsonFileName)
}

// GetVersionXmlPath get version.xml path in softlink. default: /usr/local/mindx/MEFEdge/software/version.xml
func (wpm *WorkPathMgr) GetVersionXmlPath() string {
	return filepath.Join(wpm.GetWorkDir(), constants.VersionXml)
}

// GetDockerIsolationShPath get mef_docker_isolation.sh path in softlink.
// default: /usr/local/mindx/MEFEdge/software/edge_installer/script/docker_isolate/mef_docker_isolation.sh
func (wpm *WorkPathMgr) GetDockerIsolationShPath() string {
	return filepath.Join(wpm.GetCompWorkDir(constants.EdgeInstaller), constants.Script,
		constants.DockerIsolate, constants.DockerIsolationScript)
}

// GetDockerRestoreShPath get mef_docker_restore.sh path in softlink.
// default: /usr/local/mindx/MEFEdge/software/edge_installer/script/docker_isolate/mef_docker_restore.sh
func (wpm *WorkPathMgr) GetDockerRestoreShPath() string {
	return filepath.Join(wpm.GetCompWorkDir(constants.EdgeInstaller), constants.Script,
		constants.DockerIsolate, constants.DockerRestoreScript)
}

// GetServicePath get service path by name in softlink.
// e.g. /usr/local/mindx/MEFEdge/software/edge_installer/service/docker.service
func (wpm *WorkPathMgr) GetServicePath(serviceName string) string {
	return filepath.Join(wpm.GetCompWorkDir(constants.EdgeInstaller), constants.ServiceDir, serviceName)
}

// GetUpgradeTempDir get upgrade temp dir. default: /usr/local/mindx/MEFEdge/software_temp
func (wpm *WorkPathMgr) GetUpgradeTempDir() string {
	return filepath.Join(wpm.GetMefEdgeDir(), constants.SoftwareDirTemp)
}

// GetUpgradeTempVersionXmlPath get upgrade version.xml temp path.
// default: /usr/local/mindx/MEFEdge/software_temp/version.xml
func (wpm *WorkPathMgr) GetUpgradeTempVersionXmlPath() string {
	return filepath.Join(wpm.GetUpgradeTempDir(), constants.VersionXml)
}

// GetUpgradeTempBinaryPath get upgrade binary temp path.
// default: /usr/local/mindx/MEFEdge/software_temp/edge_installer/bin/upgrade
func (wpm *WorkPathMgr) GetUpgradeTempBinaryPath() string {
	return filepath.Join(wpm.GetUpgradeTempDir(), constants.EdgeInstaller, constants.Bin, constants.Upgrade)
}

// GetCompLogLinkDir get component log link dir. e.g. /usr/local/mindx/MEFEdge/software/edge_installer/var/log
func (wpm *WorkPathMgr) GetCompLogLinkDir(component string) string {
	return filepath.Join(wpm.GetCompWorkDir(component), constants.Var, constants.Log)
}

// GetCompLogLinkPath get component log file path in softlink.
// e.g. /usr/local/mindx/MEFEdge/software/edge_installer/var/log/edge_main_run.log
func (wpm *WorkPathMgr) GetCompLogLinkPath(component, fileName string) string {
	return filepath.Join(wpm.GetCompLogLinkDir(component), fileName)
}

// GetCompLogBackupLinkDir get component log backup link dir.
// e.g. /usr/local/mindx/MEFEdge/software/edge_installer/var/log_backup
func (wpm *WorkPathMgr) GetCompLogBackupLinkDir(component string) string {
	return filepath.Join(wpm.GetCompWorkDir(component), constants.Var, constants.LogBackup)
}

// GetWorkAbsDir get work absolute dir for installing the software. e.g. /usr/local/mindx/MEFEdge/software_A
func (wpm *WorkPathMgr) GetWorkAbsDir() (string, error) {
	upgradeDir := wpm.GetUpgradeTempDir()
	if fileutils.IsExist(upgradeDir) {
		return upgradeDir, nil
	}

	workAbsDir, err := GetTargetInstallDir(wpm.GetInstallRootDir())
	if err != nil {
		hwlog.RunLog.Errorf("get target software install dir failed, error: %v", err)
		return "", errors.New("get target software install dir failed")
	}
	return workAbsDir, nil
}

// WorkAbsPathMgr work absolute path manager for installing the software
type WorkAbsPathMgr struct {
	// e.g. /usr/local/mindx/MEFEdge/software_A
	workAbsDir string
}

// NewWorkAbsPathMgr new work absolute path manager
func NewWorkAbsPathMgr(workAbsDir string) *WorkAbsPathMgr {
	return &WorkAbsPathMgr{
		workAbsDir: workAbsDir,
	}
}

// GetCompWorkDir get component work dir of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_main
func (wpm *WorkAbsPathMgr) GetCompWorkDir(component string) string {
	return filepath.Join(wpm.workAbsDir, component)
}

// GetCompVarDir get component var dir of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_main/var
func (wpm *WorkAbsPathMgr) GetCompVarDir(component string) string {
	return filepath.Join(wpm.GetCompWorkDir(component), constants.Var)
}

// GetCompConfigDir get component config dir of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_main/config
func (wpm *WorkAbsPathMgr) GetCompConfigDir(component string) string {
	return filepath.Join(wpm.GetCompWorkDir(component), constants.Config)
}

// GetCompBinDir get component bin dir of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_main/bin
func (wpm *WorkAbsPathMgr) GetCompBinDir(component string) string {
	return filepath.Join(wpm.GetCompWorkDir(component), constants.Bin)
}

// GetCompBinFilePath get file path in component bin dir of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_main/bin/edge-main
func (wpm *WorkAbsPathMgr) GetCompBinFilePath(component, fileName string) string {
	return filepath.Join(wpm.GetCompBinDir(component), fileName)
}

// GetCompScriptDir get component script dir of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_main/script
func (wpm *WorkAbsPathMgr) GetCompScriptDir(component string) string {
	return filepath.Join(wpm.GetCompWorkDir(component), constants.Script)
}

// GetCompScriptFilePath get file path in component script dir of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_core/script/prepare.sh
func (wpm *WorkAbsPathMgr) GetCompScriptFilePath(component, fileName string) string {
	return filepath.Join(wpm.GetCompScriptDir(component), fileName)
}

// GetVersionXmlPath get version.xml path of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/version.xml
func (wpm *WorkAbsPathMgr) GetVersionXmlPath() string {
	return filepath.Join(wpm.workAbsDir, constants.VersionXml)
}

// GetLibDir get lib dir of work absolute path manager. e.g. /usr/local/mindx/MEFEdge/software_A/lib
func (wpm *WorkAbsPathMgr) GetLibDir() string {
	return filepath.Join(wpm.workAbsDir, constants.Lib)
}

// GetRunShPath get run.sh path of work absolute path manager. e.g. /usr/local/mindx/MEFEdge/software_A/run.sh
func (wpm *WorkAbsPathMgr) GetRunShPath() string {
	return filepath.Join(wpm.workAbsDir, constants.RunScript)
}

// GetInstallBinaryPath get install binary path of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_installer/bin/install
func (wpm *WorkAbsPathMgr) GetInstallBinaryPath() string {
	return filepath.Join(wpm.GetCompWorkDir(constants.EdgeInstaller), constants.Bin, constants.Install)
}

// GetUpgradeBinaryPath get upgrade binary path of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_installer/bin/upgrade
func (wpm *WorkAbsPathMgr) GetUpgradeBinaryPath() string {
	return filepath.Join(wpm.GetCompWorkDir(constants.EdgeInstaller), constants.Bin, constants.Upgrade)
}

// GetUpgradeShPath get upgrade.sh path of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_installer/script/upgrade.sh
func (wpm *WorkAbsPathMgr) GetUpgradeShPath() string {
	return filepath.Join(wpm.GetCompWorkDir(constants.EdgeInstaller), constants.Script,
		constants.Upgrade+constants.ShellExt)
}

// GetResetInstallShPath get reset_install.sh path of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_installer/script/reset_install.sh
func (wpm *WorkAbsPathMgr) GetResetInstallShPath() string {
	return filepath.Join(wpm.GetCompWorkDir(constants.EdgeInstaller), constants.Script, constants.ResetInstallScript)
}

// GetServicePath get service path by name of work absolute path manager.
// e.g. /usr/local/mindx/MEFEdge/software_A/edge_installer/service/reset_mefedge.service
func (wpm *WorkAbsPathMgr) GetServicePath(serviceName string) string {
	return filepath.Join(wpm.GetCompWorkDir(constants.EdgeInstaller), constants.ServiceDir, serviceName)
}
