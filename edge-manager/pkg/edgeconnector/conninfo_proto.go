// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the connection info struct in order to operate table conn_infos
package edgeconnector

type baseInfo struct {
	Address  string `json:"address,omitempty"`
	Port     string `json:"port,omitempty"`
	Username string `json:"username"`
	Password []byte `json:"password"`
}

// UpdateConnInfo struct for updating an item in table conn_infos
type UpdateConnInfo struct {
	baseInfo
}
