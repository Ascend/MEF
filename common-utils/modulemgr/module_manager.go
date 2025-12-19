// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package modulemgr to start module_manager server
package modulemgr

import (
	"fmt"
	"sync"
	"time"

	"huawei.com/mindx/common/modulemgr/context"
	"huawei.com/mindx/common/modulemgr/model"
)

var enabledModule sync.Map
var disabledModule sync.Map
var moduleContext context.ModuleMessageContext

// ModuleInit module manager init
func ModuleInit() {
	moduleContext = context.GetContent()
}

func registryEnabledModule(m model.Module) error {
	enabledModule.Store(m.Name(), m)
	return moduleContext.Registry(m.Name())
}

func registryDisabledModule(m model.Module) {
	disabledModule.Store(m.Name(), m)
}

func isModuleExisted(m model.Module) bool {
	if _, existed := enabledModule.Load(m.Name()); existed {
		return true
	}

	if _, existed := disabledModule.Load(m.Name()); existed {
		return true
	}

	return false
}

// IsEnabledModule tests whether the module is enabled
func IsEnabledModule(name string) bool {
	_, ok := enabledModule.Load(name)
	return ok
}

// Registry new module
func Registry(m model.Module) error {
	if m == nil {
		return fmt.Errorf("input is invalid when registry module")
	}

	if isModuleExisted(m) {
		return fmt.Errorf("module existed")
	}

	if m.Enable() {
		return registryEnabledModule(m)
	}
	registryDisabledModule(m)
	return nil
}

// Unregistry unregistry module
func Unregistry(m model.Module) {
	if m.Enable() {
		enabledModule.Delete(m.Name())
		_ = moduleContext.Unregistry(m.Name())
	} else {
		disabledModule.Delete(m.Name())
	}
}

// Start the module manager
func Start() {
	enabledModule.Range(func(key, value interface{}) bool {
		module, ok := value.(model.Module)
		if !ok {
			return true
		}
		go module.Start()
		return true
	})
}

// ReceiveMessage receive inner message
func ReceiveMessage(moduleName string) (*model.Message, error) {
	return moduleContext.Receive(moduleName)
}

// SendMessage send message
func SendMessage(m *model.Message) error {
	if m == nil {
		return fmt.Errorf("input is invalid when send msg")
	}
	if m.GetParentId() == "" {
		return SendAsyncMessage(m)
	}
	return moduleContext.SendResp(m)
}

// SendAsyncMessage send async message
func SendAsyncMessage(m *model.Message) error {
	if m == nil {
		return fmt.Errorf("input is invalid when send msg")
	}

	m.SetIsSync(false)
	return moduleContext.Send(m)
}

// SendSyncMessage send sync message
func SendSyncMessage(m *model.Message, duration time.Duration) (*model.Message, error) {
	if m == nil {
		return nil, fmt.Errorf("input msg is invalid")
	}
	m.SetIsSync(true)
	return moduleContext.SendSync(m, duration)
}
