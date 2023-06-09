// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init const value
package common

import "time"

const (
	// EdgeMgrDns Edge manager dns name
	EdgeMgrDns = "ascend-edge-manager.mef-center.svc.cluster.local"
	// EdgeMgrPort Edge manager port
	EdgeMgrPort = 8101
	// SoftwareMgrDns software manager dns name
	SoftwareMgrDns = "ascend-software-manager.mef-center.svc.cluster.local"
	// SoftwareMgrPort software manager port
	SoftwareMgrPort = 8102
	// CertMgrDns cert manager dns name
	CertMgrDns = "ascend-cert-manager.mef-center.svc.cluster.local"
	// CertMgrPort cert manager port
	CertMgrPort = 8103
	// NginxMgrDns nginx manager port
	NginxMgrDns = "ascend-nginx-manager.mef-center.svc.cluster.local"
)

const (
	// InnerName inner cert name
	InnerName = "inner_cert"
	// WsSerName websocket server cert name
	WsSerName = "hub_svr"
	// WsCltName websocket client cert name
	WsCltName = "hub_client"
	// SoftwareCertName software manager cert name
	SoftwareCertName = "software"
	// ImageCertName image manager cert name
	ImageCertName = "image"
	// NginxCertName nginx apig cert name
	NginxCertName = "apig"
	// NorthernCertName dir for northbound cert and crl
	NorthernCertName = "north"
)

const (
	// MEF software name
	MEF = "MEF"
	// EdgeCore software edgecore name
	EdgeCore = "edgecore"
	// MEFEdge software mef edge name
	MEFEdge = "MEFEdge"
	// EdgeInstaller software edge-installer name
	EdgeInstaller = "edge-installer"
	// DevicePlugin software device-plugin name
	DevicePlugin = "device-plugin"
	// RestfulServiceName RestfulService
	RestfulServiceName = "RestfulService"
	// NodeManagerName NodeManager
	NodeManagerName = "NodeManager"
	// AppManagerName AppManagerName
	AppManagerName = "AppManager"
	// CloudHubName edge-connector
	CloudHubName = "CloudHub"
	// NodeMsgManagerName node msg manager
	NodeMsgManagerName = "NodeMsgManager"
	// EdgeInstallerName edge-installer
	EdgeInstallerName = "edge-installer"
	// CertManagerName CertManager
	CertManagerName = "CertManager"
	// SoftwareManagerName software manager
	SoftwareManagerName = "software manager"
	// CertManagerService CertManager module name
	CertManagerService = "CertManagerService"
	// ConfigManagerName ConfigManagerName
	ConfigManagerName = "ConfigManager"
	// LogManagerName LogManagerName
	LogManagerName = "LogManager"

	// Create option create
	Create = "create"
	// Delete option delete
	Delete = "delete"
	// Update option update
	Update = "update"
	// Upgrade option upgrade
	Upgrade = "upgrade"
	// Download option download
	Download = "download"
	// Query option query
	Query = "query"
	// Issue option issue
	Issue = "issue"
	// Get option get
	Get = "get"
	// List option get resource list
	List = "list"
	// Add option add
	Add = "add"
	// Inner option for inner message
	Inner = "inner"

	// Node resource node
	Node = "node"
	// AppInstanceByNodeGroup resource app instance by node group
	AppInstanceByNodeGroup = "appInstanceByNodeGroup"
	// NodeGroup resource nodeGroup
	NodeGroup = "nodeGroup"
	// NodeStatus resource node status
	NodeStatus = "nodeStatus"
	// NodeSoftwareInfo resource node software version info
	NodeSoftwareInfo = "nodeSoftwareInfo"
	// CheckResource resources allocatable node resources in node group
	CheckResource = "checkResource"
	// UpdateResource resources allocatable node resources in node group
	UpdateResource = "updateResource"
	// NodeList resource node list
	NodeList = "nodeList"
	// NodeID resource get node id by group id
	NodeID = "nodeID"

	// ResponseTimeout response timeout time
	ResponseTimeout = 3 * time.Second

	// Software resource software
	Software = "software"
	// SoftwareResp resource software response
	SoftwareResp = "software/response"
	// Repository resource
	Repository = "repository"
	// Token resource
	Token = "/edgecore/token"
	// URL link
	URL = "url"
)

const (
	// MaxPort is port max value
	MaxPort = 65535
	// MinPort is port min value
	MinPort = 1025
	// BaseHex  Base Parse integer need params
	BaseHex = 10
	// BitSize64 Base Parse integer need params
	BitSize64 = 64
	// BitSize8 Base Parse integer need params
	BitSize8 = 8
	// DefCmdTimeoutSec represent the default timeout time to exec cmd
	DefCmdTimeoutSec = 120
)

const (
	// DefaultMinPageSize pageSize
	DefaultMinPageSize = 1
	// DefaultMaxPageSize pageSize
	DefaultMaxPageSize = 100
	// DefaultPage 1
	DefaultPage = 1
	// ErrDbUniqueFailed sqlite error UNIQUE constraint failed
	ErrDbUniqueFailed = "UNIQUE constraint failed"
	// TimeFormat time format
	TimeFormat = "2006-01-02 15:04:05"
	// NodeGroupLabelPrefix k8s label prefix for node group
	NodeGroupLabelPrefix = "MEF-Node"
	// DeviceType for Ascend device
	DeviceType = "huawei.com/Ascend310"
)

