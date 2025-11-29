// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common this file for base task definitions for install or upgrade
package common

// FuncInfo function info struct for run
type FuncInfo struct {
	Name     string
	Function func() error
}
