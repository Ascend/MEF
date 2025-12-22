// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package pathmgr config path manager
package pathmgr

import (
	"path/filepath"

	"edge-installer/pkg/common/constants"
)

// ConfigPathMgr config path manager
type ConfigPathMgr struct {
	installRootDir string
}

// NewConfigPathMgr new config path manager
func NewConfigPathMgr(installRootDir string) *ConfigPathMgr {
	return &ConfigPathMgr{
		installRootDir: installRootDir,
	}
}

// GetInstallRootDir get install root dir. default: /usr/local/mindx
func (cpm *ConfigPathMgr) GetInstallRootDir() string {
	return cpm.installRootDir
}

// GetMefEdgeDir get mef edge dir. default: /usr/local/mindx/MEFEdge
func (cpm *ConfigPathMgr) GetMefEdgeDir() string {
	return filepath.Join(cpm.GetInstallRootDir(), constants.MEFEdgeName)
}

// GetConfigDir get config dir. default: /usr/local/mindx/MEFEdge/config
func (cpm *ConfigPathMgr) GetConfigDir() string {
	return filepath.Join(cpm.GetMefEdgeDir(), constants.Config)
}

// GetTempCertsDir get temp certs dir. default: /usr/local/mindx/MEFEdge/config/tmp_certs
func (cpm *ConfigPathMgr) GetTempCertsDir() string {
	return filepath.Join(cpm.GetConfigDir(), constants.TmpCerts)
}

// GetConfigBackupDir get config backup dir. default: /usr/local/mindx/MEFEdge/config_backup
func (cpm *ConfigPathMgr) GetConfigBackupDir() string {
	return filepath.Join(cpm.GetMefEdgeDir(), constants.ConfigBackup)
}

// GetConfigBackupTempDir get config backup temp dir. default: /usr/local/mindx/MEFEdge/config_backup_temp
func (cpm *ConfigPathMgr) GetConfigBackupTempDir() string {
	return filepath.Join(cpm.GetMefEdgeDir(), constants.ConfigBackupTmp)
}

// GetCompConfigDir get component config dir. e.g. /usr/local/mindx/MEFEdge/config/edge_main
func (cpm *ConfigPathMgr) GetCompConfigDir(component string) string {
	return filepath.Join(cpm.GetConfigDir(), component)
}

// GetCompKmcDir get kmc dir. e.g. /usr/local/mindx/MEFEdge/config/edge_main/kmc
func (cpm *ConfigPathMgr) GetCompKmcDir(component string) string {
	return filepath.Join(cpm.GetCompConfigDir(component), constants.KmcDir)
}

// GetCompKmcConfigPath get component kmc config path. e.g. /usr/local/mindx/MEFEdge/config/edge_main/kmc-config.json
func (cpm *ConfigPathMgr) GetCompKmcConfigPath(component string) string {
	return filepath.Join(cpm.GetCompConfigDir(component), constants.KmcCfgName)
}

// GetCompInnerCertsDir get component inner cert dir. e.g. /usr/local/mindx/MEFEdge/config/edge_main/inner_certs
func (cpm *ConfigPathMgr) GetCompInnerCertsDir(component string) string {
	return filepath.Join(cpm.GetCompConfigDir(component), constants.InnerCertPathName)
}

// GetCompInnerRootCertPath get component inner root cert path.
// e.g. /usr/local/mindx/MEFEdge/config/edge_main/inner_certs/root.crt
func (cpm *ConfigPathMgr) GetCompInnerRootCertPath(component string) string {
	return filepath.Join(cpm.GetCompInnerCertsDir(component), constants.RootCaName)
}

// GetCompInnerSvrCertPath get component inner server/client cert path.
// e.g. /usr/local/mindx/MEFEdge/config/edge_main/inner_certs/server.crt
func (cpm *ConfigPathMgr) GetCompInnerSvrCertPath(component string) string {
	if component == constants.EdgeMain {
		return filepath.Join(cpm.GetCompInnerCertsDir(component), constants.ServerCertName)
	}
	return filepath.Join(cpm.GetCompInnerCertsDir(component), constants.ClientCertName)
}

