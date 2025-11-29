// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager test for edge report the software upgrade progress to center
package edgemsgmanager

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
)

func TestUpdateEdgeDownloadProgress(t *testing.T) {
	convey.Convey("test report download progress should be success", t, testUpdateEdgeDownloadProgress)
	convey.Convey("test report download progress should be failed, invalid param", t, testSoftwareReportErrParam)
}

func testUpdateEdgeDownloadProgress() {
	baseContent := `{
    "serialNumber": "2102312NSF10K8000130",
    "upgradeResInfo": {
        "progress": 100,
        "res": "success",
        "msg": "aaaaasssss"
    	}
	}`

	var req types.EdgeDownloadResInfo
	err := json.Unmarshal([]byte(baseContent), &req)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal failed, error: %v", err)
		return
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent(req)
	convey.So(err, convey.ShouldBeNil)
	resp := UpdateEdgeDownloadProgress(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testSoftwareReportErrParam() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent(common.OK)
	convey.So(err, convey.ShouldBeNil)
	resp := UpdateEdgeDownloadProgress(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}
