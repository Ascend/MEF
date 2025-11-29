// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package configmanager to init util service
package configmanager

// ImageConfig image config
type ImageConfig struct {
	Domain   string `json:"domain"`
	IP       string `json:"ip"`
	Port     int64  `json:"port"`
	Account  string `json:"account"`
	Password []byte `json:"password"`
}
