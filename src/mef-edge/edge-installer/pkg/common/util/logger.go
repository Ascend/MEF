// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util this file for init logger
package util

import (
	"context"
	"fmt"
	"path/filepath"
	"syscall"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

const (
	maxLineLength     int = 1024
	opLogMaxBackups       = 10
	opLogMaxSize          = 4
	runLogMaxBackups      = 30
	runLogMaxSize         = 12
	defaultMaxSaveAge     = 30
)

// NewLogConfig create LogConfig instance
func NewLogConfig(LogFileName, logBackupDir string) *hwlog.LogConfig {
	return &hwlog.LogConfig{
		LogFileName:   LogFileName,
		BackupDirName: logBackupDir,
		IsCompress:    true,
		OnlyToFile:    true,
		MaxBackups:    hwlog.DefaultMaxBackups,
		MaxAge:        defaultMaxSaveAge,
		MaxLineLength: maxLineLength,
		EscapeHtml:    true,
	}
}

// MakeSureLogDir make sure log directories exist, if directories not exist, then create them
func MakeSureLogDir(compLogDir string) error {
	mask := syscall.Umask(constants.ModeUmask022)
	defer syscall.Umask(mask)

	edgeLogDir := filepath.Dir(compLogDir)
	if err := fileutils.CreateDir(edgeLogDir, constants.Mode755); err != nil {
		return fmt.Errorf("create edge log dir [%s] failed, error: %v", edgeLogDir, err)
	}
	if _, err := fileutils.RealDirCheck(edgeLogDir, true, false); err != nil {
		return fmt.Errorf("check edge log dir [%s] failed, error: %v", edgeLogDir, err)
	}
	if err := fileutils.CreateDir(compLogDir, constants.Mode750); err != nil {
		return fmt.Errorf("create component log dir [%s] failed, error: %v", compLogDir, err)
	}
	return nil
}

// InitComponentLog initialize logger for specific component
func InitComponentLog(component string) error {
	compLogDir, compLogBackupDir, err := path.GetCompLogDirs(component)
	if err != nil {
		return err
	}
	return InitLog(compLogDir, compLogBackupDir)
}

// InitLog initialize logger
func InitLog(compLogDir, compLogBackupDir string) error {
	if err := MakeSureLogDir(compLogDir); err != nil {
		return err
	}
	if err := MakeSureLogDir(compLogBackupDir); err != nil {
		return err
	}
	runLogConf := NewLogConfig(filepath.Join(compLogDir, fmt.Sprintf("%s_%s", filepath.Base(compLogDir),
		constants.RunLogFile)), compLogBackupDir)
	runLogConf.MaxBackups = runLogMaxBackups
	runLogConf.FileMaxSize = runLogMaxSize
	runLogConf.DisableRotationIfSwitchUser = true
	opLogConf := NewLogConfig(filepath.Join(compLogDir, fmt.Sprintf("%s_%s", filepath.Base(compLogDir),
		constants.OperateLogFile)), compLogBackupDir)
	opLogConf.MaxBackups = opLogMaxBackups
	opLogConf.FileMaxSize = opLogMaxSize

	return InitHwLogger(runLogConf, opLogConf)
}

// InitHwLogger initialize run and operate logger
func InitHwLogger(runLogConfig, opLogConfig *hwlog.LogConfig) error {
	err := hwlog.InitRunLogger(runLogConfig, context.Background())
	if err != nil {
		return err
	}
	err = hwlog.InitOperateLogger(opLogConfig, context.Background())
	if err != nil {
		return err
	}
	return nil
}
