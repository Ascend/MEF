package module_manager

import (
	"edge-manager/module_manager/model"
	"edge-manager/pkg/common"
	"github.com/stretchr/testify/assert"
	"huawei.com/mindx/common/hwlog"
	"testing"
	"time"
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
	err := Registry(m)
	assert.Nil(t, err)

	// 去注册恢复默认状态
	Unregistry(m)
}

// 去使能模块的注册测试
func TestDisabledModuleRegistry(t *testing.T) {
	ModuleManagerInit()

	m := testDisabledModule{}
	err := Registry(m)
	assert.Nil(t, err)

	// 去注册恢复默认状态
	Unregistry(m)
}

// 使用模块的重复注册测试
func TestEnabledModuleRepeatedRegistration(t *testing.T) {
	ModuleManagerInit()

	m := testEnabledModule{}

	err := Registry(m)
	assert.Nil(t, err)

	err = Registry(m)
	assert.NotNil(t, err)
	assert.Error(t, err, "module existed")

	// 去注册恢复默认状态
	Unregistry(m)
}

// 去使能模块的重复注册测试
func TestDisabledModuleRepeatedRegistration(t *testing.T) {
	ModuleManagerInit()

	m := testDisabledModule{}

	err := Registry(m)
	assert.Nil(t, err)

	err = Registry(m)
	assert.NotNil(t, err)
	assert.Error(t, err, "module existed")

	// 去注册恢复默认状态
	Unregistry(m)
}

// 使能模块的重复去注册测试
func TestEnabledModuleRepeatedUnregistration(t *testing.T) {
	ModuleManagerInit()

	m := testEnabledModule{}

	err := Registry(m)
	assert.Nil(t, err)

	// 去注册恢复默认状态
	Unregistry(m)
	Unregistry(m)
}

// 去使能模块的重复去注册测试
func TestDisenabledModuleRepeatedUnregistration(t *testing.T) {
	ModuleManagerInit()

	m := testDisabledModule{}

	err := Registry(m)
	assert.Nil(t, err)

	// 去注册恢复默认状态
	Unregistry(m)
	Unregistry(m)
}

type SyncMessageSender struct {
	TestFramework *testing.T
}

func (msg *SyncMessageSender) Name() string {
	return "TestSyncMessageSender"
}

func (msg *SyncMessageSender) Enable() bool {
	return true
}

func (msg *SyncMessageSender) Start() {
	newMsg, err := model.NewMessage()
	assert.Nil(msg.TestFramework, err)
	assert.NotNil(msg.TestFramework, newMsg)

	newMsg.SetRouter("TestSyncMessageSender", "TestMessageReceiver", "update", "app")
	newMsg.FillContent("message content")

	respMsg, err := SendSyncMessage(newMsg, 1*time.Second)
	assert.Nil(msg.TestFramework, err)
	assert.NotNil(msg.TestFramework, respMsg)

	respContent, success := respMsg.GetContent().(string)
	assert.True(msg.TestFramework, success)
	assert.Equal(msg.TestFramework, respContent, "response message content")
}

type MessageReceiver struct {
	TestFramework *testing.T
}

func (msg *MessageReceiver) Name() string {
	return "TestMessageReceiver"
}

func (msg *MessageReceiver) Enable() bool {
	return true
}

func (msg *MessageReceiver) Start() {
	receivedMsg, err := ReceiveMessage("TestMessageReceiver")
	assert.Nil(msg.TestFramework, err)
	assert.NotNil(msg.TestFramework, receivedMsg)

	content, success := receivedMsg.GetContent().(string)
	assert.True(msg.TestFramework, success)
	assert.Equal(msg.TestFramework, content, "message content")

	respMsg, err := receivedMsg.NewResponse()
	assert.Nil(msg.TestFramework, err)
	assert.NotNil(msg.TestFramework, respMsg)

	respMsg.FillContent("response message content")
	err = SendMessage(respMsg)
	assert.Nil(msg.TestFramework, err)
}

func TestSendSyncMessage(t *testing.T) {
	ModuleManagerInit()

	sender := SyncMessageSender{TestFramework: t}
	receiver := MessageReceiver{TestFramework: t}

	_ = Registry(&sender)
	_ = Registry(&receiver)

	Start()

	// 等待模块协程完成业务
	const waitFinishTime = 2 * time.Second
	time.Sleep(waitFinishTime)

	Unregistry(&sender)
	Unregistry(&receiver)

}

func TestMain(m *testing.M) {
	hwRunLogConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwOpLogConfig := &hwlog.LogConfig{
		OnlyToStdout: true,
	}
	common.InitHwlogger(hwRunLogConfig, hwOpLogConfig)
	m.Run()
}
