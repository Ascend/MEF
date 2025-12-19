// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package common for edgectl
package common

import "time"

// edge control commands
const (
	DomainConfig     = "domainconfig"
	DomainConfigDesc = "to set domain/ip mapping for image registry"
	NetConfig        = "netconfig"
	NetConfigDesc    = "to set MEF Center net configuration"

	AlarmConfig     = "alarmconfig"
	AlarmConfigDesc = "to update alarm used configuration"

	GetAlarmCfg     = "getalarmconfig"
	GetAlarmCfgDesc = "to get alarm used configuration"
)

// netconfig commands config
const (
	StandardInput      = 0
	MinTokenLen        = 16
	MaxTokenLen        = 64
	EnterTokenWaitTime = 1 * time.Minute
)

// update and get alarm config commands config
const (
	CertCheckPeriodCmd      = "cert_period"
	CertOverdueThresholdCmd = "cert_threshold"
	UnitDay                 = "day"
	DiffTime                = 3
)
