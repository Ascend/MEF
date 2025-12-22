// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

const podRestartMsg = `{
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

var msgPodRestart model.Message

func setupPodRestart() error {
	if err := json.Unmarshal([]byte(podRestartMsg), &msgPodRestart); err != nil {
		hwlog.RunLog.Errorf("unmarshal test pod restart message failed, error: %v", err)
		return err
	}
	return nil
}

func TestPodRestartHandler(t *testing.T) {
	if err := setupPodRestart(); err != nil {
		fmt.Printf("setup test pod restart environment failed: %v\n", err)
		return
	}

	p := gomonkey.ApplyFuncReturn(modulemgr.SendAsyncMessage, nil).
		ApplyFuncReturn(common.GetFdIp, "localhost", nil)
	defer p.Reset()
	convey.Convey("test pod restart handler successful", t, podRestartHandlerSuccess)
	convey.Convey("test pod restart handler failed", t, func() {
		convey.Convey("check pod restart policy failed", checkPodRestartPolicyFailed)
		convey.Convey("restart pod by edge om failed", restartPodByEdgeOmFailed)
	})
}

func podRestartHandlerSuccess() {
	testPod := v1.Pod{}
	testPod.Spec.RestartPolicy = v1.RestartPolicyAlways
	testPodByte, err := json.Marshal(testPod)
	if err != nil {
		hwlog.RunLog.Errorf("marshal test pod status failed, error: %v", err)
		return
	}

	p := gomonkey.ApplyFuncReturn(statusmanager.GetPodStatusMgr, &mockStatusMgrImpl{}).
		ApplyMethodReturn(&mockStatusMgrImpl{}, "Get", string(testPodByte), nil).
		ApplyFuncReturn(util.SendSyncMsg, constants.Success, nil)
	defer p.Reset()
	handler := podRestartHandler{}
	err = handler.Handle(&msgPodRestart)
	convey.So(err, convey.ShouldBeNil)
}

func checkPodRestartPolicyFailed() {
	testPod := v1.Pod{}
	expectErr := errors.New("check pod restart policy failed")
	p := gomonkey.ApplyFuncReturn(statusmanager.GetPodStatusMgr, &mockStatusMgrImpl{})
	defer p.Reset()

	convey.Convey("get pod status failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&mockStatusMgrImpl{}, "Get", "", test.ErrTest)
		defer p1.Reset()
		testMsg := msgPodRestart
		handler := podRestartHandler{}
		err := handler.Handle(&testMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("unmarshal pod status failed", func() {
		p2 := gomonkey.ApplyMethodReturn(&mockStatusMgrImpl{}, "Get", "", nil).
			ApplyFuncReturn(json.Unmarshal, test.ErrTest)
		defer p2.Reset()
		testMsg := msgPodRestart
		handler := podRestartHandler{}
		err := handler.Handle(&testMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("the restart policy of pod is Never, cannot restart the pod", func() {
		testPod.Spec.RestartPolicy = v1.RestartPolicyNever
		testPodByte, err := json.Marshal(testPod)
		if err != nil {
			hwlog.RunLog.Errorf("marshal test pod status failed, error: %v", err)
			return
		}
		p3 := gomonkey.ApplyMethodReturn(&mockStatusMgrImpl{}, "Get", string(testPodByte), nil)
		defer p3.Reset()
		testMsg := msgPodRestart
		handler := podRestartHandler{}
		err = handler.Handle(&testMsg)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func restartPodByEdgeOmFailed() {
	p := gomonkey.ApplyFuncReturn(CheckPodRestartPolicy, nil)
	defer p.Reset()

	convey.Convey("send restart pod message to edge om failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, "", test.ErrTest)
		defer p1.Reset()
		handler := podRestartHandler{}
		err := handler.Handle(&msgPodRestart)
		convey.So(err, convey.ShouldResemble, errors.New("send restart pod message to edge om failed"))
	})

	convey.Convey("restart pod failed by edge om failed", func() {
		p2 := gomonkey.ApplyFuncReturn(util.SendSyncMsg, constants.Failed, nil)
		defer p2.Reset()
		handler := podRestartHandler{}
		err := handler.Handle(&msgPodRestart)
		convey.So(err, convey.ShouldResemble, errors.New("restart pod failed by edge om"))
	})
}

type mockStatusMgrImpl struct {
	typ string
}

func (f *mockStatusMgrImpl) Set(string, interface{}) error      { return nil }
func (f *mockStatusMgrImpl) Patch(string, []byte) error         { return nil }
func (f *mockStatusMgrImpl) Get(string) (string, error)         { return "", nil }
func (f *mockStatusMgrImpl) GetAll() (map[string]string, error) { return nil, nil }
func (f *mockStatusMgrImpl) Delete(string) error                { return nil }
