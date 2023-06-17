// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager test for getting cert info
package edgemsgmanager

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/httpsmgr"
)

func TestGetCertInfo(t *testing.T) {
	var c *httpsmgr.ReqCertParams
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetRootCa",
		func(c *httpsmgr.ReqCertParams, certName string) (string, error) {
			return "", nil
		})
	defer p1.Reset()

	var p2 = gomonkey.ApplyFunc(util.GetImageAddress, func() (string, error) {
		return "image.addr", nil
	})
	defer p2.Reset()

	var p3 = gomonkey.ApplyFunc(modulemgr.SendMessage,
		func(m *model.Message) error {
			return nil
		})
	defer p3.Reset()

	convey.Convey("test get cert info should be success", t, testGetCertInfo)
	convey.Convey("test get cert info should be failed, invalid input", t, testGetCertInfoErrInput)
	convey.Convey("test get cert info should be failed, invalid param", t, testGetCertInfoErrParam)
	convey.Convey("test get cert info should be failed, invalid cert name", t, testGetCertInfoErrCertName)
	convey.Convey("test get cert info should be failed, get root ca error", t, testGetCertInfoErrGetRootCa)
	convey.Convey("test get cert info should be failed, marshal error", t, testGetCertInfoErrMarshal)
	convey.Convey("test get cert info should be failed, send msg to edge error", t, testGetCertInfoErrSendMsg)
}

func testGetCertInfo() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent(common.ImageCertName)

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)

	var p2 = gomonkey.ApplyFunc(util.GetImageAddress, func() (string, error) {
		return "", nil
	})
	defer p2.Reset()
	resp = GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorQueryCrt)
}

func testGetCertInfoErrInput() {
	resp := GetCertInfo("error input")
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testGetCertInfoErrParam() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent([]byte{})

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testGetCertInfoErrCertName() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent("error cert name")

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testGetCertInfoErrGetRootCa() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent(common.SoftwareCertName)

	var c *httpsmgr.ReqCertParams
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetRootCa",
		func(c *httpsmgr.ReqCertParams, certName string) (string, error) {
			return "", testErr
		})
	defer p1.Reset()

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorQueryCrt)
}

func testGetCertInfoErrMarshal() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent(common.NginxCertName)

	var p1 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return nil, testErr
		})
	defer p1.Reset()

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testGetCertInfoErrSendMsg() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent(common.SoftwareCertName)

	var p1 = gomonkey.ApplyFunc(modulemgr.SendMessage,
		func(m *model.Message) error {
			return testErr
		})
	defer p1.Reset()

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorSendMsgToNode)
}
