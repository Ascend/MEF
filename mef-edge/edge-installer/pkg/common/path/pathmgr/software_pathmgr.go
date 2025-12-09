// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package pathmgr software path manager
package pathmgr

import (
	"path/filepath"

	"edge-installer/pkg/common/constants"
)

// SoftwarePathMgr software path manager, including work path manager and config path manager
type SoftwarePathMgr struct {
	// installation package path. e.g. /home/xxx
	installationPkgDir string
	// default: /usr/local/mindx
	installRootDir string
	WorkPathMgr    *WorkPathMgr
	ConfigPathMgr  *ConfigPathMgr
}

// NewSoftwarePathMgr new software path manager, including work path manager and config path manager
func NewSoftwarePathMgr(installRootDir, installationPkgDir string) *SoftwarePathMgr {
	return &SoftwarePathMgr{
		installationPkgDir: installationPkgDir,
		installRootDir:     installRootDir,
		WorkPathMgr:        NewWorkPathMgr(installRootDir),
		ConfigPathMgr:      NewConfigPathMgr(installRootDir),
	}
}

// GetInstallRootDir get install root dir. default: /usr/local/mindx
func (spm *SoftwarePathMgr) GetInstallRootDir() string {
	return spm.installRootDir
}

// GetInstallationPkgDir get installation package dir.
func (spm *SoftwarePathMgr) GetInstallationPkgDir() string {
	return spm.installationPkgDir
}

// GetPkgVersionXmlPath get version.xml path in installation package. default: /{installation package path}/version.xml
func (spm *SoftwarePathMgr) GetPkgVersionXmlPath() string {
	return filepath.Join(spm.GetInstallationPkgDir(), constants.VersionXml)
}

// GetPkgCompSoftwareDir get component software dir in installation package.
// e.g. /{installation package path}/software/edge_installer
func (spm *SoftwarePathMgr) GetPkgCompSoftwareDir(component string) string {
	return filepath.Join(spm.GetInstallationPkgDir(), constants.SoftwareDir, component)
}

// GetPkgLibDir get lib dir in installation package. default: /{installation package path}/software/lib
func (spm *SoftwarePathMgr) GetPkgLibDir() string {
	return filepath.Join(spm.GetInstallationPkgDir(), constants.SoftwareDir, constants.Lib)
}

// GetPkgRunShPath get run.sh path in installation package. default: /{installation package path}/software/run.sh
func (spm *SoftwarePathMgr) GetPkgRunShPath() string {
	return filepath.Join(spm.GetInstallationPkgDir(), constants.SoftwareDir, constants.RunScript)
}

// GetPkgCompConfigDir get component config dir in installation package.
// default: /{installation package path}/config/edge_installer
func (spm *SoftwarePathMgr) GetPkgCompConfigDir(component string) string {
	return filepath.Join(spm.GetInstallationPkgDir(), constants.Config, component)
}
