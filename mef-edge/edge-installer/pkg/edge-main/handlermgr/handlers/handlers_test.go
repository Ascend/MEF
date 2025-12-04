// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
