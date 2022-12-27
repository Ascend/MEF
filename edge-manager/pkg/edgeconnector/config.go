// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the websocket server initial config
package edgeconnector

import "sync"

// SocketConfig the websocket config
type SocketConfig struct {
	CaFile        string
	CertFile      string
	KeyFile       string
	PasswdFile    string
	ServerAddress string
	Port          string
}

var once sync.Once

// Config defines the websocket config
var Config SocketConfig

// InitConfigure initializes the websocket config
func InitConfigure() {
	once.Do(func() {
		Config = SocketConfig{
			CaFile:        "config/certs/rootCA.crt",
			CertFile:      "config/certs/manager.crt",
			KeyFile:       "config/certs/manager.key",
			PasswdFile:    "config/certs/manager.pwd",
			ServerAddress: "127.0.0.1",
			Port:          "10001",
		}
	})
}
