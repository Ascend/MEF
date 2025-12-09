// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgemsgmanager test for querying the version of the edge downloaded software
package edgemsgmanager

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
)

func TestQuerySoftwareInfo(t *testing.T) {
	var p1 = gomonkey.ApplyFunc(common.SendSyncMessageByRestful,
		func(input interface{}, router *common.Router, timeout time.Duration) common.RespMsg {
			var rsp common.RespMsg
			rsp.Status = common.Success
			var softwareInfo types.InnerSoftwareInfoResp
			softwareInfo.SoftwareInfo = append(softwareInfo.SoftwareInfo, types.SoftwareInfo{
				InactiveVersion: "v1.12",
				Name:            "edgecore",
				Version:         "v1.12"})
			rsp.Data = softwareInfo
			return rsp
		})
	defer p1.Reset()

	convey.Convey("test query software info should be success", t, testSoftwareQueryValid)
	convey.Convey("test query software info should be failed, invalid content", t, testSoftwareQueryErrContent)
	convey.Convey("test query software info should be failed, send sync msg error", t, testSfwQueryErrSendSyncMsg)
	convey.Convey("test query software info should be failed, marshal error", t, testSoftwareQueryErrMarshal)
	convey.Convey("test query software info should be failed, unmarshal error", t, testSoftwareQueryErrUnmarshal)
}

func testSoftwareQueryValid() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}

	dataCases := []string{"2102312NSF10K8000130"}
	for _, dataCase := range dataCases {
		err = msg.FillContent(dataCase)
		convey.So(err, convey.ShouldBeNil)
		rsp := queryEdgeSoftwareVersion(msg)
		convey.So(rsp.Status, convey.ShouldEqual, common.Success)
		softwareInfo, ok := rsp.Data.([]types.SoftwareInfo)
		convey.So(ok, convey.ShouldEqual, true)
		convey.So(softwareInfo[0].Name, convey.ShouldEqual, "edgecore")
		convey.So(softwareInfo[0].Version, convey.ShouldEqual, "v1.12")
		convey.So(softwareInfo[0].InactiveVersion, convey.ShouldEqual, "v1.12")
	}
}

func testSoftwareQueryErrContent() {
	resp := queryEdgeSoftwareVersion(&model.Message{Content: []byte("")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testSfwQueryErrSendSyncMsg() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent("2102312NSF10K8000130")
	convey.So(err, convey.ShouldBeNil)

	var p1 = gomonkey.ApplyFunc(common.SendSyncMessageByRestful,
		func(input interface{}, router *common.Router, timeout time.Duration) common.RespMsg {
			var rsp common.RespMsg
			rsp.Status = common.FAIL
			return rsp
		})
	defer p1.Reset()

	rsp := queryEdgeSoftwareVersion(msg)
	convey.So(rsp.Status, convey.ShouldEqual, common.ErrorGetNodeSoftwareVersion)
}

func testSoftwareQueryErrMarshal() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent("2102312NSF10K8000130")
	convey.So(err, convey.ShouldBeNil)

	var p2 = gomonkey.ApplyFunc(json.Marshal,
		func(v interface{}) ([]byte, error) {
			return nil, test.ErrTest
		})
	defer p2.Reset()
	rsp := queryEdgeSoftwareVersion(msg)
	convey.So(rsp.Status, convey.ShouldEqual, common.ErrorGetNodeSoftwareVersion)
}

func testSoftwareQueryErrUnmarshal() {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent("2102312NSF10K8000130")
	convey.So(err, convey.ShouldBeNil)

	var p2 = gomonkey.ApplyFunc(json.Unmarshal,
		func(data []byte, v interface{}) error {
			return test.ErrTest
		})
	defer p2.Reset()
	rsp := queryEdgeSoftwareVersion(msg)
	convey.So(rsp.Status, convey.ShouldEqual, common.ErrorGetNodeSoftwareVersion)
}
