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

	WriteRetryCount     = 5
	WriteCenterDeadline = 60 * time.Second

	TimeWaitServiceCertTime = 15 * time.Second
)

// related table conn_infos
const (
	TimeFormat    = "2006-01-02 15:04:05"
	MinNameLength = 6
	MaxNameLength = 32
	MinPwdLength  = 6
	MaxPwdLength  = 45
)

// Online indicates edge-installer is online, Offline indicates edge-installer is offline
const (
	Online  = true
	Offline = false
)

// software manager info
const (
	LocationMethod = 0
	LocationUrl    = 1
	LocationIP     = 0
	LocationPort   = 1
	LocationIpPort = 2
)
