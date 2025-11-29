// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init node manager const
package appmanager

import "time"

const (
	// MaxApp MaxApp num 1000
	MaxApp = 1000
	// DecimalScale for int to string
	DecimalScale = 10
	// AppLabel for label pod
	AppLabel = "v1"
	// AppName for label app pod
	AppName = "appname"
	// AppId for label app pod
	AppId = "appid"

	informerSyncInterval = time.Duration(30) * time.Second
	houseKeepingInterval = time.Duration(60) * time.Second

	nodeStatusReady          = "ready"
	nodeStatusUnknown        = "unknown"
	podStatusUnknown         = "unknown"
	containerStateUnknown    = "unknown"
	containerStateWaiting    = "waiting"
	containerStateRunning    = "running"
	containerStateTerminated = "terminated"
)
