// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlers
package handlers

import (
	"encoding/json"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
)

func TestGetHandler(t *testing.T) {
	var msg model.Message
	if err := json.Unmarshal([]byte(podRestartMsg), &msg); err != nil {
		hwlog.RunLog.Errorf("unmarshal test message failed, error: %v", err)
		return
	}
	msg.Router.Option = msg.KubeEdgeRouter.Operation
	msg.Router.Resource = msg.KubeEdgeRouter.Resource
	convey.Convey("test get handler success", t, func() {
		p := gomonkey.ApplyMethodReturn(&podRestartHandler{}, "Handle", nil)
		defer p.Reset()
		handler := GetHandler()
		err := handler.Process(&msg)
		convey.So(err, convey.ShouldBeNil)
	})
}
