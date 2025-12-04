// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package msgconv
package msgconv

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/database"
)

var (
	testStartTime    = mustParseTime(time.RFC3339, time.Now().Format(time.RFC3339))
	testStartTimeStr = testStartTime.UTC().Format(time.RFC3339)
)

type podPatchTestcase struct {
	description string
	inputMsg    *model.Message
	currentPod  *corev1.Pod
	expectedPod corev1.Pod
}

var podPatchTestCases = []podPatchTestcase{
	{
		description: "state of container is changed",
		inputMsg: mustCreateMsg(
			messageHeader{},
			model.MessageRoute{Operation: constants.OptPatch, Source: constants.EdgedModule,
				Group: constants.MetaModule, Resource: constants.ActionPodPatch + "pod-patch-1"},
			fmt.Sprintf(`{"status": {"containerStatuses": [{"state": {"running": {"startedAt": "%s"}}}, `+
				`{"state": {"running": null, "terminated": {"exitCode": 255}}, "restartCount": 1}]}}`, testStartTimeStr),
		),
		currentPod: &corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
			{State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{StartedAt: metav1.NewTime(testStartTime)}}},
			{State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{StartedAt: metav1.NewTime(testStartTime)}}},
		}}},
		expectedPod: corev1.Pod{Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
			{State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{StartedAt: metav1.NewTime(testStartTime)}}},
			{State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{ExitCode: 255}}, RestartCount: 1},
		}}},
	},
	{
		description: "phase of container is changed",
		inputMsg: mustCreateMsg(
			messageHeader{},
			model.MessageRoute{Operation: constants.OptPatch, Source: constants.EdgedModule,
				Group: constants.MetaModule, Resource: constants.ActionPodPatch + "pod-patch-2"},
			`{"status": {"phase": "Failed"}}`,
		),
		currentPod:  &corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodRunning}},
		expectedPod: corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodFailed}},
	},
}

func TestModifyPodPatch(t *testing.T) {
	convey.Convey("update:websocket/podpatch/", t, func() {
		const maxContainer = 20
		patches := gomonkey.ApplyFuncReturn(configpara.GetPodConfig, config.PodConfig{
			ContainerConfig: config.ContainerConfig{MaxContainerNumber: maxContainer}})
		defer patches.Reset()

		for _, tc := range podPatchTestCases {
			fmt.Println(tc.description)
			msgID := tc.inputMsg.Header.ID
			res := tc.inputMsg.KubeEdgeRouter.Resource

			dataBytes, err := json.Marshal(tc.currentPod)
			convey.So(err, convey.ShouldBeNil)
			stmt := test.MockGetDb().Save(database.Meta{
				Key: strings.Replace(tc.inputMsg.KubeEdgeRouter.Resource,
					constants.ActionPodPatch, constants.ActionPod, 1),
				Type:  constants.ResourceTypePod,
				Value: string(dataBytes),
			})
			convey.So(stmt.Error, convey.ShouldBeNil)

			var podResp PodResp
			err = processEdgeCoreMsg(tc.inputMsg, &podResp)
			convey.So(err, convey.ShouldBeNil)

			convey.So(tc.inputMsg.Header.ParentID, convey.ShouldEqual, msgID)
			convey.So(tc.inputMsg.KubeEdgeRouter.Group, convey.ShouldEqual, constants.ResourceModule)
			convey.So(tc.inputMsg.KubeEdgeRouter.Operation, convey.ShouldEqual, constants.OptResponse)
			convey.So(tc.inputMsg.KubeEdgeRouter.Resource, convey.ShouldEqual, res)

			convey.So(*podResp.Object, convey.ShouldResemble, tc.expectedPod)
			convey.So(podResp.Err, convey.ShouldResemble, apierrors.StatusError{})
		}
	})
}

type nodePatchTestcase struct {
	description  string
	input        *model.Message
	currentNode  *corev1.Node
	expectedNode corev1.Node
}

var nodePatchTestCases = []nodePatchTestcase{
	{
		description: "npu is added by patch",
		input: mustCreateMsg(
			messageHeader{},
			model.MessageRoute{Operation: constants.OptPatch, Source: constants.EdgedModule,
				Group: constants.MetaModule, Resource: constants.ActionDefaultNodePatch + "node-patch-1"},
			`{"status": {"capacity": {"huawei.com/Ascend310": 1}}}`,
		),
		currentNode: &corev1.Node{Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
		}},
		expectedNode: corev1.Node{Status: corev1.NodeStatus{Capacity: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("2"), "huawei.com/Ascend310": resource.MustParse("1")},
		}},
	},
	{
		description: "cpu is changed by patch",
		input: mustCreateMsg(
			messageHeader{},
			model.MessageRoute{Operation: constants.OptPatch, Source: constants.EdgedModule,
				Group: constants.MetaModule, Resource: constants.ActionDefaultNodePatch + "node-patch-2"},
			`{"status": {"capacity": {"cpu": 1}}}`,
		),
		currentNode: &corev1.Node{Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
		}},
		expectedNode: corev1.Node{Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
		}},
	},
	{
		description: "npu is removed by patch",
		input: mustCreateMsg(
			messageHeader{},
			model.MessageRoute{Operation: constants.OptPatch, Source: constants.EdgedModule,
				Group: constants.MetaModule, Resource: constants.ActionDefaultNodePatch + "node-patch-3"},
			`{"status": {"capacity": {"huawei.com/Ascend310": null}}}`,
		),
		currentNode: &corev1.Node{Status: corev1.NodeStatus{Capacity: corev1.ResourceList{
			corev1.ResourceCPU: resource.MustParse("2"), "huawei.com/Ascend310": resource.MustParse("1")}},
		},
		expectedNode: corev1.Node{Status: corev1.NodeStatus{
			Capacity: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
		}},
	},
}

