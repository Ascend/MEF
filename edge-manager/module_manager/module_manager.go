package module_manager

import (
	"edge-manager/module_manager/model"
)

var modules map[string]model.Module

func ModuleManagerInit()  {
	modules = make(map[string]model.Module)
}

func Registry(m model.Module)  {
	if m == nil {
		return
	}

	if _, existed := modules[m.Name()]; existed {
		return
	}

	modules[m.Name()] = m
}
