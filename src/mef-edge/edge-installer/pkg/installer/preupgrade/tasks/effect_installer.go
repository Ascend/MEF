// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for effect installer
package tasks

import (
	"fmt"
	"path/filepath"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
)

type effectInstaller struct {
	edgeDir string
}

// EffectInstaller effect installer
func EffectInstaller(edgeDir string) common.Task {
	return &effectInstaller{edgeDir: edgeDir}
}

// Run task
func (e *effectInstaller) Run() error {
	if err := e.oldVersionSmooth(); err != nil {
		return err
	}

	installer := UpgradeInstaller(e.edgeDir, constants.EffectMode)
	if err := installer.Run(); err != nil {
		return fmt.Errorf("effect %s failed: %v", constants.MEFEdgeName, err)
	}
	return nil
}

func (e *effectInstaller) oldVersionSmooth() error {
	installRootDir := filepath.Dir(e.edgeDir)
	upgradeVersionPath := pathmgr.NewWorkPathMgr(installRootDir).GetUpgradeTempVersionXmlPath()
	versionMgr := config.NewVersionXmlMgr(upgradeVersionPath)
	upgradeVersion, err := versionMgr.GetVersion()
	if err != nil {
		return fmt.Errorf("get upgrade inner version failed, error: %v", err)
	}

	if err := config.EffectToOldestVersionSmooth(upgradeVersion, installRootDir); err != nil {
		return fmt.Errorf("smooth config file for old version failed, error: %v", err)
	}
	return nil
}