// regex patterns
const (
	// PaginationNameReg name reg of pagination query
	PaginationNameReg = "^[\\S]{0,253}$"
	// LowercaseCharactersRegex lowercase
	LowercaseCharactersRegex = "[a-z]{1,}"
	// UppercaseCharactersRegex uppercase
	UppercaseCharactersRegex = "[A-Z]{1,}"
	// BaseNumberRegex BaseNumberRegex
	BaseNumberRegex = "[0-9]{1,}"
	// SpecialCharactersRegex special regex
	SpecialCharactersRegex = "[!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~]{1,}"
	// PassWordRegex PassWordRegex
	PassWordRegex          = "^[a-zA-z0-9!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~]{16,64}$"
	Pbkdf2IterationCount   = 10000
	BytesOfEncryptedString = 32
	LockInterval           = 5 * time.Minute
	CheckUnlockInterval    = 15 * time.Second
	OneDay                 = 24 * time.Hour
)

// field constraints
const (
	// MinComplexCount min complex count
	MinComplexCount = 2
)

// field status
const (
	OK   = "OK"
	FAIL = "FAIL"
)

// used to check ip
const (
	ZeroAddr      = "0.0.0.0"
	BroadCastAddr = "255.255.255.255"
)

const (
	// IllegalChars the illegal chars for command
	IllegalChars = "\n!\\; &$<>`"
)

const (
	// OptGet option for get
	OptGet = "GET"
	// OptPost option for post
	OptPost = "POST"
	// OptResp option for response
	OptResp = "response"
	// OptReport option for report
	OptReport = "REPORT"
	// ResEdgeDownloadInfo resource for download software
	ResEdgeDownloadInfo = "/edge/download"
	// ResEdgeUpgradeInfo resource for effect software
	ResEdgeUpgradeInfo = "/edge/upgrade"
	// ResEdgeConfigInfo resource for edge config info
	ResEdgeConfigInfo = "/edge/config"
	// ResDownloadProgress resource progress report
	ResDownloadProgress = "/edge/download-progress"
	// ResSoftwareInfo resource software info
	ResSoftwareInfo = "/edge/version-info"
	// ResDownLoadSoftware resource for downloading software
	ResDownLoadSoftware = "/software/download_info"
	// ResEdgeCoreConfig resource for querying edgecore config
	ResEdgeCoreConfig = "/edgecore/config"
	// ResConfig resource config
	ResConfig = "/config"
	// ResSetEdgeAccount resource for setting edge account
	ResSetEdgeAccount = "/edgemanager/v1/edgeAccount"
	// ResDownLoadCert resource for downloading cert
	ResDownLoadCert = "/cert/download_info"
	// ResLogEdge resource for log of edge node
	ResLogEdge = "/logcollect/log/edge"
	// ResLogTaskProgressEdge resource for progress of edge node log collection
	ResLogTaskProgressEdge = "/logcollect/task/progress/edge"
	// LogCollectPathPrefix prefix for request url of log collection
	LogCollectPathPrefix = "/inner/v1/logcollect"
	// ResRelLogTask resource for collection task
	ResRelLogTask = "/task"
	// ResRelLogTaskProgress resource for progress of collection task
	ResRelLogTaskProgress = "/task/progress"
	// ResRelLogTaskPath resource for output path of collection task
	ResRelLogTaskPath = "/task/path"
	// CertWillOverdue cert will overdue
	CertWillOverdue = "/cert/update"
	// ResEdgeCert resource for issuing cert for a csr from mef edge
	ResEdgeCert = "/cert/edge"

	// EdgeHubName edgehub name
	EdgeHubName = "EdgeHub"
)

// memory unit
const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)

// mode constant
const (
	Mode755 = 0755
	Mode700 = 0700
	Mode600 = 0600
	Mode644 = 0644
	Mode500 = 0500
	Mode444 = 0444
	Mode400 = 0400
)

// CommandCopy constant
const (
	CommandCopy = "cp"
)

// user_mgr constant
const (
	UserGrepCommandPattern = "^%s:"
	GrepCommand            = "grep"
	EtcPasswdFile          = "/etc/passwd"
	NoLoginFlag            = "nologin"
)

// const for unpack zip file
const (
	MaxPkgSizeTimes      = 100
	MaxExtractFileCount  = 100
	MaxSingleExtractSize = 200 * MB
	MaxTotalExtractSize  = 200 * MB

	MefCenterFlag = "mefcenter"
	TarGzSuffix   = ".tar.gz"
	CrlSuffix     = ".tar.gz.crl"
	CmsSuffix     = ".tar.gz.cms"
)

// node specification
const (
	MaxNode         = 1024
	MaxNodeGroup    = 1024
	MaxNodePerGroup = 1024
	MaxGroupPerNode = 10
)

// TmpfsDevNum represents the dev number of tmpfs filesystem in linux stat struct
const TmpfsDevNum = 0x01021994

const (
	// MefUserNs represents the namespace that used by edge-manager to manager applications deployed by customer
	MefUserNs = "mef-user"
)

// http header constant
const (
	ContentType        = "Content-Type"
	ContentDisposition = "Content-Disposition"
	TransferEncoding   = "Transfer-Encoding"
)

// const to check Certs
const (
	SignAlg      = "SHA256-RSA"
	MinPubKeyLen = 3072
)
