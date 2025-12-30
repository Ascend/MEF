// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
