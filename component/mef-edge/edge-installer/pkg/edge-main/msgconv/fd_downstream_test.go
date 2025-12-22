// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

type podUpdateTestcase struct {
	description   string
	inputMsg      *model.Message
	nodeName      string
	npuCapability bool
	expectedPod   v1.Pod
}

var podUpdateTestcases = []podUpdateTestcase{
	{
		description: "test pod without npu resource",
		inputMsg: mustCreateMsg(
			messageHeader{Sync: true},
			model.MessageRoute{Operation: constants.OptUpdate, Source: constants.ControllerModule,
				Resource: constants.ActionPod + "pod-update-1"},
			v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1024")},
			}}}}},
		),
		nodeName:      "test-node-1",
		npuCapability: true,
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1024")},
		}}}}},
	},
	{
		description: "test pod with npu when npu sharing cap is enabled",
		inputMsg: mustCreateMsg(
			messageHeader{Sync: true},
			model.MessageRoute{Operation: constants.OptUpdate, Source: constants.ControllerModule,
				Resource: constants.ActionPod + "pod-update-2"},
			v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits:   v1.ResourceList{constants.CenterNpuName: resource.MustParse("0.2")},
				Requests: v1.ResourceList{constants.CenterNpuName: resource.MustParse("0.3")},
			}}}}},
		),
		nodeName:      "test-node-2",
		npuCapability: true,
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits:   v1.ResourceList{constants.CenterNpuName: resource.MustParse("20")},
			Requests: v1.ResourceList{constants.CenterNpuName: resource.MustParse("30")},
		}}}}},
	},
	{
		description: "test pod with npu when npu sharing cap is disabled",
		inputMsg: mustCreateMsg(
			messageHeader{Sync: true},
			model.MessageRoute{Operation: constants.OptUpdate, Source: constants.ControllerModule,
				Resource: constants.ActionPod + "pod-update-3"},
			v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits:   v1.ResourceList{constants.CenterNpuName: resource.MustParse("2")},
				Requests: v1.ResourceList{constants.CenterNpuName: resource.MustParse("3")},
			}}}}},
		),
		nodeName:      "test-node-3",
		npuCapability: false,
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits:   v1.ResourceList{constants.CenterNpuName: resource.MustParse("2")},
			Requests: v1.ResourceList{constants.CenterNpuName: resource.MustParse("3")},
		}}}}},
	},
}

func TestModifyPodForUpdate(t *testing.T) {
	convey.Convey("update:websocket/pod/", t, func() {
		for _, tc := range podUpdateTestcases {
			fmt.Println(tc.description)
			convey.So(ensureNodeExists(tc.nodeName, v1.Node{}), convey.ShouldBeNil)
			config.GetCapabilityCache().Set(constants.CapabilityNpuSharingConfig, tc.npuCapability)

			var actualPod v1.Pod
			convey.So(convertToEdgeCoreMsg(tc.inputMsg, &actualPod), convey.ShouldBeNil)

			convey.So(tc.inputMsg.KubeEdgeRouter.Source, convey.ShouldEqual, constants.EdgeControllerModule)
			convey.So(tc.inputMsg.Header.Sync, convey.ShouldBeFalse)
			convey.So(actualPod.Spec.NodeName, convey.ShouldEqual, tc.nodeName)
			convey.So(actualPod.Kind, convey.ShouldEqual, "Pod")
			convey.So(actualPod.Spec.EnableServiceLinks, convey.ShouldNotBeNil)
			convey.So(*actualPod.Spec.EnableServiceLinks, convey.ShouldBeTrue)
			for idx, c := range tc.expectedPod.Spec.Containers {
				checkPodResource(actualPod.Spec.Containers[idx].Resources, c.Resources)
			}
		}
	})
}

func checkPodResource(actual, expected v1.ResourceRequirements) {
	convey.So(len(actual.Limits), convey.ShouldEqual, len(expected.Limits))
	for k, v := range expected.Limits {
		convey.So(actual.Limits[k], convey.ShouldEqual, v)
	}
	convey.So(len(actual.Requests), convey.ShouldEqual, len(expected.Requests))
	for k, v := range expected.Requests {
		convey.So(actual.Requests[k], convey.ShouldEqual, v)
	}
}

