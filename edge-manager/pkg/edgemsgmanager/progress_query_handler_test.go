// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager test for querying edge software download progress
package edgemsgmanager

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

func TestProgressQueryPara(t *testing.T) {
	convey.Convey("test progress query info should be success", t, testProgressQuery)
	convey.Convey("test progress query info should be failed, invalid input", t, testProgressQueryErrInput)
	convey.Convey("test progress query info should be failed, invalid content", t, testProgressQueryErrContent)
	convey.Convey("test progress query info should be failed, invalid param", t, testProgressQueryErrParam)
}

func testProgressQuery() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}

	dataCases := []string{"2102312NSF10K8000130"}
	for _, dataCase := range dataCases {
		msg.FillContent(dataCase)
		resp := queryEdgeDownloadProgress(msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	}
}

func testProgressQueryErrInput() {
	req := createDownloadSfwBaseData()
	resp := queryEdgeDownloadProgress(req)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testProgressQueryErrContent() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	msg.FillContent([]byte{})
	resp := queryEdgeDownloadProgress(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

// querying software download information is also tested here
func testProgressQueryErrParam() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}

	dataCases := []string{
		"_2102312NSF10K8000130",
		"-2102312NSF10K8000130",
		"2102312NSF10K8000130_",
		"2102312NSF10K8000130-",
		"21!02312NSF10K800013$0",
		"2102312NSF10K80001302102312NSF10K80001302102312NSF10K800013021023",
	}
	for _, dataCase := range dataCases {
		msg.FillContent(dataCase)
		resp := queryEdgeDownloadProgress(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
		resp = queryEdgeSoftwareVersion(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}
