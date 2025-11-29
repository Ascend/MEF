// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
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
