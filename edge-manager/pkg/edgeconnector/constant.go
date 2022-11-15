// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the constants used
package edgeconnector

import "time"

// websocket connection
const (
	WriteDeadline         = 15 * time.Second
	ReadDeadline          = 15 * time.Second
	ReadBufferSize        = 1024
	WriteBufferSize       = 1024
	ReadInstallerDeadline = 5 * time.Minute
)

// related table conn_infos
const (
	TimeFormat    = "2006-01-02 15:04:05"
	MinPort       = 0
	MaxPort       = 65535
	MinNameLength = 6
	MaxNameLength = 32
	MinPwdLength  = 6
	MaxPwdLength  = 32
	ZeroAddr      = "0.0.0.0"
	BroadCastAddr = "255.255.255.255"
)
