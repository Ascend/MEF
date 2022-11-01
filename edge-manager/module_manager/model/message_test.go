package model

import (
	"regexp"
	"testing"
)

const msgIdRegex = "[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}"

func TestCreateNewMsg(t *testing.T) {

	msg, err := NewMessage()
	if err != nil {
		t.Errorf("create new message error")
		return
	}

	strReg := regexp.MustCompile(msgIdRegex)
	if !strReg.MatchString(msg.GetId()) {
		t.Errorf("new message id is invalid")
		return
	}

	if msg.GetParentId() != "" {
		t.Errorf("new message parent id is invalid")
		return
	}

	if msg.GetIsSync() {
		t.Errorf("new message sync is invalid")
		return
	}

	msg.SetRouter("src", "dst", "update", "pod")
	if msg.GetSource() != "src" || msg.GetDestination() != "dst" || msg.GetOption() != "update" ||
		msg.GetResource() != "pod" {
		t.Errorf("new message set router fail")
		return
	}

	msg.SetIsSync(true)
	if !msg.GetIsSync() {
		t.Errorf("new message set sync fail")
		return
	}

	msg.FillContent("content")
	if msg.GetContent().(string) != "content" {
		t.Errorf("new message set content fail")
		return
	}
}

func TestCreateResponeMsg(t *testing.T) {
	var newMsg *Message
	var respMsg *Message
	var err error

	if newMsg, err = NewMessage(); err != nil {
		t.Errorf("create new message fail")
		return
	}

	newMsg.SetIsSync(true)
	newMsg.SetRouter("src", "dst", "update", "pod")

	if respMsg, err = newMsg.NewResponse(); err != nil {
		t.Errorf("create respone message fail")
		return
	}

	if respMsg.GetParentId() != newMsg.GetId() {
		t.Errorf("respone message parent id is invalid")
		return
	}

	strReg := regexp.MustCompile(msgIdRegex)
	if !strReg.MatchString(respMsg.GetId()) {
		t.Errorf("new message id is invalid")
		return
	}

	if newMsg.GetIsSync() != respMsg.GetIsSync() || newMsg.GetSource() != respMsg.GetDestination() ||
		newMsg.GetDestination() != respMsg.GetSource() || newMsg.GetResource() != respMsg.GetResource() {
		t.Errorf("respone message header or router is invalid")
		return
	}
}
