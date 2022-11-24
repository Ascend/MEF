// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager controller
package appmanager

import "edge-manager/pkg/nodemanager"

// AppInstanceInfo encapsulate app instance information
type AppInstanceInfo struct {
	// AppInfo is app information
	AppInfo AppInfo
	// AppContainer is app container information
	AppContainer AppContainer
	// NodeGroup is node group information of app
	NodeGroup nodemanager.NodeGroup
}
