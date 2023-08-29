// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"time"

	"huawei.com/mindxedge/base/common"
)

// Command constant for command
const (
	CommandKubectl     = "kubectl"
	CommandDocker      = "docker"
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
	NamespaceStatusLoc = 1
	SplitStringCount   = -1
	DeleteNsTimeoutSec = 300
)

// MEF-Center Dir constant
const (
	OutMefDirName     = "MEF-Center"
	MefSoftLink       = "mef-center"
	MefWorkA          = "mef-center-A"
	MefWorkB          = "mef-center-B"
	TempUpgradeDir    = "temp-upgrade"
	TempCertDir       = "temp-cert"
	MefConfigDir      = "mef-config"
	InstallPackageDir = "install-package"
	MefCenterLock     = "mef-center.lock"
)

// MEF-Center File constant
const (
	MefRunScript     = "run.sh"
	VersionXml       = "version.xml"
	InstallParamJson = "install-param.json"
	ControllerBin    = "MEF-center-controller"
	UpgradeFlagFile  = "upgrade-flag"
	ConfigInPkg      = "config"
)

// single WorkDir constant
const (
	MefLibDir           = "lib"
	OtherLibDir         = "lib"
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
	KmcDir           = "kmc"
	PubCfgDir        = "public-config"
	RootCaDir        = "root-ca"
	RootCaFileDir    = "cert"
	RootCaKeyDir     = "key"
	RootCaFile       = "RootCA.crt"
	RootKeyFile      = "RootCA.key"
	MasterKeyFile    = "master.ks"
	BackUpKeyFile    = "backup.ks"
	CertsDir         = "mef-certs"
	KubeCertsDir     = "kube-config"
	CertSuffix       = ".crt"
	CrlSuffix        = ".crl"
	KeySuffix        = ".key"
	CaCommonName     = "MindX MEF"
	PubConfigDir     = "public-config"
	NorthernRootName = "northern-root"
	ApigDirName      = "apig"
	RootCrtName      = "root.crt"
)

// log constant
const (
	ModuleLogName       = "mef-center-log"
	ModuleLogBackupName = "mef-center-log-backup"
	MefScriptsDir       = "scripts"
	RunLogFile          = "mef-center-install-run.log"
	OperateLogFile      = "mef-center-install-operate.log"
	InstallLogDir       = "mef-center-install"
)

// module name constant
const (
	EdgeManagerName   = "edge-manager"
	CertManagerName   = "cert-manager"
	NginxManagerName  = "nginx-manager"
	AlarmManagerName  = "alarm-manager"
	MefCenterRootName = "MEF-Center-root"
)

// install constant
const (
	AllInstallFlag      = "install_all"
	LogPathFlag         = "log_path"
	LogBackupPathFlag   = "log_backup_path"
	InstallPathFlag     = "install_path"
	CloudCoreCaPathFlag = "cloud_core_ca_path"
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
	PubCfgFlag          = "${public-config}"
	LogFlag             = "${log}"
	LogBackupFlag       = "${log-backup}"
	ConfigFlag          = "${config}"
	InstalledModuleFlag = "${installed_module}"
	LineSplitter        = "\n"
	SplitCount          = 2
)

// constant for install
const (
	MefNamespace        = "mef-center"
	RootUserName        = "root"
	AscendPrefix        = "ascend-"
	HelpExitCode        = 3
	VersionExitCode     = 3
	NotGenCertErrorCode = 4
	ErrorExitCode       = 1

	InstallDiskSpace    = 750 * common.MB
	LogDiskSpace        = 1800 * common.MB
	LogBackupDiskSpace  = 3500 * common.MB
	CheckStatusInterval = 3 * time.Second
	CheckStatusTimes    = 5
)

const (
	// NotGenCertErrorStr for the root ca has not yet generated
	NotGenCertErrorStr = "the root ca has not yet generated"
)

// constant for mef control bin
const (
	UninstallFlag  = "uninstall"
	UpgradeFlag    = "upgrade"
	ExchangeCaFlag = "exchangeca"
	UpdateKmcFlag  = "updatekmc"
	ImportCrlFlag  = "importcrl"

	StartOperateFlag   = "start"
	StopOperateFlag    = "stop"
	RestartOperateFlag = "restart"
)

// constant for set k8s label
const (
	K8sLabelSet    = "mef-center-node="
	K8sLabel       = "mef-center-node"
	NodeSplitCount = 1
)

// constant for upgrade
const (
	UpgradeDiskSpace = 700 * common.MB
	InstallDirName   = "installer"

	ScriptsDirName    = "scripts"
	UpgradeShName     = "upgrade.sh"
	UpgradeTimeoutSec = 420
)

// constant for the step in upgrade to clear environment
const (
	ClearUnpackPathStep = iota
	ClearTempUpgradePathStep
	RestartPodStep
	LoadOldDockerStep
	RemoveDockerStep
	ClearNameSpaceStep
)

// constant for parse args
const (
	NoArgCount  = 1
	CtlArgIndex = 1
	CmdIndex    = 1
	CmdArgIndex = 2
)

// const for permission_mgr
const (
	DirType  = "dir"
	FileType = "file"
)

const (
	// ServerCertPath  cert-manager server cert path
	ServerCertPath = "/home/data/config/mef-certs/cert-manager.crt"
	// ServerKeyPath cert-manager server key path
	ServerKeyPath = "/home/data/config/mef-certs/cert-manager.key"
	// RootCaPath  cert-manager server root ca path
	RootCaPath = "/home/data/inner-root-ca/RootCA.crt"
	// RootCaMgrDir root ca save directory
	RootCaMgrDir = "/home/data/config/root-ca/"
	// RootCaFileName root ca save file name
	RootCaFileName = "root.crt"
	// RootKeyFileName root key save file name
	RootKeyFileName = "encrypt_root.key"
	// CrlName root ca save file name
	CrlName = "revokeList.crl"
	// ServiceName for edge-manager kubeconfig certs
	ServiceName = "server.crt"
	// KeyFileName for edge-manager kubeconfig key
	KeyFileName = "server.key"
)

// compare crls result status while upgrading
const (
	// CompareSame two crls are same
	CompareSame int = 0
	// CompareNew crl to update signed time is newer
	CompareNew int = 1
	// CompareOld crl to update signed time is older
	CompareOld int = 2
)

// Crl on device dir and name
const (
	CrlOnDeviceDir  = "/etc/hwsipcrl"
	CrlOnDeviceName = "ascendsip.crl"
)
