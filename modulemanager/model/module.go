// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package model to start module_manager model
package model

// Module for module interface function
type Module interface {
	Name() string
	Start()
	Enable() bool
}
