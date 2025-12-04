// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package model to start module_manager model
package model

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const msgIdRegex = "[0-9a-f]{8}(-[0-9a-f]{4}){3}-[0-9a-f]{12}"

type TestStruct struct {
	TagA string          `json:"A"`
	TagB int             `json:"B"`
	TagC TestInnerStruct `json:"C"`
}

type TestInnerStruct struct {
	TagD bool `json:"D"`
}

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

	if err = msg.FillContent("content"); err != nil {
		t.Errorf("fill content into message fail: %v", err)
		return
	}
	var content string
	if err = msg.ParseContent(&content); err != nil {
		t.Errorf("parse msg content failed: %v", err)
		return
	}

	if content != "content" {
		t.Errorf("parsed content is incorrect: %s", content)
	}
}

func TestCreateResponeMsg(t *testing.T) {
	var newMsg *Message
	var respMsg *Message
	var err error

	if newMsg, err = NewMessage(); err != nil || newMsg == nil {
		t.Errorf("create new message fail")
		return
	}

	newMsg.SetIsSync(true)
	newMsg.SetRouter("src", "dst", "update", "pod")

	if respMsg, err = newMsg.NewResponse(); err != nil || respMsg == nil {
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

func TestMsgMarshal(t *testing.T) {
	convey.Convey("test msg marshaling", t, func() {
		convey.Convey("test fill string content", testMarshalStringContent)
		convey.Convey("test fill bool content", testMarshalBoolContent)
		convey.Convey("test fill byte slice content", testMarshalByteSliceContent)
		convey.Convey("test fill struct content", testMarshalStruct)
		convey.Convey("test fill struct content as json-formatted string", testMarshalStructToString)
	})
}

func testOneMsg(content interface{}, jsonStr string, unmarshalledByte RawMessage, transferToString bool) Message {
	var receivedMsg Message

	msg := Message{}
	err := msg.FillContent(content, transferToString)
	convey.So(err, convey.ShouldBeNil)
	marshaledMsg, err := json.Marshal(msg)
	convey.So(err, convey.ShouldBeNil)
	convey.So(string(marshaledMsg), convey.ShouldContainSubstring, jsonStr)
	err = json.Unmarshal(marshaledMsg, &receivedMsg)
	convey.So(err, convey.ShouldBeNil)
	convey.So(receivedMsg.Content, convey.ShouldResemble, unmarshalledByte)
	return receivedMsg
}

func testOneStringMsg(content, unmarshalledContent interface{}, jsonStr string, unmarshalledByte RawMessage) {
	var receivedContent string
	receivedMsg := testOneMsg(content, jsonStr, unmarshalledByte, false)
	err := receivedMsg.ParseContent(&receivedContent)
	convey.So(err, convey.ShouldBeNil)
	convey.So(receivedContent, convey.ShouldResemble, unmarshalledContent)
}

func testOneBoolMsg(content, unmarshalledContent interface{}, jsonStr string, unmarshalledByte RawMessage) {
	var receivedContent bool
	receivedMsg := testOneMsg(content, jsonStr, unmarshalledByte, false)
	err := receivedMsg.ParseContent(&receivedContent)
	convey.So(err, convey.ShouldBeNil)
	convey.So(receivedContent, convey.ShouldResemble, unmarshalledContent)
}

func testOneByteSliceMsg(content, unmarshalledContent interface{}, jsonStr string, unmarshalledByte RawMessage) {
	var receivedContent []byte
	receivedMsg := testOneMsg(content, jsonStr, unmarshalledByte, false)
	err := receivedMsg.ParseContent(&receivedContent)
	convey.So(err, convey.ShouldBeNil)
	convey.So(receivedContent, convey.ShouldResemble, unmarshalledContent)
}

func testOneStructMsg(content, unmarshalledContent interface{}, jsonStr string, unmarshalledByte RawMessage) {
	var receivedContent TestStruct
	receivedMsg := testOneMsg(content, jsonStr, unmarshalledByte, false)
	err := receivedMsg.ParseContent(&receivedContent)
	convey.So(err, convey.ShouldBeNil)
	convey.So(receivedContent, convey.ShouldResemble, unmarshalledContent)
}

func testOneStructToStringMsg(content, unmarshalledContent interface{}, jsonStr string, unmarshalledByte RawMessage) {
	var receivedContent TestStruct
	receivedMsg := testOneMsg(content, jsonStr, unmarshalledByte, true)
	err := receivedMsg.ParseContent(&receivedContent)
	convey.So(err, convey.ShouldBeNil)
	convey.So(receivedContent, convey.ShouldResemble, unmarshalledContent)
}

func testMarshalStringContent() {
	testOneStringMsg("test", "test", `"content":"test"`, RawMessage(`"test"`))
	testOneStringMsg("\n", "\n", `"content":"\n"`, RawMessage(`"\n"`))
	testOneStringMsg(`\n`, `\n`, `"content":"\\n"`, RawMessage(`"\\n"`))
	testOneStringMsg(`"test"`, "test", `"content":"test"`, RawMessage(`"test"`))
	testOneStringMsg(`"te"st`, `"te"st`, `"content":"\"te\"st"`, RawMessage(`"\"te\"st"`))
}

func testMarshalBoolContent() {
	testOneBoolMsg(true, true, `"content":true`, RawMessage(`true`))
	testOneBoolMsg(false, false, `"content":false`, RawMessage(`false`))
}

func testMarshalByteSliceContent() {
	testOneByteSliceMsg([]byte("test"), []byte{116, 101, 115, 116}, `"content":"test"`, RawMessage(`"test"`))
	testOneByteSliceMsg([]byte("\n"), []byte{10}, `"content":"\n"`, RawMessage(`"\n"`))
	testOneByteSliceMsg([]byte(`\n`), []byte{92, 110}, `"content":"\\n"`, RawMessage(`"\\n"`))
	testOneByteSliceMsg([]byte(`"test"`), []byte{116, 101, 115, 116}, `"content":"test"`, RawMessage(`"test"`))
	testOneByteSliceMsg([]byte(`"te"st`), []byte{34, 116, 101, 34, 115, 116},
		`"content":"\"te\"st"`, RawMessage(`"\"te\"st"`))
}

func testMarshalStruct() {
	testStruct := TestStruct{
		TagA: "test",
		TagB: 0,
		TagC: TestInnerStruct{
			TagD: true,
		},
	}
	testOneStructMsg(testStruct, testStruct, `"content":{"A":"test","B":0,"C":{"D":true}}`,
		RawMessage(`{"A":"test","B":0,"C":{"D":true}}`))
}

func testMarshalStructToString() {
	testStruct := TestStruct{
		TagA: "test",
		TagB: 0,
		TagC: TestInnerStruct{
			TagD: true,
		},
	}
	testOneStructToStringMsg(testStruct, testStruct, `"content":"{\"A\":\"test\",\"B\":0,\"C\":{\"D\":true}}"`,
		RawMessage(`"{\"A\":\"test\",\"B\":0,\"C\":{\"D\":true}}"`))
}
