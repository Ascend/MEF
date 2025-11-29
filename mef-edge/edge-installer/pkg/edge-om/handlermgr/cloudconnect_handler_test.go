// Copyright (c) 2024. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

package handlermgr

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/edge-om/common/certsrequest"
)

var (
	ccHandler       = cloudConnectHandler{}
	cloudConnectMsg model.Message
)

func TestCloudConnectHandler(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil).
		ApplyFunc(certsrequest.RequestCertsFromCenter, func() {})
	defer p.Reset()
	var err error
	cloudConnectMsg, err = newCloudConnectMsg()
	if err != nil {
		fmt.Printf("new cloud connect msg failed, error: %v\n", err)
		return
	}
	convey.Convey("CloudConnectHandler test, successful case", t, testCloudConnectHandler)
	convey.Convey("CloudConnectHandler test, error parameter case", t, testCloudConnectHandlerWithErrorParam)
}

func testCloudConnectHandler() {
	err := ccHandler.Handle(&cloudConnectMsg)
	convey.So(err, convey.ShouldBeNil)
}

func testCloudConnectHandlerWithErrorParam() {
	cloudConnectMsg.FillContent("")
	err := ccHandler.Handle(&cloudConnectMsg)
	convey.So(err, convey.ShouldResemble, errors.New("get connect result failed"))
}

func newCloudConnectMsg() (model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return model.Message{}, errors.New("new message failed")
	}
	msg.Content, err = json.Marshal(true)
	if err != nil {
		fmt.Printf("new message content failed, error: %v\n", err)
		return model.Message{}, errors.New("new message content failed")
	}
	return *msg, nil
}
