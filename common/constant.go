// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init const value
package common

import "time"

const (
	// MEF software name
	MEF = "MEF"
	// RestfulServiceName RestfulService
	RestfulServiceName = "RestfulService"
	// NodeManagerName NodeManager
	NodeManagerName = "NodeManager"
	// EdgeConnectorName edge-connector
	EdgeConnectorName = "edge-connector"
	// EdgeInstallerName edge-installer
	EdgeInstallerName = "edge-installer"
	// CertManagerName CertManager
	CertManagerName = "CertManager"
	// SoftwareRepositoryName software repository
	SoftwareRepositoryName = "software repository"

	// Create option create
	Create = "create"
	// Delete option delete
	Delete = "delete"
	// Update option update
	Update = "update"
	// Upgrade option upgrade
	Upgrade = "upgrade"
	// Query option query
	Query = "query"
	// Issue option issue
	Issue = "issue"
	// Get option get
	Get = "get"
	// List option get resource list
	List = "list"

	// Node resource node
	Node = "node"
	// NodeUnManaged resource node unmanaged
	NodeUnManaged = "nodeUnManaged"
	// NodeGroup resource nodeGroup
	NodeGroup = "nodeGroup"
	// NodeStatistics node statistics
	NodeStatistics = "nodeStatistics"
	// Software resource software
	Software = "software"
	// ServiceCert resource service cert
	ServiceCert = "service cert"
	// CSR resource csr
	CSR = "csr"

	// ResponseTimeout Response timeout time
	ResponseTimeout = 3 * time.Second
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
)

const (
	// TimeFormat time format
	TimeFormat = "2006-01-02 15:04:05"
)
