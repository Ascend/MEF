// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlers for testing pod restart event handler
package handlers

import (
	"encoding/json"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

const podPatchMsg = `{
          "header":  {
                    "msg_id":  "939f7d28-de0b-4d75-8771-20f3b34e8171",
                    "sync":  false,
                    "timestamp":  1695764677880
          },
          "route":  {
                    "source":  "edged",
                    "group":  "meta",
                    "operation":  "patch",
                    "resource":  "websocket/podpatch/test-pod"
          },
          "content":  {
                    "status":  {
                              "containerStatuses":  [
                                        {
                                                  "containerID":  "docker://41493a3c6a5c2422f37deb18ea87c8d8fce41f",
                                                  "name":  "test-pod",
                                                  "restartCount":  1
                                        }
                              ]
                    }
          }
}`

var (
	oldPodStr       []byte
	msgPodPatch     model.Message
	podEventHandler = podRestartEventHandler{}
	oldPod          = v1.Pod{Status: v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{
		{Name: "test-pod", ContainerID: "docker://41493a3c6a5c2422f37deb18ea87c8d8fce41f"},
	}}}
)

func setupPodPatch() error {
	if err := json.Unmarshal([]byte(podPatchMsg), &msgPodPatch); err != nil {
		hwlog.RunLog.Errorf("unmarshal test pod patch message failed, error: %v", err)
		return err
	}

	var err error
	oldPodStr, err = json.Marshal(oldPod)
	if err != nil {
		hwlog.RunLog.Errorf("marshal test old pod failed, error: %v", err)
		return err
	}
	return nil
}

func TestPodRestartEventHandler(t *testing.T) {
	if err := setupPodPatch(); err != nil {
		panic(err)
	}

	p := gomonkey.ApplyFuncReturn(util.SendInnerMsgResponse, nil).
		ApplyMethodReturn(statusmanager.GetPodStatusMgr(), "Get", string(oldPodStr), nil).
		ApplyFuncReturn(almutils.SendAlarm, nil)
	defer p.Reset()

	convey.Convey("test pod restart event handler should be success", t, podRestartEventHandlerSuccess)
	convey.Convey("test pod restart event handler should be failed, get pod status failed", t, getPodStatusFailed)
	convey.Convey("test pod restart event handler should be failed, unmarshal pod str failed", t, unmarshalPodStrFailed)
	convey.Convey("test pod restart event handler should be failed, merge patch failed", t, mergePatchFailed)
}

func podRestartEventHandlerSuccess() {
	convey.Convey("pod restart event handler success", func() {
		err := podEventHandler.Handle(&msgPodPatch)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("warning: create alarm failed", func() {
		p1 := gomonkey.ApplyFuncReturn(almutils.CreateAlarm, nil, test.ErrTest)
		defer p1.Reset()
		err := podEventHandler.Handle(&msgPodPatch)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("warning: send alarm failed", func() {
		p1 := gomonkey.ApplyFuncReturn(almutils.SendAlarm, test.ErrTest)
		defer p1.Reset()
		err := podEventHandler.Handle(&msgPodPatch)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("warning: send inner msg response failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.SendInnerMsgResponse, test.ErrTest)
		defer p1.Reset()
		err := podEventHandler.Handle(&msgPodPatch)
		convey.So(err, convey.ShouldBeNil)
	})
}

func getPodStatusFailed() {
	p1 := gomonkey.ApplyMethodReturn(statusmanager.GetPodStatusMgr(), "Get", "", test.ErrTest)
	defer p1.Reset()
	err := podEventHandler.Handle(&msgPodPatch)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func unmarshalPodStrFailed() {
	p1 := gomonkey.ApplyFuncSeq(json.Unmarshal, []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},
	})
	defer p1.Reset()

	err := podEventHandler.Handle(&msgPodPatch)
	convey.So(err, convey.ShouldResemble, test.ErrTest)

	err = podEventHandler.Handle(&msgPodPatch)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func mergePatchFailed() {
	p1 := gomonkey.ApplyFuncReturn(util.MergePatch, []byte{}, test.ErrTest)
	defer p1.Reset()
	err := podEventHandler.Handle(&msgPodPatch)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}
