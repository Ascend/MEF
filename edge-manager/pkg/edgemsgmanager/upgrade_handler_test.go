// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager test for effecting edge software after upgrading
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

	"huawei.com/mindxedge/base/common"
)

func TestUpgradeSoftware(t *testing.T) {
	convey.Convey("test upgrade software should be success", t, testUpgradeSoftware)
	convey.Convey("test upgrade software should be failed, invalid param", t, testUpgradeSfwErrParam)
	convey.Convey("test upgrade software should be failed, invalid sn and sfwName", t, testUpgradeSfwErrSnAndSfwName)
	convey.Convey("test upgrade software should be failed, new msg error", t, testUpgradeSfwErrNewMsg)
	convey.Convey("test upgrade software should be failed, send sync msg error", t, testUpgradeSfwErrSendSyncMsg)
}

func createUpgradeSfwBaseData() UpgradeSoftwareReq {
	baseContent := `{
	"serialNumbers": ["2102312NSF10K8000130"],
    "softWareName": "MEFEdge"
	}`

	var req UpgradeSoftwareReq
	err := json.Unmarshal([]byte(baseContent), &req)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal failed, error: %v", err)
		return req
	}
	return req
}

func prepareUpgradeSfwRightMsg() *model.Message {
	req := createUpgradeSfwBaseData()

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	err = msg.FillContent(req, true)
	if err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
	}
	return msg
}

func testUpgradeSoftware() {
	msg := prepareUpgradeSfwRightMsg()
	if msg == nil {
		return
	}

	var p1 = gomonkey.ApplyFunc(modulemgr.SendSyncMessage,
		func(m *model.Message, duration time.Duration) (*model.Message, error) {
			rspMsg, err := model.NewMessage()
			if err != nil {
				hwlog.RunLog.Error("create message failed")
			}
			err = rspMsg.FillContent(common.OK)
			convey.So(err, convey.ShouldBeNil)
			return rspMsg, nil
		})
	defer p1.Reset()

	resp := upgradeEdgeSoftware(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUpgradeSfwErrParam() {
	resp := upgradeEdgeSoftware(&model.Message{Content: []byte("")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testUpgradeSfwErrSnAndSfwName() {
	var p1 = gomonkey.ApplyFunc(modulemgr.SendSyncMessage,
		func(m *model.Message, duration time.Duration) (*model.Message, error) {
			rspMsg, err := model.NewMessage()
			if err != nil {
				hwlog.RunLog.Errorf("create message failed, error: %v", err)
			}
			err = rspMsg.FillContent(common.OK)
			convey.So(err, convey.ShouldBeNil)
			return rspMsg, nil
		})
	defer p1.Reset()

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed, error: %v", err)
	}
	req := createUpgradeSfwBaseData()

	snCases := [][]string{
		{},
		{"_2102312NSF10K8000130"},
		{"-2102312NSF10K8000130"},
		{"2102312NSF10K8000130_"},
		{"2102312NSF10K8000130-"},
		{"2102312NSF10K8000130", "2102312NSF10K8000130"},
		{"21!02312NSF10K800013$0"},
		{"2102312NSF10K80001302102312NSF10K80001302102312NSF10K800013021023"},
	}
	for _, snCase := range snCases {
		req.SerialNumbers = snCase
		err = msg.FillContent(req)
		convey.So(err, convey.ShouldBeNil)

		resp := upgradeEdgeSoftware(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}

	sfwNameCases := []string{
		"",
		"AtlasEdge",
	}
	for _, sfwNameCase := range sfwNameCases {
		req.SoftwareName = sfwNameCase
		err = msg.FillContent(req)
		convey.So(err, convey.ShouldBeNil)

		resp := upgradeEdgeSoftware(msg)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	}
}

func testUpgradeSfwErrNewMsg() {
	msg := prepareUpgradeSfwRightMsg()
	if msg == nil {
		return
	}

	var p1 = gomonkey.ApplyFunc(model.NewMessage,
		func() (*model.Message, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()

	resp := upgradeEdgeSoftware(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorSendMsgToNode)
}

func testUpgradeSfwErrSendSyncMsg() {
	msg := prepareUpgradeSfwRightMsg()
	if msg == nil {
		return
	}

	var p1 = gomonkey.ApplyFunc(modulemgr.SendSyncMessage,
		func(m *model.Message, duration time.Duration) (*model.Message, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()

	resp := upgradeEdgeSoftware(msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorSendMsgToNode)
}
