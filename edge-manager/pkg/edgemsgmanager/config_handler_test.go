// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager test for getting cloud core token and ca and send to edge
package edgemsgmanager

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/kubeclient"
)

func TestGetConfigInfo(t *testing.T) {
	convey.Convey("test get config info token should be success", t, testGetConfigInfo)
	convey.Convey("test get config info token should be failed, invalid input", t, testGetConfigInfoErrInput)
	convey.Convey("test get config info token should be failed, invalid param", t, testGetConfigInfoErrParam)
	convey.Convey("test get config info token should be failed, invalid config type", t, testGetConfigInfoErrConfigType)
	convey.Convey("test get config info token should be failed, send async msg error", t, testGetConfigInfoErrSendAsyncMsg)
}

func testGetConfigInfo() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent("cloud-core-token")

	outputsGetToken := []gomonkey.OutputCell{
		{Values: gomonkey.Params{[]byte("SwyHTEx_RQppr97g4J5lKXtabJecpejuef8AqKYMAJc"), nil}},
		{Values: gomonkey.Params{[]byte(""), nil}},
	}
	var c *kubeclient.Client
	var p1 = gomonkey.ApplyMethodSeq(c, "GetToken", outputsGetToken)
	defer p1.Reset()

	outputsSendMsg := []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil}, Times: 3},
	}
	var p2 = gomonkey.ApplyFuncSeq(modulemgr.SendAsyncMessage, outputsSendMsg)
	defer p2.Reset()

	resp := GetConfigInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	resp = GetConfigInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetConfigData)

	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{nil, nil}},
		{Values: gomonkey.Params{nil, testErr}},
	}
	var p3 = gomonkey.ApplyMethodSeq(c, "GetCloudCoreCa", outputs)
	defer p3.Reset()

	msg.FillContent("cloud-core-ca")
	resp = GetConfigInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	resp = GetConfigInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetConfigData)
}

func testGetConfigInfoErrInput() {
	resp := GetConfigInfo("error message")
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testGetConfigInfoErrParam() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent([]byte{})
	resp := GetConfigInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testGetConfigInfoErrConfigType() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent("cloud-core-tokenken")

	var c *kubeclient.Client
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetToken",
		func(ki *kubeclient.Client) ([]byte, error) {
			return []byte("SwyHTEx_RQppr97g4J5lKXtabJecpejuef8AqKYMAJc"), nil
		})
	defer p1.Reset()

	var p2 = gomonkey.ApplyFunc(modulemgr.SendAsyncMessage,
		func(m *model.Message) error {
			return nil
		})
	defer p2.Reset()

	resp := GetConfigInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetConfigData)
}

func testGetConfigInfoErrSendAsyncMsg() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent("cloud-core-token")

	var c *kubeclient.Client
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetToken",
		func(ki *kubeclient.Client) ([]byte, error) {
			return []byte("SwyHTEx_RQppr97g4J5lKXtabJecpejuef8AqKYMAJc"), nil
		})
	defer p1.Reset()

	var p2 = gomonkey.ApplyFunc(modulemgr.SendAsyncMessage,
		func(m *model.Message) error {
			return testErr
		})
	defer p2.Reset()

	resp := GetConfigInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorSendMsgToNode)
}
