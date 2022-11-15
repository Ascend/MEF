// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common to init const value
package common

import "time"

const (
	// RestfulServiceName RestfulService
	RestfulServiceName = "RestfulService"
	// NodeManagerName NodeManager
	NodeManagerName = "NodeManager"
	// EdgeConnectorName edge-connector
	EdgeConnectorName = "edge-connector"
	// EdgeInstallerName edge-installer
	EdgeInstallerName = "edge-installer"
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
	// Get option get
	Get = "get"
	// List option get resource list
	List = "list"

	// Node resource node
	Node = "node"
	// NodeGroup resource nodeGroup
	NodeGroup = "nodeGroup"
	// Software resource software
	Software = "software"

	// ResponseTimeout Response timeout time
	ResponseTimeout = 3 * time.Second
)

const (
	maxPort = 40000
	minPort = 1025
)

const (
	// DefaultMaxPageSize pageSize
	DefaultMaxPageSize = 100
	// DefaultPage 1
	DefaultPage = 1
	// ErrDbUniqueFailed sqlite error UNIQUE constraint failed
	ErrDbUniqueFailed = "UNIQUE constraint failed"
)
