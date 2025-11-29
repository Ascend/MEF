// Copyright (c) 2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package handlermgr test for get cert handler
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
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

var (
	getCertMsg *model.Message
	getCert    = getCertHandler{}
)

func TestGetCertHandler(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil).
		ApplyFuncReturn(path.GetCompSpecificDir, "", nil).
		ApplyFuncReturn(certutils.GetCertContentWithBackup, nil, nil)
	defer p.Reset()

	var err error
	getCertMsg, err = newGetCertMsg()
	if err != nil {
		fmt.Printf("new get pod config msg failed, error: %v\n", err)
		return
	}
	convey.Convey("test get cert handler should be success", t, testGetCertHandler)
	convey.Convey("test get cert handler should be failed, param convert error", t, testGetCertHandlerErrParam)
	convey.Convey("test get cert handler should be failed, cert name error", t, testGetCertHandlerErrCertName)
	convey.Convey("test get cert handler should be failed, new resp error", t, testGetCertHandlerErrNewResponse)
	convey.Convey("test get cert handler should be failed, marshal error", t, testGetCertHandlerErrMarshal)
	convey.Convey("test get cert handler should be failed, send msg error", t, testGetCertHandlerErrSendMsg)
	convey.Convey("test get cert handler should be failed, get cert path error", t, testGetCertHandlerErrGetCertPath)
	convey.Convey("test get cert handler should be failed, get cert content error", t, testGetCertHandlerErrGetCertContent)
}

func testGetCertHandler() {
	err := getCert.Handle(getCertMsg)
	convey.So(err, convey.ShouldBeNil)
}

func testGetCertHandlerErrParam() {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}
	err = msg.FillContent("error content")
	convey.So(err, convey.ShouldBeNil)
	err = getCert.Handle(msg)
	convey.So(err, convey.ShouldResemble, errors.New("parse cert info para failed"))
}

func testGetCertHandlerErrCertName() {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return
	}
	req := &config.CertReq{
		CertName: "error cert name",
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("marshal failed, error: %v\n", err)
		return
	}
	err = msg.FillContent(reqBytes)
	convey.So(err, convey.ShouldBeNil)
	err = getCert.Handle(msg)
	convey.So(err, convey.ShouldResemble, errors.New("invalid cert name"))
}

func testGetCertHandlerErrNewResponse() {
	var c *model.Message
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "NewResponse",
		func(msg *model.Message) (*model.Message, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := getCert.Handle(getCertMsg)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testGetCertHandlerErrMarshal() {
	var p1 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := getCert.Handle(getCertMsg)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testGetCertHandlerErrSendMsg() {
	var p1 = gomonkey.ApplyFunc(modulemgr.SendMessage,
		func(m *model.Message) error {
			return testErr
		})
	defer p1.Reset()
	err := getCert.Handle(getCertMsg)
	convey.So(err, convey.ShouldResemble, testErr)
}

func testGetCertHandlerErrGetCertPath() {
	var p1 = gomonkey.ApplyFuncReturn(path.GetCompSpecificDir, "", testErr)
	defer p1.Reset()
	err := getCert.Handle(getCertMsg)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get software root cert failed, error: %v", testErr))
}

func testGetCertHandlerErrGetCertContent() {
	var p1 = gomonkey.ApplyFunc(certutils.GetCertContentWithBackup,
		func(path string) ([]byte, error) {
			return nil, testErr
		})
	defer p1.Reset()
	err := getCert.Handle(getCertMsg)
	convey.So(err, convey.ShouldResemble, fmt.Errorf("get software cert failed, error: %v", testErr))
}

func newGetCertMsg() (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		fmt.Printf("new message failed, error: %v\n", err)
		return nil, errors.New("new message failed")
	}
	msg.SetRouter(constants.InnerClient, constants.ConfigMgr, constants.OptGet, constants.InnerCert)
	req := &config.CertReq{
		CertName: constants.SoftwareCertName,
	}
	reqBytes, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("marshal failed, error: %v\n", err)
		return nil, errors.New("marshal failed")
	}
	err = msg.FillContent(reqBytes, true)
	if err != nil {
		fmt.Printf("fill content failed: %v", err)
		return nil, errors.New("fill content failed")
	}
	return msg, nil
}
