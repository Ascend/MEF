// Copyright (c) 2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package cloudconnect cache connection status for edge-om
package cloudconnect

var isConnectCloud = false

// SetCloudConnectStatus [method] set cloud connect result
func SetCloudConnectStatus(connectStatus bool) {
	isConnectCloud = connectStatus
}

// GetCloudConnectStatus [method] get cloud connect result
func GetCloudConnectStatus() bool {
	return isConnectCloud
}
