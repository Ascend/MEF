// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgemsgmanager test for dealing edge software download info and send to edge
package edgemsgmanager

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
)

func TestDownloadInfo(t *testing.T) {
	var p1 = gomonkey.ApplyFunc(modulemgr.SendSyncMessage,
		func(m *model.Message, duration time.Duration) (*model.Message, error) {
			rspMsg, err := model.NewMessage()
			if err != nil {
				hwlog.RunLog.Errorf("create message failed, error: %v", err)
				return nil, err
			}
			if err = rspMsg.FillContent(common.OK); err != nil {
				hwlog.RunLog.Errorf("fill content failed: %v", err)
				return nil, err
			}
			return rspMsg, nil
		}).ApplyFuncReturn(utils.IsLocalIp, false)
	defer p1.Reset()

	convey.Convey("test download software should be success", t, testDownloadSfw)
	convey.Convey("test download software should be failed, invalid input", t, testDownloadSfwErrInput)
	convey.Convey("test download software should be failed, invalid param", t, testDownloadSfwErrParam)
	convey.Convey("test download software should be failed, invalid sn", t, testDownloadSfwErrSn)
	convey.Convey("test download software should be failed, invalid softWareName", t, testDownloadSfwErrSfwName)
	convey.Convey("test download software should be failed, invalid Package", t, testDownloadSfwErrPkg)
	convey.Convey("test download software should be failed, invalid SignFile", t, testDownloadSfwErrSignFile)
	convey.Convey("test download software should be failed, invalid CrlFile", t, testDownloadSfwErrCrlFile)
	convey.Convey("test download software should be failed, invalid UserName", t, testDownloadSfwErrUserName)
	convey.Convey("test download software should be failed, invalid Password", t, testDownloadSfwErrPwd)
	convey.Convey("test download software should be failed, new msg error", t, testDownloadErrNewMsg)
	convey.Convey("test download software should be failed, send sync msg error", t, testDownloadErrSendSyncMsg)
}

func createDownloadSfwBaseData() SoftwareDownloadInfo {
	baseContent := `{
    "serialNumbers": ["2102312NSF10K8000130"],
    "softWareName": "MEFEdge",
    "softWarVersion": "1.0",
    "downLoadInfo": {
        "package": "GET https://127.0.0.1/Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
        "signFile": "GET https://127.0.0.1/Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
        "crlFile": "GET https://127.0.0.1/Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
        "userName": "FileTransferAccount",
        "password": [118,103,56,115,42,98,35,118,120,54,111]
    	}
	}`

	var req SoftwareDownloadInfo
	err := json.Unmarshal([]byte(baseContent), &req)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal failed, error: %v", err)
		return req
	}
	return req
}

func createDownloadSfwRightMsg() *model.Message {
	req := createDownloadSfwBaseData()

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return nil
	}
	if err = msg.FillContent(&req); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return nil
	}
	return msg
}

func testDownloadSfw() {
	msg := createDownloadSfwRightMsg()
	if msg == nil {
		return
	}

	resp := downloadSoftware(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDownloadSfwErrInput() {
	resp := downloadSoftware(&model.Message{Content: []byte("")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testDownloadSfwErrParam() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return
	}
	err = msg.FillContent(common.OK)
	convey.So(err, convey.ShouldBeNil)
	resp := downloadSoftware(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testDownloadSfwErrSn() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return
	}

	req := createDownloadSfwBaseData()
	dataCases := [][]string{
		{},
		{"_2102312NSF10K8000130"},
		{"-2102312NSF10K8000130"},
		{"2102312NSF10K8000130_"},
		{"2102312NSF10K8000130-"},
		{"2102312NSF10K8000130", "2102312NSF10K8000130"},
		{"21!02312NSF10K800013$0"},
		{"2102312NSF10K80001302102312NSF10K80001302102312NSF10K800013021023"},
	}
	for _, dataCase := range dataCases {
		req.SerialNumbers = dataCase
		err = msg.FillContent(req)
		convey.So(err, convey.ShouldBeNil)

		resp := downloadSoftware(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadSfwErrSfwName() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return
	}

	req := createDownloadSfwBaseData()
	dataCases := []string{
		"",
		"AtlasEdge",
	}
	for _, dataCase := range dataCases {
		req.SoftwareName = dataCase
		err = msg.FillContent(req)
		convey.So(err, convey.ShouldBeNil)

		resp := downloadSoftware(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadSfwErrPkg() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return
	}

	req := createDownloadSfwBaseData()
	dataCases := []string{
		"",
		" ",
		"GET ",
		"https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A!scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A\nscend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A$scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A\\scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A;scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A<scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A>scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://Ascend -mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A|scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A`scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
	}
	for _, dataCase := range dataCases {
		req.DownloadInfo.Package = dataCase
		err = msg.FillContent(req)
		convey.So(err, convey.ShouldBeNil)

		resp := downloadSoftware(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadSfwErrSignFile() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return
	}

	req := createDownloadSfwBaseData()
	failDataCases := []string{
		"GET ",
		"https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A!scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A\nscend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A$scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A\\scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A;scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A<scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A>scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://Ascend -mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A|scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
		"GET https://A`scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
	}
	for _, dataCase := range failDataCases {
		req.DownloadInfo.SignFile = dataCase
		err = msg.FillContent(req)
		convey.So(err, convey.ShouldBeNil)

		resp := downloadSoftware(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadSfwErrCrlFile() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return
	}

	req := createDownloadSfwBaseData()
	failDataCases := []string{
		"GET ",
		"https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A!scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A\nscend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A$scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A\\scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A;scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A<scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A>scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://Ascend -mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A|scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
		"GET https://A`scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
	}
	for _, dataCase := range failDataCases {
		req.DownloadInfo.CrlFile = dataCase
		err = msg.FillContent(req)
		convey.So(err, convey.ShouldBeNil)
		resp := downloadSoftware(msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	}
}

func testDownloadSfwErrUserName() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return
	}

	req := createDownloadSfwBaseData()
	failDataCases := []string{
		"",
		"_FileTransferAccount",
		"-FileTransferAccount",
		"FileTransferAccountFileTransferAccountFileTransferAccountFileTransf",
	}
	for _, dataCase := range failDataCases {
		req.DownloadInfo.UserName = dataCase
		err = msg.FillContent(&req)
		convey.So(err, convey.ShouldBeNil)

		resp := downloadSoftware(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadSfwErrPwd() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
		return
	}

	req := createDownloadSfwBaseData()
	req.DownloadInfo.Password = nil
	err = msg.FillContent(&req)
	convey.So(err, convey.ShouldBeNil)

	resp := downloadSoftware(msg)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testDownloadErrNewMsg() {
	msg := createDownloadSfwRightMsg()
	if msg == nil {
		return
	}

	var p1 = gomonkey.ApplyFunc(model.NewMessage,
		func() (*model.Message, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()

	resp := downloadSoftware(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorNewMsg)
}

func testDownloadErrSendSyncMsg() {
	msg := createDownloadSfwRightMsg()
	if msg == nil {
		return
	}

	var p = gomonkey.ApplyFuncReturn(modulemgr.SendSyncMessage, nil, test.ErrTest)
	defer p.Reset()
	resp := downloadSoftware(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorSendMsgToNode)
}
