// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr test for npu sharing handler
package handlermgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

var npuSharing = npuSharingHandler{topic: ""}

func TestNpuSharingHandler(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, nil).
		ApplyMethodReturn(&config.CapabilityMgr{}, "Switch", nil)
	defer p.Reset()

	convey.Convey("npu sharing report should be success", t, testNpuSharing)
	convey.Convey("npu sharing report should be failed, msg content is nil", t, testNpuSharingNilContent)
	convey.Convey("npu sharing report should be failed, msg content error", t, testNpuSharingErrContent)
	convey.Convey("test fun sendResponse should be failed, new msg error", t, testSendResponseErrNewMsg)
	convey.Convey("test fun sendResponse should be failed, marshal error", t, testSendResponseErrMarshal)
	convey.Convey("test fun sendResponse should be failed, send msg error", t, testSendResponseErrSendMsg)
}

func testNpuSharing() {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}
	msg.SetRouter(constants.InnerClient, constants.ConfigMgr, constants.OptUpdate, constants.ResNpuSharing)

	contentMap := make(map[string]interface{})
	contentMap["npu_sharing_enabled"] = true
	contentMap["npu_sharing"] = ""
	err = msg.FillContent(contentMap)
	convey.So(err, convey.ShouldBeNil)
	err = npuSharing.Handle(msg)
	convey.So(err, convey.ShouldBeNil)

	var c *config.CapabilityMgr
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "Switch",
		func(c *config.CapabilityMgr, name string, on bool) error {
			return testErr
		})
	defer p1.Reset()

	err = npuSharing.Handle(msg)
	convey.So(err, convey.ShouldBeNil)

	contentMap["npu_sharing_enabled"] = false
	err = msg.FillContent(contentMap)
	convey.So(err, convey.ShouldBeNil)
	err = npuSharing.Handle(msg)
	convey.So(err, convey.ShouldBeNil)
}

func testNpuSharingNilContent() {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}
	err = npuSharing.Handle(msg)
	convey.So(err, convey.ShouldResemble, errors.New("parse content failed"))
}

func testNpuSharingErrContent() {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}
	err = msg.FillContent(model.FormatMsg([]byte("err content")))
	convey.So(err, convey.ShouldBeNil)
	err = npuSharing.Handle(msg)
	convey.So(err, convey.ShouldResemble, errors.New("parse content failed"))

	var testMap map[string]interface{}
	err = msg.FillContent(testMap)
	convey.So(err, convey.ShouldBeNil)
	err = npuSharing.Handle(msg)
	convey.So(err, convey.ShouldResemble, errors.New("key npu_sharing_enabled not exist"))

	contentMap := make(map[string]interface{})
	err = msg.FillContent(contentMap)
	convey.So(err, convey.ShouldBeNil)
	err = npuSharing.Handle(msg)
	convey.So(err, convey.ShouldResemble, errors.New("key npu_sharing_enabled not exist"))

	contentMap["npu_sharing_enabled"] = "err val"
	err = msg.FillContent(contentMap)
	convey.So(err, convey.ShouldBeNil)
	err = npuSharing.Handle(msg)
	convey.So(err, convey.ShouldResemble, errors.New("key npu_sharing_enabled is not bool"))
}

func testSendResponseErrNewMsg() {
	var p1 = gomonkey.ApplyFunc(model.NewMessage,
		func() (*model.Message, error) {
			return nil, testErr
		})
	defer p1.Reset()
	npuSharing.sendResponse(openNpuFailTip)
}

func testSendResponseErrMarshal() {
	var p1 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return nil, testErr
		})
	defer p1.Reset()
	npuSharing.sendResponse(openNpuFailTip)
}

func testSendResponseErrSendMsg() {
	var p1 = gomonkey.ApplyFunc(modulemgr.SendAsyncMessage,
		func(m *model.Message) error {
			return testErr
		})
	defer p1.Reset()
	npuSharing.sendResponse(openNpuFailTip)
}
