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
	// DeviceType for Ascend device
	DeviceType = "huawei.com/davinci-mini"

	informerSyncInterval = time.Duration(30) * time.Second

	portMapMaxCount  = 16
	envMaxCount      = 256
	minContainerPort = 1
	maxContainerPort = 65535
	minHostPort      = 1024
	maxHostPort      = 65535
	minUserId        = 1
	maxUserId        = 65535
	minGroupId       = 1
	maxGroupId       = 65535
	commandMaxCount  = 16
	argsMaxCount     = 16

	podStatusUnknown         = "unknown"
	containerStateUnknown    = "unknown"
	containerStateWaiting    = "waiting"
	containerStateRunning    = "running"
	containerStateTerminated = "terminated"
)
