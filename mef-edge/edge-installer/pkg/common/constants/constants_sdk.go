// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package constants for
package constants

// MEFEdge SDK ports
const (
	DefaultWsPort            = 30003
	DefaultWsTestPort        = 30004
	DefaultCloudCoreCertPort = 10002
	DefaultCloudCoreWsPort   = 10000
	MinPort                  = 1025
	MaxPort                  = 65535
)

// RootCertName cloud core cert
const (
	TempSuffix               = ".tmp"
	RootCertName             = "cloud_root.crt"
	RootCertBackUpName       = "cloud_root.crt.backup"
	RootCrlName              = "cloud_root.crl"
	RootCertPrevBackUpName   = "cloud_root.crt.pre"
	HubSvrKeyName            = "edge_hub.key"
	HubSvrCrtName            = "edge_hub.crt"
	MefCenterTokenUrl        = "/token"
	MefCenterConnTestUrl     = "/token-check"
	MefCertImportPathName    = "hub_certs_import"
	HubSvrTempKey            = HubSvrKeyName + TempSuffix
	HubSvrTempCrt            = HubSvrCrtName + TempSuffix
	LinkCert                 = "imageWarehouse.crt"
	ImageCertName            = "image"
	CommonRevocationListName = "root.crl"
)

// invalid ip
const (
	IpZero      = "0.0.0.0"
	IpBroadcast = "255.255.255.255"
)

// config key constants
const (
	ImageCfgKey  = "imageCfgKey"
	DomainCfgKey = "domainCfgKey"
)

// MEFEdgeSDK upgrade manager constants
const (
	UpgradeManagerName  = "UpgradeManager"
	DownloadManagerName = "DownloadManager"

	EdgeInstallerFileName      = "edge-installer"
	LogCollectTempDir          = "/home/data/mef_logcollect"
	LogCollectTempFileName     = "edgeNode.tar.gz"
	BackUpPkgPath              = "/home/data/mefedge/backup"
	PkgPath                    = "/home/data/mefedge/package"
	EdgeDownloadPath           = "/home/data/MEFEdgeDownload"
	ConfigCertPathName         = "root-ca"
	InstallerExtractOnlineMin  = 220 * MB
	InstallerExtractWithZipMin = 290 * MB
	InstallerUpgradeSdkMin     = 220 * MB

	CertSizeLimited = 1 * MB
	CrlSizeLimit    = 1 * MB
)

// location of software name and version in url
const (
	UnpackZipDir = "zip"
	ZipExt       = ".zip"
	TarGz        = "*.tar.gz"
	TarGzExt     = ".tar.gz"
	CmsExt       = ".tar.gz.cms"
	CrlExt       = ".tar.gz.crl"
	SignExt      = ".tar.gz.sig"
)

// MEF-Edge message resource constants
const (
	// ResEdgeCloudConnection is resource for edge-main to report if connection of edge-center is ready
	ResEdgeCloudConnection = "/edge-cloud/connection"
	// ResEdgeDownloadInfo is resource for downloading software before online upgrading
	ResEdgeDownloadInfo = "/edge/download"
	// ResUpgradeInfo is resource for online upgrading
	ResUpgradeInfo = "/edge/upgrade"
	// InnerPrepareDir resource that edge-main request edge-om for preparing work dir of download
	InnerPrepareDir = "/inner/preparedir"
	// InnerSoftwareVerification resource edge-main request edge-om for verify and unpack software downloaded online
	InnerSoftwareVerification = "/inner/software"
	// InnerSoftwareVersion resource edge-main request edge-om for report software version
	InnerSoftwareVersion = "/inner/version-info"
	// ResDumpLogTask resource for the log-dumping tasks
	ResDumpLogTask = "/logmgmt/dump/task"
	// ResPackLogRequest is resource for the log-packing request
	ResPackLogRequest = "/inner/logmgmt/pack/request"
	// ResPackLogResponse is resource for the log-packing response
	ResPackLogResponse = "/inner/logmgmt/pack/response"
)

// Location of software name and version in url
const (
	LocationMethod = 0
	URLFieldNum    = 2
)
