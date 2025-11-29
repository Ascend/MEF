// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package context to start module_manager context
package context

import (
	"time"

	"huawei.com/mindx/common/modulemgr/model"
)

// ModuleMessageContext for message context interface
type ModuleMessageContext interface {
	Registry(moduleName string) error
	Unregistry(moduleName string) error

	Send(msg *model.Message) error
	Receive(moduleName string) (*model.Message, error)
	SendSync(msg *model.Message, duration time.Duration) (*model.Message, error)
	SendResp(msg *model.Message) error
}
