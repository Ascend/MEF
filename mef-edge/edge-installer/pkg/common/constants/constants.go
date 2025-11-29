// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package constants this file for constants
package constants

import (
	"time"
)

// run and operate log file
const (
	RunLogFile     = "run.log"
	OperateLogFile = "operate.log"
)

// const A500
const (
	A500Name    = "Atlas 500 A2"
	A500NameOld = "Atlas A500 A2"
	NpuSmiCmd   = "npu-smi"
)

// program exit code
const (
	ProcessFailed = 1
	PrintInfo     = 3
)

// install parameters
const (
	InstallDirName          = "install_dir"
	LogDirName              = "log_dir"
	LogBackupDirName        = "log_backup_dir"
	DefaultInstallDir       = "/usr/local/mindx"
	DefaultLogDir           = "/var/alog"
	DefaultLogBackupDir     = "/home/log"
	MEFEdgeName             = "MEFEdge"
	MEFEdgeLogName          = "MEFEdge_log"
	MEFEdgeLogBackupName    = "MEFEdge_logbackup"
	MEFEdgeLogSyncName      = "MEFEdge_logsync"
	ConfigBackup            = "config_backup"
	ConfigBackupTmp         = "config_backup_temp"
	SoftwareDir             = "software"
	SoftwareDirA            = "software_A"
	SoftwareDirB            = "software_B"
	SoftwareDirTemp         = "software_temp"
	ServiceDir              = "service"
	Config                  = "config"
	EdgeInstaller           = "edge_installer"
	EdgeMain                = "edge_main"
	EdgeMainFileName        = "edge-main"
	EdgeOm                  = "edge_om"
	EdgeOmFileName          = "edge-om"
	EdgeCore                = "edge_core"
	EdgeCoreFileName        = "edgecore"
	EdgeCoreJsonName        = "edgecore.json"
	PeerCerts               = "peer_certs"
	MindXOMDir              = "mindXOM"
	DevicePlugin            = "device_plugin"
	DevicePluginFileName    = "device-plugin"
	Bin                     = "bin"
	Script                  = "script"
	DockerIsolate           = "docker_isolate"
	MefInitServiceName      = "mef-edge-init"
	MefInitScriptName       = "mef_init.sh"
	Var                     = "var"
	Log                     = "log"
	TmpCerts                = "tmp_certs"
	SoftwareCertName        = "software"
	TmpCrls                 = "tmp_crls"
	TmpCrlFile              = "tmp.crl"
	LogBackup               = "log_backup"
	MefEdgeTargetFile       = "mef-edge.target"
	EdgeInitServiceFile     = "mef-edge-init.service"
	EdgeOmServiceFile       = "mef-edge-om.service"
	EdgeMainServiceFile     = "mef-edge-main.service"
	DockerServiceFile       = "docker.service"
	DockerServiceBackupFile = "docker.service.bak"
	EdgeCoreServiceFile     = "edgecore.service"
	DevicePluginServiceFile = "device-plugin.service"
	VersionXml              = "version.xml"
	MaxXmlSizeTimes         = 1
	Lib                     = "lib"
	DevicePluginLogFile     = "device_plugin_run.log"
	EdgeCoreLogFile         = "edge_core_run.log"
	NetCfgTempDirName       = "temp_netconfig"
	SnFileName              = "serial-number.json"

	RunScript             = "run.sh"
	DockerIsolationScript = "mef_docker_isolation.sh"
	DockerRestoreScript   = "mef_docker_restore.sh"
	CpCmd                 = "cp"
	ChmodCmd              = "chmod"
	MkdirCmd              = "mkdir"
	DockerCmd             = "docker"
	RsyncCmd              = "rsync"
	RmCmd                 = "rm"
	ForceFlag             = "-f"

	RootUserName   = "root"
	DockerUserName = "docker"
	RootUserUid    = 0
	RootUserGid    = 0
	EdgeUserName   = "MEFEdge"
	EdgeUserGroup  = "MEFEdge"
	EdgeUserUid    = 1225
	EdgeUserGid    = 1225
	CertDirMode    = 0700
	CertFileMode   = 0400
	Mode755        = 0755
	Mode750        = 0750
	Mode711        = 0711
	Mode700        = 0700
	Mode600        = 0600
	Mode640        = 0640
	Mode500        = 0500
	Mode400        = 0400
	Mode444        = 0444

	ModeUmask022 = 0022
	ModeUmask027 = 0027
	ModeUmask077 = 0077
	ModeUmask277 = 0277
	ModeUmask177 = 0177
	ModeUmask137 = 0137
)

