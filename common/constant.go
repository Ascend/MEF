// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init const value
package common

import "time"

const (
	// MEF software name
	MEF = "MEF"
	// EdgeCore software edgecore name
	EdgeCore = "edgecore"
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
	// EdgeConnectorName edge-connector
	EdgeConnectorName = "edge-connector"
	// EdgeInstallerName edge-installer
	EdgeInstallerName = "edge-installer"
	// CertManagerName CertManager
	CertManagerName = "CertManager"
	// SoftwareManagerName software manager
	SoftwareManagerName = "software manager"
	// CertManagerService CertManager module name
	CertManagerService = "CertManagerService"

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
	// Deploy option deploy application
	Deploy = "deploy"
	// Undeploy option undeploy application
	Undeploy = "undeploy"
	// Add option add
	Add = "add"
	// Inner option for inner message
	Inner = "inner"

	// Node resource node
	Node = "node"
	// NodeUnManaged resource node unmanaged
	NodeUnManaged = "nodeUnManaged"
	// App resource app
	App = "app"
	// AppInstance resource app instance
	AppInstance = "appInstance"
	// AppInstanceByNode resource app instance by node
	AppInstanceByNode = "appInstanceByNode"
	// AppInstanceByNodeGroup resource app instance by node group
	AppInstanceByNodeGroup = "appInstanceByNodeGroup"
	// NodeGroup resource nodeGroup
	NodeGroup = "nodeGroup"
	// NodeStatus resource node status
	NodeStatus = "nodeStatus"
	// NodeStatistics node statistics
	NodeStatistics = "nodeStatistics"
	// NodeGroupStatistics node group statistics
	NodeGroupStatistics = "nodeGroupStatistics"
	// NodeRelation node relation
	NodeRelation = "nodeRelation"
	// ServiceCert resource service cert
	ServiceCert = "service cert"
	// CSR resource csr
	CSR = "csr"
	// AppTemplate resource app template
	AppTemplate = "AppTemplate"
	// ResponseTimeout Response timeout time
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

// result from edge-installer when downloading and upgrading software
const (
	SuccessResult = "success"
	FailResult    = "fail"
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
)

const (
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
	DeviceType = "huawei.com/davinci-mini"
)

// regex patterns
const (
	// RegAppTemplate regex pattern of app template version name
	RegAppTemplate = `^[a-zA-Z]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`
	// RegContainerName regex pattern of container name
	RegContainerName = `^[a-zA-Z]([_a-zA-Z0-9]{0,30}[a-zA-Z0-9])?$`
	// RegImageName regex pattern of image name
	RegImageName = `^[a-zA-Z]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`
	// RegImageVersion regex pattern of image version
	RegImageVersion = `^[a-zA-Z]([-_a-zA-Z0-9]{0,14}[a-zA-Z0-9])?$`
	// RegEnvKey regex pattern of environment variable key
	RegEnvKey = `^[a-zA-Z]([_a-zA-Z0-9]{0,2046}[a-zA-Z0-9])?$`
)

// protocol
const (
	Tcp = "TCP"
	Udp = "UDP"
)

// field constraints
const (
	// AppTemplateContainersMin app template containers min count
	AppTemplateContainersMin = 1
	// AppTemplateContainersMax app template containers max count
	AppTemplateContainersMax = 10
	// AppTemplateDesMin app template group description min length
	AppTemplateDesMin = 0
	// AppTemplateDesMax app template group description max length
	AppTemplateDesMax = 255
	// CpuMin container CPU min value
	CpuMin = 0.01
	// CpuMax container CPU max value
	CpuMax = 1000
	// CpuDecimalsNum CPU number of decimal places
	CpuDecimalsNum = 2
	// MemoryMin container memory min value
	MemoryMin = 4
	// MemoryMax container memory max value
	MemoryMax = 1024000
	// NpuMin NPU min value
	NpuMin = 0.01
	// NpuMax NPU max value
	NpuMax = 32
	// NpuDecimalsNum NPU number of decimal places
	NpuDecimalsNum = 2

	// ContainerUserIdMin container min user id
	ContainerUserIdMin = 1
	// ContainerUserIdMax container max user id
	ContainerUserIdMax = 65535
	// ContainerGroupIdMin container min group id
	ContainerGroupIdMin = 1
	// ContainerGroupIdMax container max group id
	ContainerGroupIdMax = 65535
	// ContainerPortMin container port min value
	ContainerPortMin = 1
	// ContainerPortMax container port max value
	ContainerPortMax = 65535
	// HostPortMin host port min value
	HostPortMin = 1
	// HostPortMax host port max value
	HostPortMax = 65535
	// TemplateEnvValueMin environment variable value min length
	TemplateEnvValueMin = 1
	// TemplateEnvValueMax environment variable value max length
	TemplateEnvValueMax = 2048
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
	OptGet = "get"
	// OptPost option for post
	OptPost = "post"
	// OptReport option for report
	OptReport = "report"
	// ResProgressReport resource progress report
	ResProgressReport = "/progress"
	// ResDownLoadSoftware resource for downloading software
	ResDownLoadSoftware = "/software/download_info"
	// ResEdgeCoreConfig resource for querying edgecore config
	ResEdgeCoreConfig = "/edgecore/config"
)

// ServerCertPathName websocket server cert path name
const (
	ServerCertPathName = "mef_certs"
)

// memory unit
const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)
