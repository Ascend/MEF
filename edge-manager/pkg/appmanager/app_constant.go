// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init node manager const
package appmanager

import "time"

const (
	// MaxApp MaxApp num 1000
	MaxApp = 1000
	// AppNodeSelectorKey for select node
	AppNodeSelectorKey = "appmanager"
	// AppNodeSelectorValue for select node
	AppNodeSelectorValue = "test"
	// AppLabel for app label
	AppLabel = "v1"
	// AppName for pod label
	AppName = "appname"
	// AppID for pod label
	AppID = "appid"

	informerSyncInterval = time.Duration(30) * time.Second
)