// service file parameters
const (
	ExecStartPattern = "ExecStart *= *(.*)"

	InstallEdgeDir     = "install_dir"
	InstallSoftWareDir = "software_dir"
	LogEdgeDir         = "log_dir"

	CheckServiceNum      = 5
	CheckServiceWaitTime = 1 * time.Second
)

// checker parameters
const (
	MaxPathLength           = 1024
	PathMatchStr            = "^/[a-z0-9A-Z_./-]+$"
	ContainerNameRegex      = "[a-z0-9]([a-z0-9-]{0,62}[a-z0-9]){0,1}"
	UUIDRegex               = "[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}"
	FdPodNameRegex          = ContainerNameRegex + "-" + UUIDRegex
	MefPodNameRegex         = "[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}-[a-z0-9]{5}"
	ConfigmapNameRegex      = "[a-z][a-z0-9-]{2,62}[a-z0-9]"
	NodeNameRegx            = "[a-z0-9]([-_a-z0-9]{0,62}[a-z0-9])?"
	ResourceVersionRegex    = "^[0-9]{0,32}$"
	MefResourceVersionRegex = "^[0-9.]{0,32}$"
	PodNameReg              = "^[0-9a-z][0-9a-z-]{1,128}$"
	ModelFileNameReg        = "^[a-zA-Z0-9_.-]{1,256}$"
	VersionLenReg           = "^[a-z0-9]([a-z0-9._-]{0,62}[a-z0-9]){0,1}$"
	CheckCodeReg            = "^[0-9a-f]{64}$"
	ModelFileUserNameReg    = "^[a-zA-Z0-9-_]{1,64}$"
	ModeFilePwdReg          = "^.{8,64}$"
	ModelFileMaxSize        = 4 * 1024 * MB
	ResourceNPU             = "npu"
)

// net config parameters
const (
	FD          = "FD"
	MEF         = "MEF"
	FDWithOM    = "OM_FD"
	MaxCertSize = 1
)

// edge-om config
const (
	DeviceOmSvcUrl = "/device_om"
	EdgeOmSvcUrl   = "/edge_om"
	EdgeCoreSvcUrl = "/edge_core/"
)

// InnerCertPathName inner_certs
const (
	InnerCertPathName    = "inner_certs"
	GenerateCertWaitTime = 120 * time.Second
	// InnerCert is resource for edge-main to get cert info from edge-om
	InnerCert = "/inner/cert"
)

// CloudCoreCertPathName cloud core cert path
const (
	CloudCoreCertPathName = "cloud_core_certs"
)

// MefCertCommonNamePrefix mef cert common name prefix
const (
	MefCertCommonNamePrefix = "MindXMEF"
)

// SignatureCerts certs for verifying software package integrity
const (
	ModeFileCheckAgl = "sha256"
)

// ImageCertPathName image_cert
const (
	ImageCertFileName = "root.crt"
	ImageCertPathName = "image_certs"
	DockerCertDir     = "/etc/docker/certs.d"
)

// related table configurations
const (
	TimeFormat         = "2006-01-02 15:04:05"
	NetMgrConfigKey    = "netMgrConfig"
	InstallerConfigKey = "installerConfig"
	SoftwareCert       = "softwareCert"
	AlarmCertConfig    = "alarmCertConfig"
	EdgeOmCapabilities = "edgeOmCapabilities"
	Token              = "token"
)

