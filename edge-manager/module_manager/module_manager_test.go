package module_manager

import (
	"testing"
)

type testEnabledModule struct {
}

func (testEnabledModule) Name() string {
	return "TestEnabledModule"
}

func (testEnabledModule) Enable() bool {
	return true
}

func (testEnabledModule) Start() {

}

type testDisabledModule struct {
}

func (testDisabledModule) Name() string {
	return "TestEnabledModule"
}

func (testDisabledModule) Enable() bool {
	return false
}

func (testDisabledModule) Start() {

}

// 使能模块的注册测试
func TestEnableModuleRegistry(t *testing.T) {
	ModuleManagerInit()

	m := testEnabledModule{}

	if err := Registry(m); err != nil {
		t.Errorf("registry test fail")
	}

	// 去注册恢复默认状态
	Unregistry(m)
}

// 去使能模块的注册测试
func TestDisabledModuleRegistry(t *testing.T) {
	ModuleManagerInit()

	m := testDisabledModule{}

	if err := Registry(m); err != nil {
		t.Errorf("registry test fail")
	}

	// 去注册恢复默认状态
	Unregistry(m)
}

// 使用模块的重复注册测试
func TestEnabledModuleRepeatedRegistration(t *testing.T) {
	ModuleManagerInit()

	m := testEnabledModule{}

	if err := Registry(m); err != nil {
		t.Errorf("registry test fail")
	}

	if err := Registry(m); err == nil || err.Error() != "module existed" {
		t.Errorf("repeated registration test fail")
	}

	// 去注册恢复默认状态
	Unregistry(m)
}

// 去使能模块的重复注册测试
func TestDisabledModuleRepeatedRegistration(t *testing.T) {
	ModuleManagerInit()

	m := testDisabledModule{}

	if err := Registry(m); err != nil {
		t.Errorf("registry test fail")
	}

	if err := Registry(m); err == nil || err.Error() != "module existed" {
		t.Errorf("repeated registration test fail")
	}

	// 去注册恢复默认状态
	Unregistry(m)
}
