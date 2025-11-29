// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
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
