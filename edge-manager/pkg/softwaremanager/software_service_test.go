// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package softwaremanager module test
package softwaremanager

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
)

func createSftAuthBaseData() SftAuthInfo {
	baseContent := `{
		"userName": "FileTransferAccount",
		"password": [118,103,56,115,42,98,35,118,120,54,111]
	}`

	var req SftAuthInfo
	err := json.Unmarshal([]byte(baseContent), &req)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal failed")
		return req
	}
	return req
}

func testAuthInfoShouldSuccess() {
	req := createSftAuthBaseData()
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}

	resp := updateAuthInfo(string(content))

	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testAuthUserNameShouldFailed() {
	req := createSftAuthBaseData()

	failDataCases := []string{
		"",
		"_FileTransferAccount",
		"-FileTransferAccount",
		"FileTransferAccountFileTransferAccountFileTransferAccountFileTransf",
	}
	for _, dataCase := range failDataCases {
		req.UserName = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}
		resp := updateAuthInfo(string(content))

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testAuthPasswordShouldFailed() {
	req := createSftAuthBaseData()

	req.Password = nil
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}

	resp := updateAuthInfo(string(content))
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func createSftUrlBaseData() UrlUpdateInfo {
	baseContent := `{
		"Option": "ADD",
		"urlInfos" : [
			{
				"type": "edgecore", 
				"url": "GET https://xxxxx",
				"version": "v1.12.3",
				"createdAt": "2021-02-01 10:56:59"
			}
		]
	}`

	var req UrlUpdateInfo
	err := json.Unmarshal([]byte(baseContent), &req)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal failed")
		return req
	}
	return req
}

func testSftUrlInfoShouldSuccess() {
	req := createSftUrlBaseData()

	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}

	resp := updateSftUrlInfo(string(content))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testSftUrlInfoOptionShouldFailed() {
	req := createSftUrlBaseData()
	req.Option = "ABC"
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}

	resp := updateSftUrlInfo(string(content))
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testSftUrlInfoTypeShouldFailed() {
	req := createSftUrlBaseData()
	req.UrlInfos[0].Type = "MEFEdge"
	content, err := json.Marshal(req)
	if err != nil {
		hwlog.RunLog.Errorf("marshal failed")
	}

	resp := updateSftUrlInfo(string(content))
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testSftUrlTypeShouldFailed() {
	req := createSftUrlBaseData()

	failDataCases := []string{
		"GET ",
		"https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A!scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A\nscend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A$scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A\\scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A;scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A&scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A<scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://A>scend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"GET https://Ascend -mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
		"PATCH https://Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz",
	}
	for _, dataCase := range failDataCases {
		req.UrlInfos[0].Url = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}

		resp := updateSftUrlInfo(string(content))

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testSftUrlVersionShouldFailed() {
	req := createSftUrlBaseData()

	failDataCases := []string{
		"",
		"$",
		"*",
	}
	for _, dataCase := range failDataCases {
		req.UrlInfos[0].Version = dataCase
		content, err := json.Marshal(req)
		if err != nil {
			hwlog.RunLog.Errorf("marshal failed")
		}

		resp := updateSftUrlInfo(string(content))

		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testInnerGetSftDownloadInfo() {
	resp := innerGetSftDownloadInfo("edgecore")
	convey.So(resp.Status, convey.ShouldEqual, common.Success)

}

func TestAuthInfo(t *testing.T) {
	convey.Convey("test software url info", t, func() {
		convey.Convey("test software url auth info", func() {
			convey.Convey("test software url auth info should success", testAuthInfoShouldSuccess)
			convey.Convey("test software url auth user name should failed", testAuthUserNameShouldFailed)
			convey.Convey("test software url auth password should failed", testAuthPasswordShouldFailed)
			convey.Convey("test software url info should success", testSftUrlInfoShouldSuccess)
			convey.Convey("test software url Option should failed", testSftUrlInfoOptionShouldFailed)
			convey.Convey("test software url Type should failed", testSftUrlInfoTypeShouldFailed)
			convey.Convey("test software url para should failed", testSftUrlTypeShouldFailed)
			convey.Convey("test software url version should failed", testSftUrlVersionShouldFailed)
			convey.Convey("test inner get software url version should failed", testInnerGetSftDownloadInfo)
		})
	})
}
