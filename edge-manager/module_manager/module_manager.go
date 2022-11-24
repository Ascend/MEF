package module_manager

import (
	"fmt"
	"sync"
	"time"

	"edge-manager/module_manager/context"
	"edge-manager/module_manager/model"
)

var enabledModule sync.Map
var disabledModule sync.Map
var moduleContext context.ModuleMessageContext

// ModuleManagerInit module manager init
func ModuleManagerInit() {
	moduleContext = context.GetContent()
}

func registryEnabledModule(m model.Module) error {
	enabledModule.Store(m.Name(), m)
	return moduleContext.Registry(m.Name())
}

func registryDisabledModule(m model.Module) {
	disabledModule.Store(m.Name(), m)
}

func isModuleExised(m model.Module) bool {
	if _, existed := enabledModule.Load(m.Name()); existed {
		return true
	}

	if _, existed := disabledModule.Load(m.Name()); existed {
		return true
	}

	return false
}

// Registry registry new module
func Registry(m model.Module) error {
	if m == nil {
		return fmt.Errorf("input is invalid when registry module")
	}

	if isModuleExised(m) {
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

// Start start the module manager
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
		return fmt.Errorf("input is invalid weh send msg")
	}
	if m.GetParentId() == "" {
		return moduleContext.Send(m)
	}
	return moduleContext.SendResp(m)
}

// SendSyncMessage send sync message
func SendSyncMessage(m *model.Message, dutation time.Duration) (*model.Message, error) {
	return moduleContext.SendSync(m, dutation)
}
