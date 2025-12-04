// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
