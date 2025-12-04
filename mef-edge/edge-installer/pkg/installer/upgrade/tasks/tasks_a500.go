// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package tasks for some methods that are performed only on the a500 device
package tasks

import (
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/common"
)

func (swp *SetWorkPathTask) prepareCfgBackupDir() error {
	cfgBackupDir := swp.PathMgr.ConfigPathMgr.GetConfigBackupTempDir()
	if err := fileutils.DeleteAllFileWithConfusion(cfgBackupDir); err != nil {
		return fmt.Errorf("clean config backup dir failed, error: %v", err)
	}

	if err := fileutils.CreateDir(cfgBackupDir, constants.Mode755); err != nil {
		return fmt.Errorf("create config backup dir failed, error: %v", err)
	}

	hwlog.RunLog.Info("create config backup dir success")
	return nil
}

func (p *PostEffectProcessTask) copyResetScriptToP7() error {
	if err := common.CopyResetScriptToP7(); err != nil {
		hwlog.RunLog.Warn(err)
	}
	return nil
}

func (p *PostEffectProcessTask) smoothConfig() error {
	return p.smoothCommonConfig()
}

// refreshDefaultCfgDir effect default config backup dir for recovery
func (p *PostEffectProcessTask) refreshDefaultCfgDir() error {
	cfgBackupTempDir := p.ConfigPathMgr.GetConfigBackupTempDir()
	cfgBackupDir := p.ConfigPathMgr.GetConfigBackupDir()

	if err := fileutils.DeleteAllFileWithConfusion(cfgBackupDir); err != nil {
		return fmt.Errorf("remove old config backup dir failed, error: %v", err)
	}
	if err := fileutils.RenameFile(cfgBackupTempDir, cfgBackupDir); err != nil {
		return fmt.Errorf("rename config backup temp directory to [%s] failed, error: %v", cfgBackupDir, err)
	}
	hwlog.RunLog.Info("effect config backup dir success")
	return nil
}
