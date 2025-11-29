// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package context to start module_manager context
package context

// GetContent to get message content
func GetContent() ModuleMessageContext {
	return NewChannelContext()
}
