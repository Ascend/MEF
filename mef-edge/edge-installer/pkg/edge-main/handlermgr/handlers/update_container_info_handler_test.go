// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlers
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
)

const updateContainerInfoMsg = `{
    "header":{
        "msg_id":"05843e95-9069-45a7-93f2-25302be579d9",
        "timestamp":1690278946,
        "sync":true
    },
    "route":{
        "source":"controller",
        "group":"hardware",
        "operation":"update",
        "resource":"websocket/container_info"
    },
    "content":{}
}`

var msgUpdateContainer model.Message

func setupUpdateContainerInfoHandler() error {
	if err := json.Unmarshal([]byte(updateContainerInfoMsg), &msgUpdateContainer); err != nil {
		hwlog.RunLog.Errorf("unmarshal test update container info handler message failed, error: %v", err)
		return err
	}
	return nil
}

func TestUpdateContainerInfoHandler(t *testing.T) {
	if err := setupUpdateContainerInfoHandler(); err != nil {
		fmt.Printf("setup test update container info handler environment failed: %v\n", err)
		return
	}

	p := gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, nil)
	defer p.Reset()

	convey.Convey("test update container info handler successful", t, updateContainerInfoHandlerSuccess)
	convey.Convey("test update container info handler failed", t, func() {
		convey.Convey("parse container info content failed", parseContentFailed)
		convey.Convey("update container info failed", updateContainerInfoFailed)
	})
}

func updateContainerInfoHandlerSuccess() {
	p := gomonkey.ApplyMethodReturn(&UpdateContainerInfo{}, "EffectModelFile", nil)
	defer p.Reset()
	handler := updateContainerInfoHandler{}
	err := handler.Handle(&msgUpdateContainer)
	convey.So(err, convey.ShouldBeNil)
}

func parseContentFailed() {
	testMsg := msgUpdateContainer
	err := testMsg.FillContent(model.RawMessage{})
	convey.So(err, convey.ShouldBeNil)
	handler := updateContainerInfoHandler{}
	err = handler.Handle(&testMsg)
	convey.So(err, convey.ShouldResemble, errors.New("parse container info content failed"))
}

func updateContainerInfoFailed() {
	p := gomonkey.ApplyMethodReturn(&UpdateContainerInfo{}, "EffectModelFile", test.ErrTest)
	defer p.Reset()
	handler := updateContainerInfoHandler{}
	err := handler.Handle(&msgUpdateContainer)
	convey.So(err, convey.ShouldResemble, errors.New("update container info failed"))
}
