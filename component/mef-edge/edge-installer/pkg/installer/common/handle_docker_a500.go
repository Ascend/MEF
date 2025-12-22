// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package common
package common

import (
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

func (c ComponentMgr) changeDocker() error {
	if !config.CheckIsA500() {
		hwlog.RunLog.Info("this is not a Atlas 500 A2 device, no need change docker")
		return nil
	}
	dockerIsolationShPath := c.workPathMgr.GetDockerIsolationShPath()
	dockerServicePath := c.workPathMgr.GetServicePath(constants.DockerServiceFile)
	installerConfigDir := c.configPathMgr.GetCompConfigDir(constants.EdgeInstaller)
	realDockerIsolationSh, err := fileutils.EvalSymlinks(dockerIsolationShPath)
	if err != nil {
		hwlog.RunLog.Errorf("change docker isolation failed: %v", err)
		return err
	}
	out, err := envutils.RunCommand(realDockerIsolationSh, envutils.DefCmdTimeoutSec,
		dockerServicePath, installerConfigDir)
	if err != nil {
		hwlog.RunLog.Errorf("execute change docker isolation cmd failed: output: %s, err:%v", out, err)
		return err
	}
	hwlog.RunLog.Infof("execute change docker isolation cmd succeeded: output: %s", out)
	return nil
}
