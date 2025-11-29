// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package msgconv
package msgconv

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
)

type mefNodeResponseTestcase struct {
	description   string
	inputMsg      *model.Message
	realNpuName   string
	npuCapability bool
	existingNode  v1.Node
	expectedNode  v1.Node
}

var mefNodeResponseTestcases = []mefNodeResponseTestcase{
	{
		description: "test node response without npu",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ActionDefaultNodeStatus + "mef-node-response-1"},
			v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-response-1"}, Status: v1.NodeStatus{
				Allocatable: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1")}}}),
		existingNode: v1.Node{Status: v1.NodeStatus{Allocatable: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("2")}}},
		expectedNode: v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-response-1"},
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("2")}}},
		realNpuName: ascend310p,
	},
	{
		description: "test node response with npu without npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ActionDefaultNodeStatus + "mef-node-response-2"},
			v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-response-2"}, Status: v1.NodeStatus{
				Allocatable: v1.ResourceList{constants.CenterNpuName: resource.MustParse("1")}}}),
		existingNode: v1.Node{
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{
				ascend310p: resource.MustParse("2")}}},
		expectedNode: v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-response-2"},
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{
				ascend310p: resource.MustParse("2")}}},
		realNpuName: ascend310p,
	},
	{
		description: "test node response with npu with npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ActionDefaultNodeStatus + "mef-node-response-3"},
			v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-response-3"}, Status: v1.NodeStatus{
				Allocatable: v1.ResourceList{constants.CenterNpuName: resource.MustParse("1")}}}),
		existingNode: v1.Node{
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{
				ascend310p: resource.MustParse("200")}}},
		expectedNode: v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-response-3"},
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{
				ascend310p: resource.MustParse("200")}}},
		npuCapability: true,
		realNpuName:   ascend310p,
	},
}

func TestModifyNodeResponse(t *testing.T) {
	convey.Convey("response:default/node/", t, func() {
		var patches *gomonkey.Patches
		defer func() {
			if patches != nil {
				patches.Reset()
			}
		}()
		for _, tc := range mefNodeResponseTestcases {
			fmt.Println(tc.description)
			patches = gomonkey.ApplyFuncReturn(common.LoadNpuFromDb, tc.realNpuName, true)
			config.GetCapabilityCache().Set(constants.CapabilityNpuSharingConfig, tc.npuCapability)
			err := ensureNodeExists(filepath.Base(tc.inputMsg.KubeEdgeRouter.Resource), tc.existingNode)
			convey.So(err, convey.ShouldBeNil)

			var actualNode v1.Node
			err = convertToEdgeCoreMsg(tc.inputMsg, &actualNode)
			convey.So(err, convey.ShouldBeNil)
			convey.So(actualNode, convey.ShouldResemble, tc.expectedNode)

			patches.Reset()
			patches = nil
		}
	})
}

type mefNodePatchResponseTestcase struct {
	description   string
	inputMsg      *model.Message
	realNpuName   string
	npuCapability bool
	existingNode  v1.Node
	expectedNode  v1.Node
}

var mefNodePatchResponseTestcases = []mefNodePatchResponseTestcase{
	{
		description: "test node patch without npu",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ActionDefaultNodePatch + "mef-node-patch-1"},
			NodeResp{Object: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-patch-1"}, Status: v1.NodeStatus{
				Allocatable: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1")}}}}),
		existingNode: v1.Node{
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{v1.ResourceCPU: resource.MustParse("2")}}},
		expectedNode: v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-patch-1"},
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{v1.ResourceCPU: resource.MustParse("2")}}},
		realNpuName: ascend310p,
	},
	{
		description: "test node patch with npu without npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ActionDefaultNodePatch + "mef-node-patch-2"},
			NodeResp{Object: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-patch-2"}, Status: v1.NodeStatus{
				Allocatable: v1.ResourceList{constants.CenterNpuName: resource.MustParse("1")}}}}),
		existingNode: v1.Node{
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{ascend310p: resource.MustParse("2")}}},
		expectedNode: v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-patch-2"},
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{ascend310p: resource.MustParse("2")}}},
		realNpuName: ascend310p,
	},
	{
		description: "test node patch with npu with npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ActionDefaultNodePatch + "mef-node-patch-3"},
			NodeResp{Object: &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-patch-3"}, Status: v1.NodeStatus{
				Allocatable: v1.ResourceList{constants.CenterNpuName: resource.MustParse("1")}}}}),
		existingNode: v1.Node{
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{ascend310p: resource.MustParse("200")}}},
		expectedNode: v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "mef-node-patch-3"},
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{ascend310p: resource.MustParse("200")}}},
		npuCapability: true,
		realNpuName:   ascend310p,
	},
}

func TestModifyNodePatchResponse(t *testing.T) {
	convey.Convey("response:default/nodepatch/", t, func() {
		var patches *gomonkey.Patches
		defer func() {
			if patches != nil {
				patches.Reset()
			}
		}()
		for _, tc := range mefNodePatchResponseTestcases {
			fmt.Println(tc.description)
			patches = gomonkey.ApplyFuncReturn(common.LoadNpuFromDb, tc.realNpuName, true)
			config.GetCapabilityCache().Set(constants.CapabilityNpuSharingConfig, tc.npuCapability)
			err := ensureNodeExists(filepath.Base(tc.inputMsg.KubeEdgeRouter.Resource), tc.existingNode)
			convey.So(err, convey.ShouldBeNil)

			var resp NodeResp
			err = convertToEdgeCoreMsg(tc.inputMsg, &resp)
			convey.So(err, convey.ShouldBeNil)
			convey.So(*resp.Object, convey.ShouldResemble, tc.expectedNode)

			patches.Reset()
			patches = nil
		}
	})
}

