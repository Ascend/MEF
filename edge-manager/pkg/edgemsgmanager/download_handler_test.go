// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager ut test
package edgemsgmanager

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func createBaseData() SoftwareDownloadInfo {
	baseContent := `{
    "serialNumbers": ["2102312NSF10K8000130"],
    "softWareName": "MEFEdge",
    "softWarVersion": "1.0",
    "downLoadInfo": {
        "package": "GET https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
        "signFile": "GET https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.cms",
        "crlFile": "GET https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz.crl",
        "userName": "FileTransferAccount",
        "password": [118,103,56,115,42,98,35,118,120,54,111]
    	}
	}`

	var req SoftwareDownloadInfo
	err := json.Unmarshal([]byte(baseContent), &req)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal failed")
		return req
	}
	return req
}
func testDownloadInfo() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	req := createBaseData()
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}

	msg.FillContent(string(content))

	var p2 = gomonkey.ApplyFunc(modulemanager.SendSyncMessage, func(m *model.Message,
		duration time.Duration) (*model.Message, error) {
		rspMsg, err := model.NewMessage()
		if err != nil {
			hwlog.RunLog.Error("create message failed")
		}
		rspMsg.FillContent(common.OK)
		return rspMsg, nil
	})
	defer p2.Reset()

	resp := downloadSoftware(msg)

	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDownloadInfoSerialNumbersInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendSyncMessage, func(m *model.Message,
		duration time.Duration) (*model.Message, error) {
		rspMsg, err := model.NewMessage()
		if err != nil {
			hwlog.RunLog.Error("create message failed")
		}
		rspMsg.FillContent(common.OK)
		return rspMsg, nil
	})
	defer p2.Reset()

	req := createBaseData()

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
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

}

func testDownloadInfoSoftWareNameInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendSyncMessage, func(m *model.Message,
		duration time.Duration) (*model.Message, error) {
		rspMsg, err := model.NewMessage()
		if err != nil {
			hwlog.RunLog.Error("create message failed")
		}
		rspMsg.FillContent(common.OK)
		return rspMsg, nil
	})
	defer p2.Reset()
	req := createBaseData()

	dataCases := []string{
		"",
		"AtlasEdge",
	}
	for _, dataCase := range dataCases {
		req.SoftwareName = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

}

func testDownloadInfoPackageInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendSyncMessage, func(m *model.Message,
		duration time.Duration) (*model.Message, error) {
		rspMsg, err := model.NewMessage()
		if err != nil {
			hwlog.RunLog.Error("create message failed")
		}
		rspMsg.FillContent(common.OK)
		return rspMsg, nil
	})
	defer p2.Reset()

	req := createBaseData()

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
	}
	for _, dataCase := range dataCases {
		req.DownloadInfo.Package = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadInfoSignFileInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendSyncMessage, func(m *model.Message,
		duration time.Duration) (*model.Message, error) {
		rspMsg, err := model.NewMessage()
		if err != nil {
			hwlog.RunLog.Error("create message failed")
		}
		rspMsg.FillContent(common.OK)
		return rspMsg, nil
	})
	defer p2.Reset()
	req := createBaseData()

	failDataCases := []string{
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
	}
	for _, dataCase := range failDataCases {
		req.DownloadInfo.SignFile = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadInfoUserNameInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendSyncMessage, func(m *model.Message,
		duration time.Duration) (*model.Message, error) {
		rspMsg, err := model.NewMessage()
		if err != nil {
			hwlog.RunLog.Error("create message failed")
		}
		rspMsg.FillContent(common.OK)
		return rspMsg, nil
	})
	defer p2.Reset()

	req := createBaseData()
	failDataCases := []string{
		"",
		"_FileTransferAccount",
		"-FileTransferAccount",
		"FileTransferAccountFileTransferAccountFileTransferAccountFileTransf",
	}
	for _, dataCase := range failDataCases {
		req.DownloadInfo.UserName = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		msg.FillContent(string(content))

		resp := downloadSoftware(msg)

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testDownloadInfoPasswdInvalid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
	}

	var p2 = gomonkey.ApplyFunc(modulemanager.SendSyncMessage, func(m *model.Message,
		duration time.Duration) (*model.Message, error) {
		rspMsg, err := model.NewMessage()
		if err != nil {
			hwlog.RunLog.Error("create message failed")
		}
		rspMsg.FillContent(common.OK)
		return rspMsg, nil
	})
	defer p2.Reset()

	req := createBaseData()

	req.DownloadInfo.Password = nil
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}
	msg.FillContent(string(content))

	resp := downloadSoftware(msg)

	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)

}

func TestDownloadInfo(t *testing.T) {
	convey.Convey("test download info", t, func() {
		convey.Convey("test download info serialNumbers", func() {
			convey.Convey("create configmap should success", testDownloadInfo)
			convey.Convey("test invalid serialNumbers", testDownloadInfoSerialNumbersInvalid)
			convey.Convey("test invalid softWareName", testDownloadInfoSoftWareNameInvalid)
			convey.Convey("test invalid Package", testDownloadInfoPackageInvalid)
			convey.Convey("test invalid SignFile", testDownloadInfoSignFileInvalid)
			convey.Convey("test invalid UserName", testDownloadInfoUserNameInvalid)
			convey.Convey("test invalid Password", testDownloadInfoPasswdInvalid)
		})
	})
}
