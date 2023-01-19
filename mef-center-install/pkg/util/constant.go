// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"time"

	"huawei.com/mindxedge/base/common"
)

// Command constant for command
const (
	CommandKubectl   = "kubectl"
	CommandNamespace = "namespace"
	CommandCopy      = "cp"
	ArchCommand      = "uname"
	Arch64           = "aarch64"
	X86              = "x86_64"
	ActiveFlag       = "Active"
	ReadyFlag        = "1/1"
	StopFlag         = "0/0"
	StopReplicasNum  = 0
	StartReplicasNum = 1
	IllegalChars     = "\n!\\; &$<>`"
)

// MEF-Center Dir constant
const (
	OutMefDirName     = "MEF-Center"
	MefSoftLink       = "mef-center"
	MefWorkA          = "mef-center-A"
	MefWorkB          = "mef-center-B"
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
	KmcDir            = "kmc"
	RootCaDir         = "root-ca"
	RootCaFileDir     = "cert"
	RootCaKeyDir      = "key"
	RootCaFile        = "RootCA.crt"
	RootKeyFile       = "RootCA.key"
	MasterKeyFile     = "master.ks"
	BackUpKeyFile     = "backup.ks"
	CertsDir          = "mef-certs"
	CertSuffix        = ".crt"
	KeySuffix         = ".key"
	CaCommonName      = "MindX MEF"
	NginxServerSuffix = "-server"
)

// log constant
const (
	ModuleLogName  = "mef-center-log"
	MefScriptsDir  = "scripts"
	RunLogFile     = "mef-center-install.log"
	OperateLogFile = "mef-center-install-operate.log"
	InstallLogDir  = "mef-center-install"
)

// module name constant
const (
	EdgeManagerName     = "edge-manager"
	CertManagerName     = "cert-manager"
	NginxManagerName    = "nginx-manager"
	ImageManagerName    = "image-manager"
	ResourceManagerName = "resource-manager"
	SoftwareManagerName = "software-manager"
)

// install constant
const (
	ImageManagerFlag    = ImageManagerName
	ResourceManagerFlag = ResourceManagerName
	SoftwareManagerFlag = SoftwareManagerName
	AllInstallFlag      = "install_all"
	LogPathFlag         = "log_path"
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
	RootCaFlag          = "root-ca"
	LogSuffix           = "-log"
	ConfigSuffix        = "-config"
	InstalledModuleName = "installed-module"

	PathSplitter  = "path:"
	LineSplitter  = "\n"
	ValueSplitter = "value:"

	ComponentSplitCount       = 3
	PathSplitCount            = 2
	LineSplitCount            = 2
	InstalledModuleSpiltCount = 2
	ValueSplitCount           = 2
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
