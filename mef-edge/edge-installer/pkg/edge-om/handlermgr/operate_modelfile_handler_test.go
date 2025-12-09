// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr for deal every handler
package handlermgr

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

const optInvalid = "Invalid"

func getOperateMsg(operate string) (*model.Message, error) {
	operateContent, err := getOperateContent(operate)
	if err != nil {
		hwlog.RunLog.Errorf("get test operate content failed, error: %v", err)
		return nil, err
	}
	msg, err := util.NewInnerMsgWithFullParas(util.InnerMsgParams{
		Source:                constants.ModHandlerMgr,
		Destination:           constants.ModEdgeOm,
		Operation:             constants.OptRaw,
		Resource:              constants.ActionModelFiles,
		Content:               operateContent,
		TransferStructIntoStr: true,
	})
	if err != nil {
		hwlog.RunLog.Errorf("new test message failed, error: %v", err)
		return nil, err
	}
	return msg, nil
}

func TestOperateModelFileHandler(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
	defer p.Reset()
	convey.Convey("test operate model file handler successful", t, operateModelFileHandlerSuccess)
	convey.Convey("test operate model file handler failed", t, func() {
		convey.Convey("parse operate content failed", parseContentFailed)
		convey.Convey("check operate content failed", operateContentCheckFailed)
		convey.Convey("sync files failed", syncFilesFailed)
		convey.Convey("operate model file failed", operateModelFileFailed)
	})
}

func operateModelFileHandlerSuccess() {
	convey.Convey("sync files success", func() {
		testMsg, err := getOperateMsg(constants.OptSync)
		if err != nil {
			return
		}
		p := gomonkey.ApplyFuncReturn(json.Marshal, []byte{}, nil)
		defer p.Reset()
		handler := operateModelFileHandler{}
		err = handler.Handle(testMsg)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("operate model file success", func() {
		testMsg, err := getOperateMsg(constants.OptCheck)
		if err != nil {
			return
		}
		p := gomonkey.ApplyMethodReturn(&OperateModelFile{}, "OperateModelFile", nil)
		defer p.Reset()
		handler := operateModelFileHandler{}
		err = handler.Handle(testMsg)
		convey.So(err, convey.ShouldBeNil)
	})
}

func parseContentFailed() {
	testMsg, err := getOperateMsg(optInvalid)
	if err != nil {
		return
	}
	err = testMsg.FillContent(model.RawMessage{})
	convey.So(err, convey.ShouldBeNil)
	handler := operateModelFileHandler{}
	err = handler.Handle(testMsg)
	convey.So(err, convey.ShouldResemble, errors.New("parse operate content failed"))
}

func operateContentCheckFailed() {
	testMsg, err := getOperateMsg(optInvalid)
	if err != nil {
		return
	}
	handler := operateModelFileHandler{}
	err = handler.Handle(testMsg)
	convey.So(err, convey.ShouldResemble, errors.New("check operate content failed"))
}

func syncFilesFailed() {
	testMsg, err := getOperateMsg(constants.OptSync)
	if err != nil {
		return
	}
	p := gomonkey.ApplyFuncReturn(json.Marshal, nil, testErr)
	defer p.Reset()
	handler := operateModelFileHandler{}
	err = handler.Handle(testMsg)
	convey.So(err, convey.ShouldResemble, errors.New("marshal sync list fail"))
}

func operateModelFileFailed() {
	testMsg, err := getOperateMsg(constants.OptCheck)
	if err != nil {
		return
	}
	p := gomonkey.ApplyMethodReturn(&OperateModelFile{}, "OperateModelFile", testErr)
	defer p.Reset()
	handler := operateModelFileHandler{}
	err = handler.Handle(testMsg)
	convey.So(err, convey.ShouldResemble, errors.New("operate model file failed"))
}
