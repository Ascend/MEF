// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package msgconv
package msgconv

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/database"
)

const (
	ascend310p = "huawei.com/Ascend310P"
)

type mefNodeInsertionTestcase struct {
	description   string
	inputMsg      *model.Message
	realNpuName   string
	npuCapability bool
	expectedNode  v1.Node
}

var mefNodeInsertionTestcases = []mefNodeInsertionTestcase{
	{
		description: "test node insertion without npu",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptInsert, Resource: constants.ActionDefaultNodeStatus + "mef-node-insertion-1"},
			v1.Node{Status: v1.NodeStatus{Allocatable: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("2")}}}),
		expectedNode: v1.Node{
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{v1.ResourceCPU: resource.MustParse("2")}}},
		realNpuName: ascend310p,
	},
	{
		description: "test node insertion with npu without npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptInsert, Resource: constants.ActionDefaultNodeStatus + "mef-node-insertion-2"},
			v1.Node{Status: v1.NodeStatus{Allocatable: v1.ResourceList{
				ascend310p: resource.MustParse("2")}}}),
		expectedNode: v1.Node{
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{constants.CenterNpuName: resource.MustParse("2")}}},
		realNpuName: ascend310p,
	},
	{
		description: "test node insertion with npu with npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptInsert, Resource: constants.ActionDefaultNodeStatus + "mef-node-insertion-3"},
			v1.Node{Status: v1.NodeStatus{Allocatable: v1.ResourceList{
				ascend310p: resource.MustParse("2")}}}),
		expectedNode: v1.Node{
			Status: v1.NodeStatus{Allocatable: v1.ResourceList{constants.CenterNpuName: resource.MustParse("2")}}},
		npuCapability: true,
		realNpuName:   ascend310p,
	},
}

func TestModifyNodeInsertionForMef(t *testing.T) {
	convey.Convey("mef:insert:default/node/", t, func() {
		var patches *gomonkey.Patches
		defer func() {
			if patches != nil {
				patches.Reset()
			}
		}()
		for _, tc := range mefNodeInsertionTestcases {
			fmt.Println(tc.description)
			patches = gomonkey.ApplyFuncReturn(common.LoadNpuFromDb, tc.realNpuName, true)
			config.GetCapabilityCache().Set(constants.CapabilityNpuSharingConfig, tc.npuCapability)

			var actualNode v1.Node
			err := convertToCloudcoreMsg(tc.inputMsg, &actualNode)
			convey.So(err, convey.ShouldBeNil)
			convey.So(actualNode, convey.ShouldResemble, tc.expectedNode)

			patches.Reset()
			patches = nil
		}
	})
}

type mefNodePatchTestcase struct {
	description   string
	inputMsg      *model.Message
	realNpuName   string
	npuCapability bool
	expectedPatch map[string]interface{}
}

var mefNodePatchTestcases = []mefNodePatchTestcase{
	{
		description: "test node patch without npu",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptPatch, Resource: constants.ActionDefaultNodePatch + "mef-node-patch-1"},
			`{"status": {"allocatable": {"cpu": "1", "npu": null}}}`),
		expectedPatch: map[string]interface{}{"status": map[string]interface{}{"allocatable": map[string]interface{}{
			"cpu": "1", "npu": nil}}},
		realNpuName: ascend310p,
	},
	{
		description: "test node patch with npu without npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptPatch, Resource: constants.ActionDefaultNodePatch + "mef-node-patch-2"},
			fmt.Sprintf(`{"status": {"allocatable": {"cpu": "1", "%s": "2"}}}`, ascend310p)),
		expectedPatch: map[string]interface{}{"status": map[string]interface{}{"allocatable": map[string]interface{}{
			"cpu": "1", constants.CenterNpuName: "2"}}},
		realNpuName: ascend310p,
	},
	{
		description: "test node patch with npu with npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptPatch, Resource: constants.ActionDefaultNodePatch + "mef-node-patch-3"},
			fmt.Sprintf(`{"status": {"allocatable": {"cpu": "1", "%s": "2"}}}`, ascend310p)),
		expectedPatch: map[string]interface{}{"status": map[string]interface{}{"allocatable": map[string]interface{}{
			"cpu": "1", constants.CenterNpuName: "2"}}},
		npuCapability: true,
		realNpuName:   ascend310p,
	},
}

