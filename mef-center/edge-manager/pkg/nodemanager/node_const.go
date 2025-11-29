// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node manager const
package nodemanager

const (
	// TimeFormat used for friendly display
	TimeFormat         = "2006-01-02 15:04:05"
	masterNodeLabelKey = "node-role.kubernetes.io/master"
	snNodeLabelKey     = "serialNumber"
	managed            = 1
	unmanaged          = 0
)

// node status
const (
	statusReady    = "ready"
	statusOffline  = "offline"
	statusNotReady = "notReady"
	statusUnknown  = "unknown"
	statusAbnormal = "abnormal"
)
