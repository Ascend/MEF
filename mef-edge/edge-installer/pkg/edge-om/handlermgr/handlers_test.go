// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr for deal every handler
package handlermgr

import (
	"encoding/json"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
)

func TestGetHandlerMgr(t *testing.T) {
	var msg model.Message
	if err := json.Unmarshal([]byte(restartPodMsg), &msg); err != nil {
		hwlog.RunLog.Errorf("unmarshal test message failed, error: %v", err)
		return
	}
	msg.Router.Option = msg.KubeEdgeRouter.Operation
	msg.Router.Resource = msg.KubeEdgeRouter.Resource
	convey.Convey("test get handler mgr success", t, func() {
		p := gomonkey.ApplyMethodReturn(&restartPodHandler{}, "Handle", nil)
		defer p.Reset()
		handler := GetHandlerMgr()
		err := handler.Process(&msg)
		convey.So(err, convey.ShouldBeNil)
	})
}
