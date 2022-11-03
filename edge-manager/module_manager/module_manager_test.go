package module_manager

import (
	"edge-manager/module_manager/model"
	"fmt"
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

type SyncMessageSender struct {
}

func (msg *SyncMessageSender) Name() string {
	return "TestSyncMessageSender"
}

func (msg *SyncMessageSender) Enable() bool {
	return true
}

func (msg *SyncMessageSender) Start() {
	newMsg, err := model.NewMessage()
	if err != nil || newMsg == nil {
		panic("create new msg fail when test send message")
	}
	newMsg.SetRouter("TestSyncMessageSender", "TestMessageReceiver", "update", "app")
	newMsg.FillContent("message content")

	respMsg, err := SendSyncMessage(newMsg, 1*time.Second)
	if err != nil || respMsg == nil {
		panic(fmt.Sprintf("send sync msg fail when test send sync message: %v", err))
	}
	respContent, success := respMsg.GetContent().(string)
	if !success || respContent != "response message content" {
		panic("received response invalid in sync message sender")
	}
}

type MessageReceiver struct {
}

func (msg *MessageReceiver) Name() string {
	return "TestMessageReceiver"
}

func (msg *MessageReceiver) Enable() bool {
	return true
}

func (msg *MessageReceiver) Start() {
	receivedMsg, err := ReceiveMessage("TestMessageReceiver")
	if err != nil || receivedMsg == nil {
		panic("receiver message fail in message receiver")
	}
	content, success := receivedMsg.GetContent().(string)
	if !success || content != "message content" {
		panic("message content is error in message receiver")
	}
	respMsg, err := receivedMsg.NewResponse()
	if err != nil || respMsg == nil {
		panic("create response message fail in message receiver")
	}

	respMsg.FillContent("response message content")
	err = SendMessage(respMsg)
	if err != nil {
		panic("send response message fail in message receiver")
	}
}

func TestSendSyncMessage(t *testing.T) {
	ModuleManagerInit()

	sender := SyncMessageSender{}
	receiver := MessageReceiver{}

	_ = Registry(&sender)
	_ = Registry(&receiver)

	Start()

	// 等待模块协程完成业务
	const waitFinishTime = 2 * time.Second
	time.Sleep(waitFinishTime)

	Unregistry(&sender)
	Unregistry(&receiver)

}
