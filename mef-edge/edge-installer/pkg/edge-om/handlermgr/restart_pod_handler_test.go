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
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
)

const (
	dockerPsOutNone = `CONTAINER ID  IMAGE  COMMAND  CREATED  STATUS  PORTS  NAMES`
	dockerPsOut     = `CONTAINER ID  IMAGE  COMMAND  CREATED  STATUS  PORTS  NAMES
84eba223c0d1  k8s.gcr.io/pause  "/pause" 23 hours  Up 23 hours  k8s_POD_test-model-1e6328ca-33ee-40a7-bf94-767d17491ee3`
	restartPodMsg = `{
    "header":{
        "msg_id":"b5b0a18d-3af2-47d9-a60d-ce6e94b58eca",
        "parent_msg_id":"",
        "timestamp":1691579247,
        "sync":true
    },
    "route":{
        "source":"controller",
        "group":"resource",
        "operation":"restart",
        "resource":"websocket/pod/"
    },
    "content":"test-model-1e6328ca-33ee-40a7-bf94-767d17491ee3"
}`
)

var podRestartMsg model.Message

func setupRestartPod() error {
	if err := json.Unmarshal([]byte(restartPodMsg), &podRestartMsg); err != nil {
		hwlog.RunLog.Errorf("unmarshal test pod restart message failed, error: %v", err)
		return err
	}
	return nil
}

func TestPodRestartHandler(t *testing.T) {
	if err := setupRestartPod(); err != nil {
		return
	}
	p := gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, nil)
	defer p.Reset()
	convey.Convey("test restart pod handler successful", t, restartPodSuccess)
	convey.Convey("test restart pod handler failed", t, restartPodFailed)
}

func restartPodSuccess() {
	p := gomonkey.ApplyFuncSeq(envutils.RunCommand, []gomonkey.OutputCell{
		{Values: gomonkey.Params{dockerPsOut, nil}},
		{Values: gomonkey.Params{"", nil}},
	})
	defer p.Reset()
	handler := restartPodHandler{}
	err := handler.Handle(&podRestartMsg)
	convey.So(err, convey.ShouldBeNil)
}

func restartPodFailed() {
	expectErr := errors.New("restart pod failed")
	convey.Convey("get running container info failed", func() {
		p := gomonkey.ApplyFuncReturn(envutils.RunCommand, "", testErr)
		defer p.Reset()
		handler := restartPodHandler{}
		err := handler.Handle(&podRestartMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("do not find running pod", func() {
		p := gomonkey.ApplyFuncReturn(envutils.RunCommand, dockerPsOutNone, nil)
		defer p.Reset()
		handler := restartPodHandler{}
		err := handler.Handle(&podRestartMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("stop pod failed", func() {
		p := gomonkey.ApplyFuncSeq(envutils.RunCommand, []gomonkey.OutputCell{
			{Values: gomonkey.Params{dockerPsOut, nil}},
			{Values: gomonkey.Params{"", testErr}},
		})
		defer p.Reset()
		handler := restartPodHandler{}
		err := handler.Handle(&podRestartMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
