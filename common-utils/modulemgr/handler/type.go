// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handler for
package handler

// RegisterInfo registers info
type RegisterInfo struct {
	// MsgOpt message option
	MsgOpt string
	// MsgOpt message resource
	MsgRes string
	// Handler work handler
	Handler HandleBase
}