// message flow conversion used const
const (
	Install = "install"
	Upgrade = "upgrade"
)

// module message options
const (
	OptGet       = "GET"
	OptPost      = "POST"
	OptReport    = "REPORT"
	OptQuery     = "query"
	OptInsert    = "insert"
	OptUpdate    = "update"
	OptDelete    = "delete"
	OptPatch     = "patch"
	OptResponse  = "response"
	OptRaw       = "operate"
	OptError     = "error"
	OptRestart   = "restart"
	OptCheck     = "check"
	OptSync      = "sync"
	OptKeepalive = "keepalive"
)

// ResSysSwInfo module system info
const (
	// ResMefPodPrefix cloudcore -> edgecore pod update
	ResMefPodPrefix = "mef-user/pod/"
	// ResMefPodPatchPrefix edgecore -> cloudcore pod status update
	ResMefPodPatchPrefix = "mef-user/podpatch/"
	// ResMefImagePullSecret cloudcore -> edgecore image pull secret update
	ResMefImagePullSecret = "mef-user/secret/image-pull-secret"
	// ResMefNodeLease edgecore -> cloudcore node-lease resource
	ResMefNodeLease = "kube-node-lease/lease/"

	// ResImageCertInfo DeviceOM -> edge-main fd image repository certs
	ResImageCertInfo = "/edge/system/image-cert-info"
	// ActionSecret FD -> edgecore image repository account
	ActionSecret = "websocket/secret/fusion-director-docker-registry-secret"
	// ActionConfigmap FD -> edgecore configmap
	ActionConfigmap = "websocket/configmap/"
	// ActionPod FD -> edgecore pod update
	ActionPod = "websocket/pod/"
	// ActionPodPatch edgecore -> edge-main pod status update (incremental)
	ActionPodPatch = "websocket/podpatch/"
	// ActionPodsData FD -> edge-main delete all pods data
	ActionPodsData = "websocket/pods_data"
	// ActionContainerInfo FD -> edge-main configure container info
	ActionContainerInfo = "websocket/container_info"
	// ActionDefaultNodeStatus edgecore -> edge-main node status
	ActionDefaultNodeStatus = "default/node/"
	// ActionDefaultNodePatch edgecore -> edge-main node status update (incremental)
	ActionDefaultNodePatch = "default/nodepatch/"
	// ActionModelFiles FD -> edge-main model files
	ActionModelFiles    = "websocket/modelfiles"
	ActionModelFileInfo = "websocket/modelfile_info"
	// ResPodStatus edge-main -> FD pod status
	ResPodStatus = "websocket/podstatus"
	// ResNodeStatus edge-main -> FD node status
	ResNodeStatus = "websocket/nodestatus"
	// QueryAllAlarm DeviceOM -> edge-main query all alarm
	QueryAllAlarm = "/edge/system/all-alarm"
	// ResAlarm edge-main -> FD alarm
	ResAlarm = "websocket/alarm"
	// ResConfigResult edge-main -> FD async result
	ResConfigResult = "websocket/config_result"
	// ModifiedNodePrefix edge-main -> FD node status
	ModifiedNodePrefix = "websocket/nodestatus/edge-"
	// ResNpuSharing FD -> edge-main npu sharing
	ResNpuSharing   = "websocket/npu_sharing"
	CenterNpuName   = "huawei.com/Ascend310"
	SharableNpuName = "huawei.com/davinci-mini"
)

// kubeedge resource types
const (
	ResourceTypePod           = "pod"
	ResourceTypeNode          = "node"
	ResourceTypePodPatch      = "podpatch"
	ResourceTypeNodePatch     = "nodepatch"
	ResourceTypeSecret        = "secret"
	ResourceTypeConfigMap     = "configmap"
	ResourceTypePodsData      = "pods_data"
	ResourceTypeAlarm         = "alarm"
	ResourceTypeNpuSharing    = "npu_sharing"
	ResourceTypeContainerInfo = "container_info"
	ResourceTypeModelFile     = "modelfiles"
)

