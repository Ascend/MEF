// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node manager const
package nodemanager

const (
	// TimeFormat used for friendly display
	TimeFormat = "2006-01-02 15:04:05"
	// MaxNode MaxNode num 20000
	MaxNode = 20000
	// MaxNodeGroup MaxNodeGroup num 100
	MaxNodeGroup = 100
)

// node status
const (
	statusReady    = "Ready"
	statusOffline  = "Offline"
	statusNotReady = "NotReady"
	statusUnknown  = "Unknown"

	NodeGroupLabel = "ascend-mef-group"
)
