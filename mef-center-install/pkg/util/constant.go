// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"time"

	"huawei.com/mindxedge/base/common"
)

// Command constant for command
const (
	CommandKubectl     = "kubectl"
	CommandNamespace   = "namespace"
	Haveged            = "haveged"
	ArchCommand        = "uname"
	Arch64             = "aarch64"
	X86                = "x86_64"
	ActiveFlag         = "Active"
	ReadyFlag          = "1/1"
	StopFlag           = "0/0"
	StopReplicasNum    = 0
	StartReplicasNum   = 1
	DockerImageExist   = 2
	NamespaceExist     = 1
	DeleteNsTimeoutSec = 300
)

// MEF-Center Dir constant
const (
	OutMefDirName     = "MEF-Center"
	MefSoftLink       = "mef-center"
	MefWorkA          = "mef-center-A"
	MefWorkB          = "mef-center-B"
	TempUpgradeDir    = "temp-upgrade"
	MefConfigDir      = "mef-config"
	InstallPackageDir = "install-package"
)

// MEF-Center File constant
const (
	MefRunScript     = "run.sh"
	VersionXml       = "version.xml"
	InstallParamJson = "install-param.json"
	InstallBin       = "MEF-center-installer"
)

// single WorkDir constant
const (
	MefLibDir           = "lib"
	MefVarDir           = "var"
	MefZipDir           = "zip"
	MefTarDir           = "tar"
	ComponentLibDir     = "lib"
	MefKmcLibDir        = "kmc-lib"
	ImageConfigDir      = "image-config"
	ImageDir            = "image"
	ImagesDirName       = "images"
	MefBinDir           = "bin"
	DockerFileName      = "Dockerfile"
	NginxDirName        = "nginx"
	ImageTarNamePattern = "Ascend-mef-%s-linux-%s.tar.gz"
	ImagePrefix         = "ascend-"
)

// MEF-Config constant
const (
	KmcDir        = "kmc"
	RootCaDir     = "root-ca"
	RootCaFileDir = "cert"
	RootCaKeyDir  = "key"
	RootCaFile    = "RootCA.crt"
	RootKeyFile   = "RootCA.key"
	MasterKeyFile = "master.ks"
	BackUpKeyFile = "backup.ks"
	CertsDir      = "mef-certs"
	CertSuffix    = ".crt"
	KeySuffix     = ".key"
	CaCommonName  = "MindX MEF"
)

// log constant
const (
	ModuleLogName       = "mef-center-log"
	ModuleLogBackupName = "mef-center-log-backup"
	MefScriptsDir       = "scripts"
	RunLogFile          = "mef-center-install.log"
	OperateLogFile      = "mef-center-install-operate.log"
	InstallLogDir       = "mef-center-install"
)

// module name constant
const (
	EdgeManagerName     = "edge-manager"
	CertManagerName     = "cert-manager"
	NginxManagerName    = "nginx-manager"
	SoftwareManagerName = "software-manager"
	UserManagerName     = "user-manager"
)

// install constant
const (
	SoftwareManagerFlag = SoftwareManagerName
	AllInstallFlag      = "install_all"
	LogPathFlag         = "log_path"
	LogBackupPathFlag   = "log_backup_path"
	InstallPathFlag     = "install_path"
	HelpFlag            = "help"
	HelpShortFlag       = "h"
	VersionFlag         = "version"
	MefCenterUid        = 8000
	RootUid             = 0
	MefCenterGid        = 8000
	RootGid             = 0
	MefCenterName       = "MEFCenter"
	MefCenterGroup      = "MEFCenter"
	DockerTag           = "v1"
)

// yaml editor constant
const (
	RootCaFlag          = "${root-ca}"
	LogFlag             = "${log}"
	LogBackupFlag       = "${log-backup}"
	ConfigFlag          = "${config}"
	InstalledModuleFlag = "${installed_module}"
	LineSplitter        = "\n"
	SplitCount          = 2
)

// constant for install
const (
	MefNamespace    = "mef-center"
	RootUserName    = "root"
	AscendPrefix    = "ascend-"
	HelpExitCode    = 3
	VersionExitCode = 3
	ErrorExitCode   = 1
	RunFlagCount    = 3

	InstallDiskSpace    = 750 * common.MB
	CheckStatusInterval = 3 * time.Second
	CheckStatusTimes    = 5
)

// constant for mef control bin
const (
	OperateFlag   = "operate"
	UninstallFlag = "uninstall"
	UpgradeFlag   = "upgrade"

	StartOperateFlag   = "start"
	StopOperateFlag    = "stop"
	RestartOperateFlag = "restart"
)

// constant for set k8s label
const (
	K8sLabel             = "mef-center-node="
	GetNodeCmdPattern    = "kubectl get nodes -o wide | grep -w %s | awk '{print$1}'"
	SetLabelCmdPattern   = "kubectl label nodes %s --overwrite %s"
	CheckLabelCmdPattern = "kubectl get nodes -l %s | grep -w %s | wc -l"
	LabelCount           = 1
)

// constant for upgrade
const (
	UpgradeDiskSpace = 700 * common.MB
	InstallDirName   = "installer"

	ScriptsDirName    = "scripts"
	UpgradeShName     = "upgrade.sh"
	UpgradeTimeoutSec = 420
)
