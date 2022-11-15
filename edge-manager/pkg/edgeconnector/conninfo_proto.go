// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeconnector the connection info struct in order to operate table conn_infos
package edgeconnector

type baseInfo struct {
	Address  string
	Port     string
	UserName string
	Password []byte
}

// UpdateConnInfo struct for updating an item in table conn_infos
type UpdateConnInfo struct {
	baseInfo
}