// module message modules
const (
	LogMgrName    = "log-manager"
	InnerClient   = "inner-client"
	ConfigMgr     = "config-manager"
	ModEdgeProxy  = "EdgeProxy"
	ModEdgeOm     = "EdgeOm"
	ModEdgeHub    = "EdgeHub"
	ModEdgeCore   = "EdgeCore"
	ModCloudCore  = "CloudCore"
	ModDeviceOm   = "DeviceOm"
	CfgRestore    = "CfgRestore"
	AlarmManager  = "alarm"
	OmJobManager  = "Om-job-manager"
	OmAlarmMgr    = "Om-alarm-manager"
	MainAlarmMgr  = "Main-alarm-manager"
	ModHandlerMgr = "handler-manager"
	ModEdgeMain   = "EdgeMain"
)

// database paths
const (
	DbEdgeOmPath   = "edge_om.db"
	DbEdgeMainPath = "edge_main.db"
	DbEdgeCorePath = "edgecore.db"
)

// database table key
const (
	// MetaAlarmKey table meta alarm key
	MetaAlarmKey = "alarm"
)

// module message resources
const (
	ResConfig = "/config"
	// ResSoftwareVersion resource software info
	ResSoftwareVersion = "/edge/version-info"
	// ResDownloadCert resource for downloading cert
	ResDownloadCert = "/cert/download_info"
	ResCertUpdate   = "/cert/update"
	ResEdgeCert     = "/cert/edge"
	// DeviceOmConnectMsg resource to inform edgeOM that deviceOM successfully connects
	DeviceOmConnectMsg = "/deviceOm/connect"
	DeleteNodeMsg      = "/edgemanager/delete/node"
	// ReportAlarmMsg resource for report edgeOM alarm
	ReportAlarmMsg = "/report/alarm"
	// ResMefAlarmReport resource for report alarm to mef
	ResMefAlarmReport = "/edge/alarm/report"

	// ResDownloadProgress is resource for edge-main to report progress of software download
	ResDownloadProgress = "/edge/download-progress"
	// ResDumpLogTaskError resource for log-dumping errors
	ResDumpLogTaskError = "/logmgmt/dump/error"
)

// module constants
const (
	ControllerModule     = "controller"
	EdgeControllerModule = "edgecontroller"
	ResourceModule       = "resource"
	DeviceOmModule       = "device-om"
	HardwareModule       = "hardware"
	EdgeManagerModule    = "EdgeManager"
	EdgedModule          = "edged"
	MetaModule           = "meta"
	WebSocketModule      = "websocket"
)

// upgrade manager
const (
	DefaultTryCount       = 10
	TryConnectNet         = 5
	StartWsWaitTime       = 5 * time.Second
	CenterSycMsgWaitTime  = 5 * time.Second
	WsSycMsgWaitTime      = 30 * time.Second
	WsSycMsgRetryInterval = 5 * time.Second
	PreUpgradePath        = "/home/data/mefedge"
	UnpackPath            = "/home/data/mefedge/unpack"
	ShellExt              = ".sh"
	InstallMinDiskSpace   = 520 * MB
	LogMinDiskSpace       = 84 * MB
	LogBackupMinDiskSpace = 108 * MB
	InstallerExtractMin   = 130 * MB
	InstallerUpgradeMin   = 110 * MB
	MaxCrlSizeInMb        = 10

	InstallerTarGzSizeMaxInMB = 120

	TarFlag = "file"
	CmsFlag = "cms"
	CrlFlag = "crl"

	B         = 1
	KB        = 1024 * B
	MB        = 1024 * KB
	Base8     = 8
	Base10    = 10
	BitSize64 = 64
	BitSize0  = 0

	Day = 24 * time.Hour
)