func TestModifySecret(t *testing.T) {
	type testcase struct {
		description string
		inputMsg    *model.Message
	}
	testcases := []testcase{
		{
			description: "test secret update",
			inputMsg: mustCreateMsg(
				messageHeader{},
				model.MessageRoute{Operation: constants.OptUpdate, Source: constants.ControllerModule,
					Resource: constants.ActionSecret},
				v1.Secret{Data: map[string][]byte{"key": []byte("my-value")}},
			),
		},
	}

	convey.Convey("websocket/secret/", t, func() {
		for _, tc := range testcases {
			fmt.Println(tc.description)

			var data v1.Secret
			err := tc.inputMsg.ParseContent(&data)
			convey.So(err, convey.ShouldBeNil)
			var actualSecret v1.Secret
			err = convertToEdgeCoreMsg(tc.inputMsg, &actualSecret)
			convey.So(err, convey.ShouldBeNil)

			convey.So(tc.inputMsg.KubeEdgeRouter.Source, convey.ShouldEqual, constants.EdgeControllerModule)
			convey.So(actualSecret.Kind, convey.ShouldEqual, "Secret")
			convey.So(actualSecret.Data, convey.ShouldResemble, data.Data)
		}
	})
}

func TestModifyConfigMap(t *testing.T) {
	type testcase struct {
		description string
		inputMsg    *model.Message
	}
	testcases := []testcase{
		{
			description: "test configmap update",
			inputMsg: mustCreateMsg(
				messageHeader{},
				model.MessageRoute{Operation: constants.OptUpdate, Source: constants.ControllerModule,
					Resource: constants.ActionConfigmap + "configmap-update"},
				v1.ConfigMap{Data: map[string]string{"key": "my-value"}},
			),
		},
		{
			description: "test configmap deletion",
			inputMsg: mustCreateMsg(
				messageHeader{},
				model.MessageRoute{Operation: constants.OptDelete, Source: constants.ControllerModule,
					Resource: constants.ActionConfigmap + "configmap-delete"},
				v1.ConfigMap{Data: map[string]string{"key": "my-value"}},
			),
		},
	}

	convey.Convey("websocket/configmap/", t, func() {
		for _, tc := range testcases {
			fmt.Println(tc.description)

			var data v1.ConfigMap
			err := tc.inputMsg.ParseContent(&data)
			convey.So(err, convey.ShouldBeNil)
			var actualConfigMap v1.ConfigMap
			err = convertToEdgeCoreMsg(tc.inputMsg, &actualConfigMap)
			convey.So(err, convey.ShouldBeNil)

			convey.So(tc.inputMsg.KubeEdgeRouter.Source, convey.ShouldEqual, constants.EdgeControllerModule)
			convey.So(actualConfigMap.Kind, convey.ShouldEqual, "Configmap")
			if tc.inputMsg.KubeEdgeRouter.Operation == constants.OptUpdate {
				convey.So(actualConfigMap.Data, convey.ShouldResemble, data.Data)
			}
		}
	})
}

func TestModifyPodForDelete(t *testing.T) {
	type testcase struct {
		description string
		inputMsg    *model.Message
	}
	testcases := []testcase{
		{
			description: "delete-pod message modification functional test",
			inputMsg: mustCreateMsg(
				messageHeader{},
				model.MessageRoute{Operation: constants.OptDelete, Source: constants.ControllerModule,
					Resource: constants.ActionPod + "pod-delete-1"},
				v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Name: "container-0"}}}},
			),
		},
	}

	convey.Convey("delete:websocket/pod/", t, func() {
		for _, tc := range testcases {
			fmt.Println(tc.description)

			var resultPod v1.Pod
			err := convertToEdgeCoreMsg(tc.inputMsg, &resultPod)
			convey.So(err, convey.ShouldBeNil)

			convey.So(tc.inputMsg.KubeEdgeRouter.Source, convey.ShouldEqual, constants.EdgeControllerModule)
			convey.So(resultPod.Kind, convey.ShouldEqual, "Pod")
			convey.So(resultPod.Spec, convey.ShouldResemble, v1.PodSpec{})
		}
	})
}
