// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller the constants used in edge-installer module
package edgeinstaller

import "time"

// WaitSfwSyncTime waiting for a response from the software manager
const WaitSfwSyncTime = 10 * time.Second

// HttpTimeout timeout in http
const HttpTimeout = 60 * time.Second

// location of software name and version in url
const (
	LocationSfwVersion = 2
	LocationSfwName    = 0
)

// software manager info
const (
	SoftwareIP   = "software-manager-service-mindx-edge.svc.cluster.local"
	SoftwarePort = "8102"
	SoftRoute    = "softwaremanager/v1"
	HttpMethod   = "GET"
)
