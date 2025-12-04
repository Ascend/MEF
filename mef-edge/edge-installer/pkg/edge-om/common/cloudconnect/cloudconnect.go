// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
