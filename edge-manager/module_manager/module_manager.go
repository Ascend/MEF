package module_manager

import (
	"edge-manager/module_manager/context"
	"edge-manager/module_manager/model"
	"fmt"
	"time"
)

var enabledModule map[string]model.Module
var disabledModule map[string]model.Module
var moduleContext context.ModuleMessageContext

func ModuleManagerInit()  {
	enabledModule = make(map[string]model.Module)
	disabledModule = make(map[string]model.Module)
	moduleContext = context.GetContent()
}

func registryEnabledModule(m model.Module) error {
	enabledModule[m.Name()] = m

	return moduleContext.Registry(m.Name())
}

func registryDisabledModule(m model.Module) {
	disabledModule[m.Name()] = m
}

func isModuleExised(m model.Module) bool {
	if _, existed := enabledModule[m.Name()]; existed {
		return true
	}

	if _, existed := disabledModule[m.Name()]; existed {
		return true
	}

	return false
}

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

func Unregistry(m model.Module)  {
	if m.Enable() {
		delete(enabledModule, m.Name())
		_ = moduleContext.Unregistry(m.Name())
	} else {
		delete(disabledModule, m.Name())
	}
}

func ReceiveMessage(moduleName string) (*model.Message, error) {
	return moduleContext.Receive(moduleName)
}

func SendMessage(m *model.Message) error {
	if m == nil {
		return fmt.Errorf("input is invalid weh send msg")
	}
	if m.GetParentId() == "" {
		return moduleContext.Send(m)
	}
	return moduleContext.SendResp(m)
}

func SendSyncMessage(m *model.Message, dutation time.Duration) (*model.Message, error) {
	return moduleContext.SendSync(m, dutation)
}