func TestModifyNodePatchForMef(t *testing.T) {
	convey.Convey("mef:patch:default/nodepatch/", t, func() {
		var patches *gomonkey.Patches
		defer func() {
			if patches != nil {
				patches.Reset()
			}
		}()
		for _, tc := range mefNodePatchTestcases {
			fmt.Println(tc.description)
			patches = gomonkey.ApplyFuncReturn(common.LoadNpuFromDb, tc.realNpuName, true)
			err := ensureNodeExists(filepath.Base(tc.inputMsg.KubeEdgeRouter.Resource), v1.Node{})
			convey.So(err, convey.ShouldBeNil)
			config.GetCapabilityCache().Set(constants.CapabilityNpuSharingConfig, tc.npuCapability)

			var actualPatch interface{}
			convey.So(convertToCloudcoreMsg(tc.inputMsg, &actualPatch), convey.ShouldBeNil)
			convey.So(actualPatch, convey.ShouldResemble, tc.expectedPatch)

			patches.Reset()
			patches = nil
		}
	})
}

type mefPodPatchTestcase struct {
	description   string
	inputMsg      *model.Message
	realNpuName   string
	npuCapability bool
	expectedPatch map[string]interface{}
}

var mefPodPatchTestcases = []mefPodPatchTestcase{
	{
		description: "test pod patch without npu",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptPatch, Resource: constants.ResMefPodPatchPrefix + "mef-node-patch-1"},
			`{"spec": {"containers": [{"resources": {"limits" :{"cpu": "1", "npu": null}}}]}}`),
		expectedPatch: map[string]interface{}{
			"spec": map[string]interface{}{"containers": []interface{}{map[string]interface{}{
				"resources": map[string]interface{}{"limits": map[string]interface{}{"cpu": "1", "npu": nil}}}}}},
		realNpuName: ascend310p,
	},
	{
		description: "test pod patch with npu without npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptPatch, Resource: constants.ResMefPodPatchPrefix + "mef-node-patch-2"},
			fmt.Sprintf(`{"spec": {"containers": [{"resources": {"limits" :{"cpu": "1", "%s": "2"}}}]}}`, ascend310p)),
		expectedPatch: map[string]interface{}{
			"spec": map[string]interface{}{"containers": []interface{}{map[string]interface{}{
				"resources": map[string]interface{}{"limits": map[string]interface{}{"cpu": "1", constants.CenterNpuName: "2"}}}}}},
		realNpuName: ascend310p,
	},
	{
		description: "test pod patch with npu with npu capability",
		inputMsg: mustCreateMsg(messageHeader{},
			model.MessageRoute{Source: constants.EdgedModule, Group: constants.MetaModule,
				Operation: constants.OptPatch, Resource: constants.ResMefPodPatchPrefix + "mef-node-patch-3"},
			fmt.Sprintf(`{"spec": {"containers": [{"resources": {"limits" :{"cpu": "1", "%s": "2"}}}]}}`, ascend310p)),
		expectedPatch: map[string]interface{}{
			"spec": map[string]interface{}{"containers": []interface{}{map[string]interface{}{
				"resources": map[string]interface{}{"limits": map[string]interface{}{"cpu": "1", constants.CenterNpuName: "2"}}}}}},
		npuCapability: true,
		realNpuName:   ascend310p,
	},
}

func TestModifyPodPatchForMef(t *testing.T) {
	convey.Convey("mef:patch:mef-user/nodepatch/", t, func() {
		var patches *gomonkey.Patches
		defer func() {
			if patches != nil {
				patches.Reset()
			}
		}()
		for _, tc := range mefPodPatchTestcases {
			fmt.Println(tc.description)
			patches = gomonkey.ApplyFuncReturn(common.LoadNpuFromDb, tc.realNpuName, true)
			config.GetCapabilityCache().Set(constants.CapabilityNpuSharingConfig, tc.npuCapability)

			dataBytes, err := json.Marshal(v1.Pod{})
			convey.So(err, convey.ShouldBeNil)
			stmt := test.MockGetDb().Create(database.Meta{
				Key: strings.ReplaceAll(tc.inputMsg.KubeEdgeRouter.Resource,
					constants.ResMefPodPatchPrefix, constants.ResMefPodPrefix),
				Type:  constants.ResourceTypePod,
				Value: string(dataBytes),
			})
			convey.So(stmt.Error, convey.ShouldBeNil)

			var actualPatch interface{}
			err = convertToCloudcoreMsg(tc.inputMsg, &actualPatch)
			convey.So(err, convey.ShouldBeNil)
			convey.So(actualPatch, convey.ShouldResemble, tc.expectedPatch)

			patches.Reset()
			patches = nil
		}
	})
}
