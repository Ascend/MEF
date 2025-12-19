// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package pathmgr config path manager for sdk
package pathmgr

import (
	"path/filepath"

	"edge-installer/pkg/common/constants"
)

// GetTempCrlPath get temp crl path. default: /usr/local/mindx/MEFEdge/config/tmp_crls/tmp.crl
func (cpm *ConfigPathMgr) GetTempCrlPath() string {
	return filepath.Join(cpm.GetConfigDir(), constants.TmpCrls, constants.TmpCrlFile)
}

// GetHubSvrCertDir get hub_svr cert dir. default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import
func (cpm *ConfigPathMgr) GetHubSvrCertDir() string {
	return filepath.Join(cpm.GetCompConfigDir(constants.EdgeMain), constants.MefCertImportPathName)
}

// GetHubSvrCertPath get hub_svr cert path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import/edge_hub.crt
func (cpm *ConfigPathMgr) GetHubSvrCertPath() string {
	return filepath.Join(cpm.GetHubSvrCertDir(), constants.HubSvrCrtName)
}

// GetHubSvrKeyPath get hub_svr key path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import/edge_hub.key
func (cpm *ConfigPathMgr) GetHubSvrKeyPath() string {
	return filepath.Join(cpm.GetHubSvrCertDir(), constants.HubSvrKeyName)
}

// GetHubSvrRootCertPath get hub_svr root cert path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import/cloud_root.crt
func (cpm *ConfigPathMgr) GetHubSvrRootCertPath() string {
	return filepath.Join(cpm.GetHubSvrCertDir(), constants.RootCertName)
}

// GetHubSvrRootCertBackupPath get hub_svr root cert backup path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import/cloud_root.crt.backup
func (cpm *ConfigPathMgr) GetHubSvrRootCertBackupPath() string {
	return filepath.Join(cpm.GetHubSvrCertDir(), constants.RootCertBackUpName)
}

// GetHubSvrCrlPath get hub_svr crl path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import/cloud_root.crl
func (cpm *ConfigPathMgr) GetHubSvrCrlPath() string {
	return filepath.Join(cpm.GetHubSvrCertDir(), constants.RootCrlName)
}

// GetHubSvrRootCertPrevBackupPath get hub_svr previous backup root cert path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import/cloud_root.crt.pre
func (cpm *ConfigPathMgr) GetHubSvrRootCertPrevBackupPath() string {
	return filepath.Join(cpm.GetHubSvrCertDir(), constants.RootCertPrevBackUpName)
}

// GetHubSvrTempCertPath get hub_svr temp cert path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import/edge_hub.crt.tmp
func (cpm *ConfigPathMgr) GetHubSvrTempCertPath() string {
	return filepath.Join(cpm.GetHubSvrCertDir(), constants.HubSvrTempCrt)
}

// GetHubSvrTempKeyPath get hub_svr temp key path.
// default: /usr/local/mindx/MEFEdge/config/edge_main/hub_certs_import/edge_hub.key.tmp
func (cpm *ConfigPathMgr) GetHubSvrTempKeyPath() string {
	return filepath.Join(cpm.GetHubSvrCertDir(), constants.HubSvrTempKey)
}

// GetNetConfigTempDir get net config temp dir. default: /usr/local/mindx/MEFEdge/config/temp_netconfig
func (cpm *ConfigPathMgr) GetNetConfigTempDir() string {
	return filepath.Join(cpm.GetConfigDir(), constants.NetCfgTempDirName)
}

// GetNetCfgTempRootCertPath get net config temp root cert path.
// default: /usr/local/mindx/MEFEdge/config/temp_netconfig/cloud_root.crt
func (cpm *ConfigPathMgr) GetNetCfgTempRootCertPath() string {
	return filepath.Join(cpm.GetNetConfigTempDir(), constants.RootCertName)
}

// GetNetCfgTempRootCertBackupPath get net config temp root cert backup path.
// default: /usr/local/mindx/MEFEdge/config/temp_netconfig/cloud_root.crt.backup
func (cpm *ConfigPathMgr) GetNetCfgTempRootCertBackupPath() string {
	return filepath.Join(cpm.GetNetConfigTempDir(), constants.RootCertBackUpName)
}
