// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
//go:build !MEFEdge_SDK || MEFEdge_A500

// Package config this file for config model define
package config

// NetManager net manager struct
type NetManager struct {
	NetType string
	WithOm  bool
}
