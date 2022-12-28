// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package install constants used in mef center
package install

// constant for installation
const (
	MefWorkPath          = "/usr/local/mef-center"
	MefComponentWorkPath = "/usr/local/mef-center/component"
	MefWorkCertDir       = "certs"
	MefWorkLogDir        = "logs"
	MefScriptsDir        = "scripts"
	MefLibsDir           = "lib"
	MefRunScript         = "run.sh"
	ScriptMode           = 0500
	MefSbinDir           = "sbin"
	LogDir               = "mef-center-install"
	CertsDir             = "MEF_certs"
	RunLogFile           = "mef-center-script.log"
	OperateLogFile       = "mef-center-script-operate.log"
	RootCaDir            = "rootCA"
	RootCaFile           = "RootCA.crt"
	RootKeyFile          = "RootCA.key"
	MEFCenterUserUid     = 8000
	MEFCenterUserGid     = 8000
	AllInstallFlag       = "install_all"
	ImageManagerFlag     = "image_manager"
	ResourceManagerFlag  = "resource_manager"
	SoftwareManagerFlag  = "software_manager"
	CertPathFlag         = "cert_path"
	LogPathFlag          = "log_path"
	EdgeManagerName      = "edge-manager"
	CertManagerName      = "cert-manager"
	NginxManagerName     = "nginx-manager"
	ImageManagerName     = "image-manager"
	ResourceManagerName  = "resource-manager"
	SoftwareManagerName  = "software-manager"
	CaCommonName         = "MindX MEF"
	MefNamespace         = "mindx-edge"
)
