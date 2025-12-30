// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package path utils related to path
package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
)

// GetInstallRootDir get install root dir. After installation, default: /usr/local/mindx
func GetInstallRootDir() (string, error) {
	installDir, err := GetInstallDir()
	if err != nil {
		return "", fmt.Errorf("get install dir failed, %v", err)
	}
	return filepath.Dir(installDir), nil
}

// GetInstallDir get install dir, always used during installation. During installation, e.g. /tmp/xxx
func GetInstallDir() (string, error) {
	// currentPath: e.g. /usr/local/mindx/MEFEdge/software_A/edge_main/bin/edge-main
	// during installation, e.g. /tmp/xxx/software/edge_installer/bin/install
	currentPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("get current path failed, %v", err)
	}
	realInstallDir, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(currentPath)))))
	if err != nil {
		return "", fmt.Errorf("eval install dir symlink failed, %v", err)
	}
	return realInstallDir, nil
}

// GetWorkPathMgr get work path mgr
func GetWorkPathMgr() (*pathmgr.WorkPathMgr, error) {
	// installRootDir: default: /usr/local/mindx
	installRootDir, err := GetInstallRootDir()
	if err != nil {
		hwlog.RunLog.Errorf("get install root dir failed, error: %v", err)
		return nil, err
	}
	return pathmgr.NewWorkPathMgr(installRootDir), nil
}

// GetConfigPathMgr get config path mgr
func GetConfigPathMgr() (*pathmgr.ConfigPathMgr, error) {
	// installRootDir: default: /usr/local/mindx
	installRootDir, err := GetInstallRootDir()
	if err != nil {
		hwlog.RunLog.Errorf("get install root dir failed, error: %v", err)
		return nil, err
	}
	return pathmgr.NewConfigPathMgr(installRootDir), nil
}

// GetLogPathMgr get log path mgr
func GetLogPathMgr() (*pathmgr.LogPathMgr, error) {
	// installRootDir: default: /usr/local/mindx
	installRootDir, err := GetInstallRootDir()
	if err != nil {
		hwlog.RunLog.Errorf("get install root dir failed, error: %v", err)
		return nil, err
	}

	logRootDir, err := GetLogRootDir(installRootDir)
	if err != nil {
		hwlog.RunLog.Errorf("get log root dir failed, error: %v", err)
		return nil, errors.New("get log root dir failed")
	}
	logBackupRootDir, err := GetLogBackupRootDir(installRootDir)
	if err != nil {
		hwlog.RunLog.Errorf("get log backup root dir failed, error: %v", err)
		return nil, errors.New("get log backup root dir failed")
	}

	return pathmgr.NewLogPathMgr(logRootDir, logBackupRootDir), nil
}

// GetEdgeLogDirs get edgeLogDir and edgeLogBackupDir
func GetEdgeLogDirs() (string, string, error) {
	logPathMgr, err := GetLogPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get log path manager failed, error: %v", err)
		return "", "", errors.New("get log path manager failed")
	}

	edgeLogDir := logPathMgr.GetEdgeLogDir()
	if err = fileutils.IsSoftLink(edgeLogDir); err != nil {
		return "", "", fmt.Errorf("edge log dir link check failed, %v", err)
	}
	edgeLogBackupDir := logPathMgr.GetEdgeLogBackupDir()
	if err = fileutils.IsSoftLink(edgeLogBackupDir); err != nil {
		return "", "", fmt.Errorf("edge log backup dir link check failed, %v", err)
	}
	return edgeLogDir, edgeLogBackupDir, nil
}

// GetLogRootDir get log root dir. default: /var/alog
func GetLogRootDir(installRootDir string) (string, error) {
	workPathMgr := pathmgr.NewWorkPathMgr(installRootDir)

	// /usr/local/mindx/MEFEdge/software/edge_installer/var/log -> /var/alog/MEFEdge_log/edge_installer
	resoledPath, err := filepath.EvalSymlinks(workPathMgr.GetCompLogLinkDir(constants.EdgeInstaller))
	if err != nil {
		return "", fmt.Errorf("eval log link dir symlink failed, %v", err)
	}
	index := strings.LastIndex(resoledPath, constants.MEFEdgeLogName)
	if index < 0 {
		return "", fmt.Errorf("[%s] not found in path [%s]", constants.MEFEdgeLogName, resoledPath)
	}
	return resoledPath[:index], nil
}

// GetLogBackupRootDir get log backup root dir. default: /home/log
func GetLogBackupRootDir(installRootDir string) (string, error) {
	workPathMgr := pathmgr.NewWorkPathMgr(installRootDir)

	// /usr/local/mindx/MEFEdge/software/edge_installer/var/log_backup -> /home/log/MEFEdge_logbackup/edge_installer
	resoledPath, err := filepath.EvalSymlinks(workPathMgr.GetCompLogBackupLinkDir(constants.EdgeInstaller))
	if err != nil {
		return "", fmt.Errorf("eval log backup link dir symlink failed, %v", err)
	}
	index := strings.LastIndex(resoledPath, constants.MEFEdgeLogBackupName)
	if index < 0 {
		return "", fmt.Errorf("[%s] not found in path [%s]", constants.MEFEdgeLogBackupName, resoledPath)
	}
	return resoledPath[:index], nil
}

// GetCompLogDirs get component log and log backup dir.
// e.g. /var/alog/MEFEdge_log/edge_installer, /home/log/MEFEdge_logbackup/edge_installer
func GetCompLogDirs(component string) (string, string, error) {
	workPathMgr, err := GetWorkPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get work path manager failed, error: %v", err)
		return "", "", errors.New("get work path manager failed")
	}

	// e.g. /usr/local/mindx/MEFEdge/software/edge_installer/var/log -> /var/alog/MEFEdge_log/edge_installer
	compLogDir, err := fileutils.ReadLink(workPathMgr.GetCompLogLinkDir(component))
	if err != nil {
		return "", "", fmt.Errorf("read compnonent log link failed, %v", err)
	}
	// e.g /usr/local/mindx/MEFEdge/software/edge_installer/var/log_backup -> /home/log/MEFEdge_logbackup/edge_installer
	compLogBackupDir, err := fileutils.ReadLink(workPathMgr.GetCompLogBackupLinkDir(component))
	if err != nil {
		return "", "", fmt.Errorf("read compnonent log backup link failed, %v", err)
	}
	return compLogDir, compLogBackupDir, nil
}