func TestModifyNodeForPatch(t *testing.T) {
	convey.Convey("update:default/nodepatch/", t, func() {
		for _, tc := range nodePatchTestCases {
			fmt.Println(tc.description)
			msgID := tc.input.Header.ID
			res := tc.input.KubeEdgeRouter.Resource

			err := ensureNodeExists(filepath.Base(tc.input.KubeEdgeRouter.Resource), *tc.currentNode)
			convey.So(err, convey.ShouldBeNil)

			var nodeResp NodeResp
			err = processEdgeCoreMsg(tc.input, &nodeResp)
			convey.So(err, convey.ShouldBeNil)

			convey.So(tc.input.Header.ParentID, convey.ShouldEqual, msgID)
			convey.So(tc.input.KubeEdgeRouter.Group, convey.ShouldEqual, constants.ResourceModule)
			convey.So(tc.input.KubeEdgeRouter.Operation, convey.ShouldEqual, constants.OptResponse)
			convey.So(tc.input.KubeEdgeRouter.Resource, convey.ShouldEqual, res)

			convey.So(*nodeResp.Object, convey.ShouldResemble, tc.expectedNode)
			convey.So(nodeResp.Err, convey.ShouldResemble, apierrors.StatusError{})
		}
	})
}

func TestModifyNodeInsertion(t *testing.T) {
	type testcase struct {
		description string
		input       *model.Message
	}
	testcases := []testcase{{
		description: "insert-node message modification functional test",
		input: mustCreateMsg(
			messageHeader{},
			model.MessageRoute{Operation: constants.OptInsert, Source: constants.EdgedModule,
				Group: constants.MetaModule, Resource: constants.ActionDefaultNodeStatus + "node-insert-1"},
			corev1.Node{Status: corev1.NodeStatus{
				Capacity:    corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
				Allocatable: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
			}},
		),
	}}

	convey.Convey("insert:default/node/", t, func() {
		for _, tc := range testcases {
			fmt.Println(tc.description)
			msgID := tc.input.Header.ID
			res := tc.input.KubeEdgeRouter.Resource

			var data corev1.Node
			err := tc.input.ParseContent(&data)
			convey.So(err, convey.ShouldBeNil)
			var resp NodeResp
			err = processEdgeCoreMsg(tc.input, &resp)
			convey.So(err, convey.ShouldBeNil)

			convey.So(tc.input.Header.ParentID, convey.ShouldEqual, msgID)
			convey.So(tc.input.KubeEdgeRouter.Group, convey.ShouldEqual, constants.ResourceModule)
			convey.So(tc.input.KubeEdgeRouter.Operation, convey.ShouldEqual, constants.OptResponse)
			convey.So(tc.input.KubeEdgeRouter.Resource, convey.ShouldEqual, res)

			data.Kind = "Node"
			convey.So(*resp.Object, convey.ShouldResemble, data)
			convey.So(resp.Err, convey.ShouldResemble, apierrors.StatusError{})
		}
	})
}

func TestModifyNodeQuery(t *testing.T) {
	type testcase struct {
		description       string
		input             *model.Message
		currentNodeStatus corev1.Node
	}
	var testcases = []testcase{
		{
			description: "insert-node message modification functional test",
			input: mustCreateMsg(
				messageHeader{},
				model.MessageRoute{Operation: constants.OptQuery, Source: constants.EdgedModule,
					Group: constants.MetaModule, Resource: constants.ActionDefaultNodeStatus + "node-query-1"},
				nil,
			),
			currentNodeStatus: corev1.Node{Status: corev1.NodeStatus{
				Capacity: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")},
			}},
		},
	}

	convey.Convey("query:default/node/", t, func() {
		for _, tc := range testcases {
			fmt.Println(tc.description)
			msgID := tc.input.Header.ID
			res := tc.input.KubeEdgeRouter.Resource

			err := ensureNodeExists(filepath.Base(tc.input.KubeEdgeRouter.Resource), tc.currentNodeStatus)
			convey.So(err, convey.ShouldBeNil)

			var actualNode corev1.Node
			err = processEdgeCoreMsg(tc.input, &actualNode)
			convey.So(err, convey.ShouldBeNil)

			convey.So(tc.input.Header.ParentID, convey.ShouldEqual, msgID)
			convey.So(tc.input.KubeEdgeRouter.Group, convey.ShouldEqual, constants.ResourceModule)
			convey.So(tc.input.KubeEdgeRouter.Operation, convey.ShouldEqual, constants.OptResponse)
			convey.So(tc.input.KubeEdgeRouter.Resource, convey.ShouldEqual, res)

			convey.So(len(actualNode.Spec.PodCIDR), convey.ShouldBeGreaterThan, 0)
			convey.So(len(actualNode.Spec.PodCIDRs), convey.ShouldBeGreaterThan, 0)
			convey.So(actualNode.Status, convey.ShouldResemble, tc.currentNodeStatus.Status)
		}
	})
}

func processEdgeCoreMsg(input *model.Message, outputContent interface{}) error {
	var output *model.Message
	patches := gomonkey.ApplyFunc(modulemgr.SendAsyncMessage, func(msg *model.Message) error {
		if msg.Header.ParentID == input.Header.ID {
			output = msg
		}
		return nil
	})
	defer patches.Reset()

	if err := convertToFdMsg(input, nil); err != nil {
		return err
	}
	if output == nil {
		return errors.New("no response")
	}

	*input = *output
	return json.Unmarshal(output.Content, outputContent)
}