// edge core config parameters
const (
	ConfigDatabase           = "database"
	ConfigDataSource         = "dataSource"
	ConfigModules            = "modules"
	ConfigEdgeHub            = "edgeHub"
	ConfigTlsCaFile          = "tlsCaFile"
	ConfigTlsCertFile        = "tlsCertFile"
	ConfigTlsPrivateKeyFile  = "tlsPrivateKeyFile"
	ConfigHostnameOverride   = "hostnameOverride"
	ConfigNodeLabels         = "nodeLabels"
	ConfigSerialNumber       = "serialNumber"
	ConfigEdged              = "edged"
	ConfigNodeIP             = "nodeIP"
	ConfigTailoredKubelet    = "tailoredKubeletConfig"
	ConfigReadOnlyPort       = "readOnlyPort"
	ConfigRootDirectory      = "rootDirectory"
	ConfigEvictionHard       = "evictionHard"
	SignalNodeFsAvailable    = "nodefs.available"
	SignalNodeFsInodesFree   = "nodefs.inodesFree"
	ConfigServerTLSBootstrap = "serverTLSBootstrap"
	ConfigDeviceTwin         = "deviceTwin"
	ConfigEventBus           = "eventBus"
	ConfigCgroupDriver       = "cgroupDriver"
	NewTlsPrivateKeyFile     = "/run/edgecore.pipe"
	OldTlsPrivateKeyFile     = "/usr/local/mindx/MEFEdge/config/edge_core/inner_certs/client.key.pipe"
)

// edge om config parameters
const (
	ConfigMaxContainerNumber = "maxContainerNumber"
	ConfigHostPath           = "hostPath"
	MaxContainerNumber       = 20
)

// edge websocket config
const (
	ClientIdName    = "client_edge_om"
	ServerIdName    = "server_edge_main"
	LocalIp         = "127.0.0.1"
	InnerServerPort = 10010
	RootCaDir       = "root_certs"
	RootCaName      = "root.crt"
	RootCaKeyName   = "root.key"
	CrlName         = "root.crl"
	ServerCertName  = "server.crt"
	ServerKeyName   = "server.key"
	ClientCertName  = "client.crt"
	ClientKeyName   = "client.key"
	KmcDir          = "kmc"
	KmcCfgName      = "kmc-config.json"
	KmcMasterName   = "master.ks"
	KmcBackupName   = "backup.ks"
	IpPortSeparator = ":"
	IpPortSliceLen  = 2
)

// process lock flag parameters
const (
	FlagPath         = "/run"
	ProcessFlag      = "process_flag"
	ProcessFlagUid   = 0
	ProcessFlagUmask = 077
)

// upgrade setting
const (
	UpgradePath  = "/home/data/mefedge/unpack/edge_installer/software/edge_installer/script/upgrade.sh"
	UpgradeUid   = 0
	UpgradeUmask = 077
	UpgradeMode  = "upgrade"
	EffectMode   = "effect"
	DefaultMode  = "default"
)

// systemctl
const (
	// SystemdServiceDir system service directory
	SystemdServiceDir = "/lib/systemd/system"
	// Systemctl systemctl
	Systemctl = "systemctl"
	// SystemctlEnable command systemctl enable
	SystemctlEnable = "enable"
	// SystemctlDisable command systemctl disable
	SystemctlDisable = "disable"
	// SystemctlIsActive command systemctl is-active
	SystemctlIsActive = "is-active"
	// SystemctlIsEnabled is-enabled
	SystemctlIsEnabled = "is-enabled"
	// SystemctlStart command systemctl start
	SystemctlStart = "start"
	// SystemctlStop command systemctl stop
	SystemctlStop = "stop"
	// SystemctlRestart command systemctl restart
	SystemctlRestart = "restart"
	// SystemctlReload command systemctl daemon-reload
	SystemctlReload = "daemon-reload"
	// SystemctlResetFailed command systemctl reset-failed
	SystemctlResetFailed = "reset-failed"
	// SystemctlStatusActive command systemctl is-active status active
	SystemctlStatusActive = "active"
	// SystemctlEnabled  command systemctl is-active status inactive
	SystemctlEnabled = "enabled"
)

