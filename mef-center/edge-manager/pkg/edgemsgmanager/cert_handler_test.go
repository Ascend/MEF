// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgemsgmanager test for getting cert info
package edgemsgmanager

import (
	"encoding/json"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-manager/pkg/config"
	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
)

func TestGetCertInfo(t *testing.T) {
	var p1 = gomonkey.ApplyFuncReturn(config.GetCertCrlPairCache, config.CertCrlPair{}, nil)
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
	err = msg.FillContent(common.ImageCertName)
	convey.So(err, convey.ShouldBeNil)

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testGetCertInfoErrParam() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent([]byte{})
	convey.So(err, convey.ShouldBeNil)

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testGetCertInfoErrCertName() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent("error cert name")
	convey.So(err, convey.ShouldBeNil)

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testGetCertInfoErrGetRootCa() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent(common.SoftwareCertName)
	convey.So(err, convey.ShouldBeNil)

	var p1 = gomonkey.ApplyFuncReturn(config.GetCertCrlPairCache, config.CertCrlPair{}, test.ErrTest)
	defer p1.Reset()

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorQueryCrt)
}

func testGetCertInfoErrMarshal() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent(common.NginxCertName)
	convey.So(err, convey.ShouldBeNil)

	var p1 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return nil, test.ErrTest
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
	err = msg.FillContent(common.SoftwareCertName)
	convey.So(err, convey.ShouldBeNil)

	var p1 = gomonkey.ApplyFunc(modulemgr.SendMessage,
		func(m *model.Message) error {
			return test.ErrTest
		})
	defer p1.Reset()

	resp := GetCertInfo(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorSendMsgToNode)
}
