// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init const value
package common

import "time"

const (
	// EdgeMgrDns Edge manager dns name
	EdgeMgrDns = "ascend-edge-manager.mef-center.svc.cluster.local"
	// EdgeMgrPort Edge manager port
	EdgeMgrPort = 8101
	// CertMgrDns cert manager dns name
	CertMgrDns = "ascend-cert-manager.mef-center.svc.cluster.local"
	// CertMgrPort cert manager port
	CertMgrPort = 8103
	// NginxMgrDns nginx manager dns name
	NginxMgrDns = "ascend-nginx-manager.mef-center.svc.cluster.local"
	// NginxMgrPort nginx manager inner RESTful server port
	NginxMgrPort = 8104
	// AlarmMgrDns alarm manager dns name
	AlarmMgrDns = "ascend-alarm-manager.mef-center.svc.cluster.local"
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
	// IcsCertName for icsmanager cert name dir
	IcsCertName = "ics"
)

// consts for ModuleName
const (
	MEFEdge                = "MEFEdge"
	RestfulServiceName     = "RestfulService"
	NodeManagerName        = "NodeManager"
	AppManagerName         = "AppManager"
	CloudHubName           = "CloudHub"
	InnerServerName        = "InnerServer"
	AlarmManagerClientName = "AlarmManagerClient"
	NodeMsgManagerName     = "NodeMsgManager"
	CertManagerName        = "CertManager"
	ConfigManagerName      = "ConfigManager"
	AlarmManagerName       = "AlarmManager"
	CertUpdaterName        = "CertUpdater"
)

const (
	// Create option create
	Create = "create"
	// Delete option delete
	Delete = "delete"
	// Update option update
	Update = "update"
	// Get option get
	Get = "get"
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
	// GetIpBySn is the route for inner msg that sent to nodemanager to get the node info by sn
	GetIpBySn = "/inner/v1/getIpBySn"

	// ResponseTimeout response timeout time
	ResponseTimeout = 30 * time.Second
	// RestfulTimeout restful timeout time
	RestfulTimeout = 6 * time.Minute
	// EdgeManagerRestfulWriteTimeout edge-manager restful write timeout time
	EdgeManagerRestfulWriteTimeout = 2 * time.Hour
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
	// ProgressMax max progress
	ProgressMax = 100
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
	// PassWordRegex PassWordRegex
	PassWordRegex          = "^[a-zA-z0-9!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~]{16,64}$"
	Pbkdf2IterationCount   = 10000
	BytesOfEncryptedString = 32
	LockInterval           = 5 * time.Minute
	CheckUnlockInterval    = 15 * time.Second
	OneDay                 = 24 * time.Hour
)

// field status
const (
	OK   = "OK"
	FAIL = "FAIL"
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
	// ResDownloadProgress resource progress report
	ResDownloadProgress = "/edge/download-progress"
	// ResSoftwareInfo resource software info
	ResSoftwareInfo = "/edge/version-info"
	// ResConfig resource config
	ResConfig = "/config"
	// ResDownLoadCert resource for downloading cert
	ResDownLoadCert = "/cert/download_info"
	// CertWillOverdue cert will overdue
	CertWillOverdue = "/cert/update"
	// CertWillExpired cert will expire
	CertWillExpired = "/cert/update"
	// ResEdgeCert resource for issuing cert for a csr from mef edge
	ResEdgeCert = "/cert/edge"
	// DeleteNodeMsg when delete node send msg to edgehub to stop connection
	DeleteNodeMsg = "/edgemanager/delete/node"
	// EdgeHubName edgehub name
	EdgeHubName = "EdgeHub"
	// CloudCoreName cloud core name
	CloudCoreName = "CloudCore"
	// ResNodeChanged when edge node added or deleted, nodemanager send notify to certupdater
	ResNodeChanged = "/nodemanager/node/changed"
	// ResCertUpdate cert update notify from cert-manager to certupdater, both ca and svc.
	ResCertUpdate = "/inner/v1/cert/update"
	// ResEdgeMgrCertUpdate cert update notify in nginx-manager
	ResEdgeMgrCertUpdate = "/inner/cert/edge-manger"
)

// memory unit
const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
)

// mode constant
const (
	Mode755  = 0755
	Mode700  = 0700
	Mode600  = 0600
	Mode644  = 0644
	Mode500  = 0500
	Mode400  = 0400
	Umask077 = 0077
)

// CommandCopy constant
const (
	CommandCopy = "cp"
)

// TarGzSuffix is the suffix of tar.gz file
const (
	TarGzSuffix = ".tar.gz"
)

// node specification
const (
	MaxNode         = 1024
	MaxNodeGroup    = 1024
	MaxNodePerGroup = 1024
	MaxGroupPerNode = 10
)

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

// constants fo database backup
const (
	BackupDbSuffix = ".backup"
	DbTestInterval = time.Minute
)

// consts for inner websocket Port
const (
	EdgeManagerInnerWsPort = 20000
)

// const for alarm config
const (
	AlarmConfigDBName      = "alarm-manager.db"
	CertCheckPeriodDB      = "cert_check_period"
	CertOverdueThresholdDB = "cert_overdue_threshold"
)

// MefCertCommonNamePrefix mef cert common name prefix
const (
	MefCertCommonNamePrefix = "MindXMEF"
)