// linux command and path
const (
	Cat                       = "cat"
	Chattr                    = "chattr"
	ProcPath                  = "/proc"
	UuidPath                  = "/proc/sys/kernel/random/uuid"
	IptablesPath              = "/usr/sbin/iptables"
	Iptables                  = "iptables"
	PortLimitIptablesRuleName = "PORT-LIMIT-RULE"
)

// Failed message result
const (
	Failed  = "failed"
	Success = "success"
	OK      = "OK"
	Start   = "start"
	Open    = "open"
	Close   = "close"
)

// Status message
const (
	GroupHub       = "hub"
	SourceHardware = "hardware"
)

// Reset factory files
const (
	ResetService          = "reset_mefedge.service"
	ResetLogFile          = "reset_mefedge.log"
	ResetInstallScript    = "reset_install.sh"
	ResetMiddlewareScript = "reset_middleware.sh"
)

// podConfig
const (
	PodCfgResource   = "pod-config"
	PodCfgFile       = "pod-config.json"
	ContainerCfgFile = "container-config.json"
)

// MEF-Edge edge capability constant
const (
	ResStatic                  = "websocket/sys_info"
	CapabilityNpuSharingConfig = "npu_sharing_config"
	CapabilityNpuSharing       = "npu_sharing"
	CapabilityResourceConfig   = "resource_files_config"
	CapabilityPodConfig        = "pod_config"
	CapabilityPodRestart       = "pod_restart"
	CapabilityPodResource      = "pod_resource"
	CapabilityAppTaskStop      = "container_app_task_stop"
	CapabilityUdpContainerPort = "support_udp_container_port"
)

// constants that uses in prepare_edgecore cmd
const (
	EdgeCorePipePath = "/run/edgecore.pipe"
)

// version info
const (
	Version5Rc1 = "5.0.RC1"
	Version5Rc3 = "5.0.RC3"
	Version5    = "5.0.0"
)

// const for import crl cmd
const (
	CrlPathSubCmd = "crl_path"
	PeerSubCmd    = "peer"
	MefCenterPeer = "mef_center"
)

// compare crls result status while upgrading; crl path on device
const (
	// CompareSame two crls are same
	CompareSame int = 0
	// CompareNew crl to update signed time is newer
	CompareNew int = 1
	// CompareOld crl to update signed time is older
	CompareOld int = 2
	// CrlOnDevicePath path on device
	CrlOnDevicePath = "/etc/hwsipcrl/ascendsip_g2.crl"
)

// constants for model file
const (
	HwHiAiUser          = "HwHiAiUser"
	KubeletRootDir      = "/var/lib/docker/kubelet"
	OldKubeletRootDir   = "/var/lib/kubelet"
	ModelFileRootPath   = "/var/lib/docker"
	ModeFileActiveDir   = "/var/lib/docker/modelfile"
	ModeFileDownloadDir = "/var/lib/docker/model_file_download"
	TargetTypeAll       = "all"
	TargetTypeTemp      = "temp"
)

// const for msg limiter
const (
	MsgRate             = 5
	BurstSize           = 5
	MaxMsgThroughput    = 10 * MB
	MsgThroughputPeriod = 30 * time.Second
)

// const for log recover
const (
	RsyncTimeWaitTime = 30
)

// const for cert alarm config
const (
	CertCheckPeriodDB       = "cert_check_period"
	CertOverdueThresholdDB  = "cert_overdue_threshold"
	DefaultCheckPeriod      = 7
	MinCheckPeriod          = 1
	DefaultOverdueThreshold = 90
	MinOverdueThreshold     = 7
	MaxOverdueThreshold     = 180
)

const (
	// MinArgsLen const for args
	MinArgsLen = 1
	// MaxIterationCount const for loops
	MaxIterationCount = 10000
)

// constants for config result
const (
	ResultProcessing = "processing"
	ResultSuccess    = "success"
	ResultFailed     = "failed"
)
