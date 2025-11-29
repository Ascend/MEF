// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build !MEFEdge_SDK || MEFEdge_A500

// Package common for setting nodeIP to edge core configuration file
package common

// SetNodeIPToEdgeCore set nodeIP to edge core configuration file
func SetNodeIPToEdgeCore() error {
	// do not need to set nodeIP
	return nil
}
