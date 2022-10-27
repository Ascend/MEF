package module_manager

import "testing"

type testModule struct {

}

func (testModule) Name() string {
	return "TestModule"
}

func (testModule) Enable() bool {
	return true
}

func (testModule) Start() {

}

func TestRegistry(t *testing.T) {
	ModuleManagerInit()

	m := testModule{}

	Registry(m)
}
