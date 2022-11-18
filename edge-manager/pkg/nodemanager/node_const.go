// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node manager const
package nodemanager

const (
	// TimeFormat used for friendly display
	TimeFormat      = "2006-01-02 15:04:05"
	maxNode         = 1024
	maxNodeGroup    = 1024
	maxNodePerGroup = 1024
	managed         = 1
	unmanaged       = 0
)

// node status
const (
	statusReady    = "Ready"
	statusOffline  = "Offline"
	statusNotReady = "NotReady"
	statusUnknown  = "Unknown"

	NodeGroupLabel = "ascend-mef-group"
)
