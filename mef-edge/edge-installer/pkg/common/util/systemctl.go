// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package util this file for exec systemctl command
package util

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

// IsServiceInSystemd is service file in systemd service directory
func IsServiceInSystemd(name string) bool {
	return fileutils.IsExist(filepath.Join(constants.SystemdServiceDir, name))
}

// ReloadServiceDaemon reload systemd service config
func ReloadServiceDaemon() error {
	if _, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlReload); err != nil {
		return err
	}
	return nil
}

// ResetFailedService systemd service reset failed
func ResetFailedService() error {
	if _, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlResetFailed); err != nil {
		return err
	}
	return nil
}

// StartService start systemd service
func StartService(name string) error {
	if _, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlStart, name); err != nil {
		return err
	}
	return nil
}

// StopService stop systemd service
func StopService(name string) error {
	if _, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlStop, name); err != nil {
		return err
	}
	return nil
}

// RestartService restart systemd service
func RestartService(name string) error {
	if _, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlRestart, name); err != nil {
		return err
	}
	return nil
}

// EnableService enable systemd service
func EnableService(name string) error {
	if _, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlEnable, name); err != nil {
		return err
	}
	return nil
}

// IsServiceEnabled is systemd service enabled
func IsServiceEnabled(name string) bool {
	output, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlIsEnabled, name)
	if err != nil {
		return false
	}
	return output == constants.SystemctlEnabled
}

// DisableService disable systemd service
func DisableService(name string) error {
	if _, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlDisable, name); err != nil {
		return err
	}
	return nil
}

// IsServiceActive is systemd service active
func IsServiceActive(name string) bool {
	output, err := envutils.RunCommand(constants.Systemctl, envutils.DefCmdTimeoutSec,
		constants.SystemctlIsActive, name)
	if err != nil {
		return false
	}
	return output == constants.SystemctlStatusActive
}

// CopyServiceFileToSystemd copy .service file to systemd service path
func CopyServiceFileToSystemd(servicePath string, mode uint32, userName string) error {
	if !fileutils.IsExist(servicePath) {
		return fmt.Errorf("service file[%s] not exist", servicePath)
	}
	uid, err := envutils.GetUid(userName)
	if err != nil {
		return fmt.Errorf("get user id faild, error: %v", err)
	}
	srcPath, err := fileutils.CheckOwnerAndPermission(servicePath, os.FileMode(mode), uid)
	if err != nil {
		return fmt.Errorf("check service file owner and permission failed, error: %v", err)
	}
	_, name := filepath.Split(srcPath)
	absSrvPath, err := filepath.EvalSymlinks(constants.SystemdServiceDir)
	if err != nil {
		return fmt.Errorf("get abs srv path failed: %s", err.Error())
	}
	dstPath := filepath.Join(absSrvPath, name)
	if err = fileutils.CopyFile(srcPath, dstPath); err != nil {
		return fmt.Errorf("copy service file to systemd failed, error: %v", err)
	}
	if err = ReloadServiceDaemon(); err != nil {
		hwlog.RunLog.Warnf("reload service daemon, warning: %v", err)
	}
	hwlog.RunLog.Infof("copy service file [%s] to systemd success", servicePath)
	return nil
}

// RemoveServiceFileInSystemd remove .service file in systemd service path
func RemoveServiceFileInSystemd(name string) error {
	absServiceDir, err := filepath.EvalSymlinks(constants.SystemdServiceDir)
	if err != nil {
		return fmt.Errorf("get abs service dir failed: %s", err.Error())
	}
	serviceFile := filepath.Join(absServiceDir, name)
	if !fileutils.IsExist(serviceFile) {
		return nil
	}
	if err := fileutils.DeleteFile(serviceFile); err != nil {
		return fmt.Errorf("remove service file [%s] failed, error: %v", serviceFile, err)
	}
	if err := ReloadServiceDaemon(); err != nil {
		return fmt.Errorf("reload service daemon failed, error: %v", err)
	}
	hwlog.RunLog.Infof("remove service file [%s] in systemd success", serviceFile)
	return nil
}

// ReplaceValueInService replace value in systemd service file
func ReplaceValueInService(servicePath string, mode uint32, userName string, replaceDic map[string]string) error {
	uid, err := envutils.GetUid(userName)
	if err != nil {
		return fmt.Errorf("get user id faild, error: %v", err)
	}
	validPath, err := fileutils.CheckOwnerAndPermission(servicePath, os.FileMode(mode), uid)
	if err != nil {
		return fmt.Errorf("systemd service file[%s] is invalid, error: %v", servicePath, err)
	}
	data, err := fileutils.LoadFile(validPath)
	if err != nil {
		return fmt.Errorf("load systemd service file[%s] failed, error: %v", validPath, err)
	}
	for mark, value := range replaceDic {
		regPattern := fmt.Sprintf("\\{%s\\}", mark)
		reg := regexp.MustCompile(regPattern)
		data = reg.ReplaceAll(data, []byte(value))
	}
	if err = fileutils.WriteData(validPath, data); err != nil {
		return fmt.Errorf("save systemd service file[%s] failed, error: %v", validPath, err)
	}
	return nil
}

// GetExecStartInService get value in systemd service file
func GetExecStartInService(servicePath string, mode, uid uint32) (string, error) {
	validPath, err := fileutils.CheckOwnerAndPermission(servicePath, os.FileMode(mode), uid)
	if err != nil {
		return "", fmt.Errorf("systemd service file[%s] is invalid, error: %v", servicePath, err)
	}
	data, err := fileutils.LoadFile(validPath)
	if err != nil {
		return "", fmt.Errorf("load systemd service file[%s] failed, error: %v", validPath, err)
	}
	reg := regexp.MustCompile(constants.ExecStartPattern)
	matches := reg.FindSubmatch(data)
	if len(matches) <= 1 {
		return "", fmt.Errorf("'ExecStart' not found in service file[%s]", servicePath)
	}
	return string(matches[1]), nil
}