// GetCompInnerSvrKeyPath get component inner server/client key path.
// e.g. /usr/local/mindx/MEFEdge/config/edge_main/inner_certs/server.key
func (cpm *ConfigPathMgr) GetCompInnerSvrKeyPath(component string) string {
	if component == constants.EdgeMain {
		return filepath.Join(cpm.GetCompInnerCertsDir(component), constants.ServerKeyName)
	}
	return filepath.Join(cpm.GetCompInnerCertsDir(component), constants.ClientKeyName)
}

// GetTempRootCertDir get temp root_certs dir. default: /usr/local/mindx/MEFEdge/config/edge_installer/root_certs
func (cpm *ConfigPathMgr) GetTempRootCertDir() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeInstaller), constants.RootCaDir)
}

// GetTempRootCertPath get temp root cert path.
// default: /usr/local/mindx/MEFEdge/config/edge_installer/root_certs/root.crt
func (cpm *ConfigPathMgr) GetTempRootCertPath() string {
	return filepath.Join(cpm.GetTempRootCertDir(), constants.RootCaName)
}

// GetTempRootCerKeyPath get temp root cert key path.
// default: /usr/local/mindx/MEFEdge/config/edge_installer/root_certs/root.key
func (cpm *ConfigPathMgr) GetTempRootCerKeyPath() string {
	return filepath.Join(cpm.GetTempRootCertDir(), constants.RootCaKeyName)
}

// GetSnPath get serial-number.json path. default: /usr/local/mindx/MEFEdge/config/edge_installer/serial-number.json
func (cpm *ConfigPathMgr) GetSnPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeInstaller), constants.SnFileName)
}

// GetDockerBackupPath get docker.service.bak path.
// default: /usr/local/mindx/MEFEdge/config/edge_installer/docker.service.bak
func (cpm *ConfigPathMgr) GetDockerBackupPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeInstaller), constants.DockerServiceBackupFile)
}

// GetEdgeMainDbPath get edgemain db path. default: /usr/local/mindx/MEFEdge/config/edge_main/edge_main.db
func (cpm *ConfigPathMgr) GetEdgeMainDbPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeMain), constants.DbEdgeMainPath)
}

// GetOMCertDir get mindXOm cert dir. default: /usr/local/mindx/MEFEdge/config/edge_main/peer_certs/mindXOM
func (cpm *ConfigPathMgr) GetOMCertDir() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeMain), constants.PeerCerts, constants.MindXOMDir)
}

// GetOMRootCertPath get mindXOm cert path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/peer_certs/mindXOM/root.crt
func (cpm *ConfigPathMgr) GetOMRootCertPath() string {
	return filepath.Join(cpm.GetOMCertDir(), constants.RootCaName)
}

// GetContainerConfigPath get container-config.json path.
// default: /usr/local/mindx/MEFEdge/config/edge_om/container-config.json
func (cpm *ConfigPathMgr) GetContainerConfigPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeOm), constants.ContainerCfgFile)
}

// GetPodConfigPath get pod-config.json path. default: /usr/local/mindx/MEFEdge/config/edge_om/pod-config.json
func (cpm *ConfigPathMgr) GetPodConfigPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeOm), constants.PodCfgFile)
}

// GetImageCertPath get image cert path. default: /usr/local/mindx/MEFEdge/config/edge_om/image_certs/root.crt
func (cpm *ConfigPathMgr) GetImageCertPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeOm), constants.ImageCertPathName, constants.ImageCertFileName)
}

// GetEdgeOmDbPath get edgeom db path. default: /usr/local/mindx/MEFEdge/config/edge_om/edge_om.db
func (cpm *ConfigPathMgr) GetEdgeOmDbPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeOm), constants.DbEdgeOmPath)
}

// GetEdgeCoreDbPath get edgecore db path. default: /usr/local/mindx/MEFEdge/config/edge_core/edgecore.db
func (cpm *ConfigPathMgr) GetEdgeCoreDbPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeCore), constants.DbEdgeCorePath)
}

// GetEdgeCoreConfigPath get edgecore.json path. default: /usr/local/mindx/MEFEdge/config/edge_core/edgecore.json
func (cpm *ConfigPathMgr) GetEdgeCoreConfigPath() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeCore), constants.EdgeCoreJsonName)
}
