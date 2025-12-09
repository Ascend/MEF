// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common this file for edge control constants
package common

// edge control commands
const (
	Start             = "start"
	Stop              = "stop"
	Restart           = "restart"
	Uninstall         = "uninstall"
	Upgrade           = "upgrade"
	GetNetCfg         = "getnetconfig"
	CollectLog        = "collectlog"
	GetCertInfo       = "getcertinfo"
	Effect            = "effect"
	UpdateKmc         = "updatekmc"
	ImportCrl         = "importcrl"
	UpdateCrl         = "updatecrl"
	GetUnusedCertInfo = "getunusedcert"
	DeleteCert        = "deletecert"
	RestoreCert       = "restorecert"
)

// const for inner control commands
const (
	ExchangeCertsCmd   = "exchange_certs"
	RecoveryCmd        = "recovery"
	PrepareEdgecoreCmd = "prepare_edgecore"
	RecoverLogCmd      = "recover_log"

	ImportPathFlag = "import_path"
	ExportPathFlag = "export_path"
)

// edge control command descriptions
const (
	StartDesc             = "to start all edge component"
	StopDesc              = "to stop all edge component"
	RestartDesc           = "to restart all edge component"
	UninstallDesc         = "to uninstall the software"
	UpgradeDesc           = "to upgrade the software"
	GetNetCfgDesc         = "to get current net config"
	CollectLogDesc        = "to collect the log"
	GetCertInfoDesc       = "to print certificate information"
	EffectDesc            = "to effect the software"
	UpdateKmcDesc         = "to update kmc key"
	ImportCrlDesc         = "to import crl of the certificate used for cloud edge interconnection"
	UpdateCrlDesc         = "to update crl file for package signature verification"
	GetUnusedCertInfoDesc = "to print previous backup certificate info"
	DeleteUnusedCertDesc  = "to delete unused previous backup certificate"
	RestoreCertDesc       = "to restore previous backup certificate"
)

// inner control commands descriptyions
const (
	ExchangeCertsDesc   = "to exchange root ca with OM"
	RecoveryDesc        = "to recovery the environment when upgrading firmware failed"
	PrepareEdgecoreDesc = "to prepare edgecore"
	RecoverLogDesc      = "to recover log from disk to memory"
)
