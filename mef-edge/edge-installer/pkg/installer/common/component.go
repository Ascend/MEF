// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common this file for get component information
package common

import (
	"errors"
	"fmt"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

// FileInfo file information
type FileInfo struct {
	Name      string
	Path      string
	ModeUmask uint32
	UserName  string
}

// Component edge component
type Component struct {
	Name    string
	Dir     string
	Service FileInfo
	Bin     FileInfo
}

func (c Component) isServiceActive() bool {
	serviceActive := false
	for i := 1; i <= constants.CheckServiceNum; i++ {
		time.Sleep(constants.CheckServiceWaitTime)
		if active := util.IsServiceActive(c.Service.Name); active {
			serviceActive = true
			break
		}
	}
	return serviceActive
}

// Start component
func (c Component) Start() error {
	if active := util.IsServiceActive(c.Service.Name); active {
		hwlog.RunLog.Infof("component [%s] is already active", c.Name)
		return nil
	}

	if err := c.RegisterService(); err != nil {
		hwlog.RunLog.Errorf("register service [%s] failed, error: %v", c.Service.Name, err)
		return err
	}
	if err := util.StartService(c.Service.Name); err != nil {
		hwlog.RunLog.Errorf("start service [%s] failed, error: %v", c.Service.Name, err)
		return err
	}

	if !c.isServiceActive() {
		hwlog.RunLog.Warnf("service [%s] is not active", c.Service.Name)
	}
	hwlog.RunLog.Infof("start service [%s] success", c.Service.Name)
	return nil
}

// Stop component
func (c Component) Stop() error {
	if !util.IsServiceInSystemd(c.Service.Name) {
		fmt.Printf("The service file [%s] does not exist. The running process may not be stopped. "+
			"Please check and manually stop it.\n", c.Service.Name)
		hwlog.RunLog.Warnf("service [%s] not in systemd", c.Service.Name)
		return nil
	}

	if err := util.StopService(c.Service.Name); err != nil {
		hwlog.RunLog.Errorf("stop service [%s] failed, error: %v", c.Service.Name, err)
		return errors.New("stop service failed")
	}

	if active := util.IsServiceActive(c.Service.Name); active {
		fmt.Printf("warning: service [%s] is still active.\n", c.Name)
		hwlog.RunLog.Warnf("service [%s] is still active", c.Name)
		return nil
	}
	hwlog.RunLog.Infof("stop service [%s] success", c.Service.Name)
	return nil
}

// Restart component
func (c Component) Restart() error {
	if err := c.RegisterService(); err != nil {
		hwlog.RunLog.Errorf("register service [%s] failed, error: %v", c.Service.Name, err)
		return fmt.Errorf("register service [%s] failed", c.Service.Name)
	}
	if err := util.RestartService(c.Service.Name); err != nil {
		hwlog.RunLog.Errorf("restart service [%s] failed, error: %v", c.Service.Name, err)
		return fmt.Errorf("restart service [%s] failed", c.Service.Name)
	}

	if !c.isServiceActive() {
		hwlog.RunLog.Warnf("service [%s] is not active", c.Service.Name)
	}
	hwlog.RunLog.Infof("restart service [%s] success", c.Service.Name)
	return nil
}

// RegisterService register service
func (c Component) RegisterService() error {
	if err := util.CopyServiceFileToSystemd(c.Service.Path, c.Service.ModeUmask, c.Service.UserName); err != nil {
		hwlog.RunLog.Errorf("copy service file [%s] failed, error: %v", c.Service.Path, err)
		return fmt.Errorf("copy service file [%s] failed", c.Service.Path)
	}
	if err := util.EnableService(c.Service.Name); err != nil {
		hwlog.RunLog.Warnf("enable service [%s], warning: %v", c.Service.Name, err)
	}
	hwlog.RunLog.Infof("register service [%s] success", c.Service.Name)
	return nil
}

// UnregisterService unregister service
func (c Component) UnregisterService() error {
	if !util.IsServiceInSystemd(c.Service.Name) {
		fmt.Printf("The service file [%s] does not exist. The service may not be unregistered. "+
			"Please check and manually unregister it.\n", c.Service.Name)
		return nil
	}

	if active := util.IsServiceActive(c.Service.Name); active {
		return fmt.Errorf("please stop service [%s] first", c.Service.Name)
	}

	if util.IsServiceEnabled(c.Service.Name) {
		if err := util.DisableService(c.Service.Name); err != nil {
			hwlog.RunLog.Errorf("disable service [%s] failed, error: %v", c.Service.Name, err)
			return fmt.Errorf("disable service [%s] failed", c.Service.Name)
		}
	}
	if err := util.RemoveServiceFileInSystemd(c.Service.Name); err != nil {
		hwlog.RunLog.Errorf("remove service [%s] failed, error: %v", c.Service.Name, err)
		return fmt.Errorf("remove service [%s] failed", c.Service.Name)
	}
	hwlog.RunLog.Infof("unregister service [%s] success", c.Service.Name)
	return nil
}

// IsExist is component exist
func (c Component) IsExist() bool {
	return fileutils.IsExist(c.Bin.Path)
}