type mefPodPatchResponseTestcase struct {
	description   string
	inputMsg      *model.Message
	realNpuName   string
	npuCapability bool
	expectedPod   v1.Pod
}

var mefPodPatchResponseTestcases = []mefPodPatchResponseTestcase{
	{
		description: "test pod patch without npu",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ResMefPodPatchPrefix + "mef-pod-patch-1"},
			PodResp{Object: &v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("2")}}}}}}}),
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("2")}}}}}},
		realNpuName: ascend310p,
	},
	{
		description: "test pod patch with npu without npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ResMefPodPatchPrefix + "mef-pod-patch-2"},
			PodResp{Object: &v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{constants.CenterNpuName: resource.MustParse("2")}}}}}}}),
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{ascend310p: resource.MustParse("2")}}}}}},
		realNpuName: ascend310p,
	},
	{
		description: "test pod patch with npu with npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptResponse, Resource: constants.ResMefPodPatchPrefix + "mef-pod-patch-3"},
			PodResp{Object: &v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{constants.CenterNpuName: resource.MustParse("2")}}}}}}}),
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{ascend310p: resource.MustParse("200")}}}}}},
		npuCapability: true,
		realNpuName:   ascend310p,
	},
}

func TestModifyPodPatchResponse(t *testing.T) {
	convey.Convey("response:mef-user/podpatch/", t, func() {
		var patches *gomonkey.Patches
		defer func() {
			if patches != nil {
				patches.Reset()
			}
		}()
		for _, tc := range mefPodPatchResponseTestcases {
			fmt.Println(tc.description)
			patches = gomonkey.ApplyFuncReturn(common.LoadNpuFromDb, tc.realNpuName, true)
			config.GetCapabilityCache().Set(constants.CapabilityNpuSharingConfig, tc.npuCapability)

			var resp PodResp
			err := convertToEdgeCoreMsg(tc.inputMsg, &resp)
			convey.So(err, convey.ShouldBeNil)
			convey.So(*resp.Object, convey.ShouldResemble, tc.expectedPod)

			patches.Reset()
			patches = nil
		}
	})
}

func TestModifyPodDelete(t *testing.T) {
	type testcase struct {
		description string
		inputMsg    *model.Message
	}
	testcases := []testcase{
		{
			description: "mef delete pod message modification functional test",
			inputMsg: mustCreateMsg(messageHeader{},
				model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
					Operation: constants.OptDelete, Resource: constants.ResMefPodPrefix + "mef-pod-delete-1"},
				v1.Pod{}),
		},
	}

	convey.Convey("delete:mef-user/pod/", t, func() {
		for _, tc := range testcases {
			var originPod v1.Pod
			err := tc.inputMsg.ParseContent(&originPod)
			convey.So(err, convey.ShouldBeNil)
			var pod v1.Pod
			err = convertToEdgeCoreMsg(tc.inputMsg, &pod)
			convey.So(err, convey.ShouldBeNil)

			convey.So(pod, convey.ShouldResemble, originPod)
		}
	})
}

type mefPodUpdateTestcase struct {
	description   string
	inputMsg      *model.Message
	realNpuName   string
	npuCapability bool
	expectedPod   v1.Pod
}

var trueValue = true
var defaultSecurityContext = &v1.SecurityContext{
	AllowPrivilegeEscalation: new(bool),
	Privileged:               new(bool),
	RunAsNonRoot:             &trueValue,
	ReadOnlyRootFilesystem:   &trueValue,
	Capabilities:             &v1.Capabilities{Drop: []v1.Capability{"All"}},
	SeccompProfile:           &v1.SeccompProfile{Type: v1.SeccompProfileTypeRuntimeDefault},
}

var mefPodUpdateTestcases = []mefPodUpdateTestcase{
	{
		description: "test pod update without npu",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptUpdate, Resource: constants.ResMefPodPrefix + "mef-pod-update-1"},
			v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("2")}}}}}}),
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("2")}},
			SecurityContext: defaultSecurityContext}}}},
		realNpuName: ascend310p,
	},
	{
		description: "test pod update with npu without npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptUpdate, Resource: constants.ResMefPodPrefix + "mef-pod-update-2"},
			v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{constants.CenterNpuName: resource.MustParse("2")}}}}}}),
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{ascend310p: resource.MustParse("2")}},
			SecurityContext: defaultSecurityContext}}}},
		realNpuName: ascend310p,
	},
	{
		description: "test pod update with npu with npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgeControllerModule, Group: constants.ResourceModule,
				Operation: constants.OptUpdate, Resource: constants.ResMefPodPrefix + "mef-pod-update-3"},
			v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{constants.CenterNpuName: resource.MustParse("2")}}}}}}),
		expectedPod: v1.Pod{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{ascend310p: resource.MustParse("200")}},
			SecurityContext: defaultSecurityContext}}}},
		npuCapability: true,
		realNpuName:   ascend310p,
	},
}

func TestModifyPodUpdate(t *testing.T) {
	convey.Convey("update:mef-user/pod/", t, func() {
		var patches *gomonkey.Patches
		defer func() {
			if patches != nil {
				patches.Reset()
			}
		}()
		for _, tc := range mefPodUpdateTestcases {
			fmt.Println(tc.description)
			patches = gomonkey.ApplyFuncReturn(common.LoadNpuFromDb, tc.realNpuName, true)
			config.GetCapabilityCache().Set(constants.CapabilityNpuSharingConfig, tc.npuCapability)

			var actualPod v1.Pod
			err := convertToEdgeCoreMsg(tc.inputMsg, &actualPod)
			convey.So(err, convey.ShouldBeNil)
			convey.So(actualPod, convey.ShouldResemble, tc.expectedPod)

			patches.Reset()
			patches = nil
		}
	})
}
