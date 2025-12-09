// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
