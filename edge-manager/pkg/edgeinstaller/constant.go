// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller the constants used in edge-installer module
package edgeinstaller

import "time"

// HttpTimeout timeout in http
const HttpTimeout = 60 * time.Second

// location of software name and version in url
const (
	LocationSfwVersion     = 2
	LocationSfwName        = 0
	LocationSfw            = 0
	LocationRespSfwName    = 1
	LocationRespSfwVersion = 2
)

// software manager info
const (
	SoftwareIP   = "software-manager-service-mindx-edge.svc.cluster.local"
	SoftwarePort = "8102"
	SoftRoute    = "softwaremanager/v1"
	HttpsMethod  = "GET"
)

// set edge account checker
const (
	accountReg         = "^[a-zA-Z0-9-_]{1,256}$"
	passwordMinLen     = 8
	passwordMaxLen     = 256
	DefaultAccountName = "EdgeAccount"
)
