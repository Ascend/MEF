// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package appmanager

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

func TestInnerMessage(t *testing.T) {
	convey.Convey("test getAppInstanceCountByNodeGroup ", t, testGetAppInstanceCountByNodeGroup)
	convey.Convey("test checkNodeGroupResource", t, testCheckNodeGroupResource)
	convey.Convey("test updateAllocatedNodeRes", t, testUpdateAllocatedNodeRes)
	convey.Convey("test getNodeGroupInfos", t, testGetNodeGroupInfos)
	convey.Convey("test getNodeInfoByUniqueName", t, testGetNodeInfoByUniqueName)
	convey.Convey("test getNodeStatus", t, testGetNodeStatus)
	convey.Convey("test getAppResReqs", t, testGetAppResReqs)
}

func testGetAppInstanceCountByNodeGroup() {
	message, err := model.NewMessage()
	if err != nil {
		panic(err)
	}
	respMsg := getAppInstanceCountByNodeGroup(message)
	convey.So(respMsg.Status, convey.ShouldEqual, common.ErrorParamConvert)
	convey.So(respMsg.Msg, convey.ShouldEqual, "parse content failed")

	if err := message.FillContent([]uint64{1}); err != nil {
		panic(err)
	}
	patches := gomonkey.ApplyPrivateMethod(&AppRepositoryImpl{}, "countDeployedAppByGroupID",
		func(uint64) (int64, error) { return 0, test.ErrTest })
	defer patches.Reset()
	respMsg = getAppInstanceCountByNodeGroup(message)
	convey.So(respMsg.Status, convey.ShouldEqual, common.ErrorGetAppInstanceCountByNodeGroup)
	convey.So(respMsg.Msg, convey.ShouldEqual, "")

	patches.Reset()
	patches = gomonkey.ApplyPrivateMethod(&AppRepositoryImpl{}, "countDeployedAppByGroupID",
		func(uint64) (int64, error) { return 1, nil })
	respMsg = getAppInstanceCountByNodeGroup(message)
	convey.So(respMsg.Status, convey.ShouldEqual, common.Success)
	convey.So(respMsg.Data, convey.ShouldResemble, map[uint64]int64{1: 1})
}

func testCheckNodeGroupResource() {
	set := &appv1.DaemonSet{}
	set.Spec.Template.Spec.Containers = []corev1.Container{
		{Resources: corev1.ResourceRequirements{Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("100Mi"),
		}}},
	}
	outputCell := []gomonkey.OutputCell{
		{Values: gomonkey.Params{common.RespMsg{Status: common.FAIL, Msg: test.ErrTest.Error()}}, Times: 1},
		{Values: gomonkey.Params{common.RespMsg{Status: common.Success}}, Times: 1},
	}
	patches := gomonkey.ApplyFuncSeq(common.SendSyncMessageByRestful, outputCell).
		ApplyFuncReturn(getAppResReqs, corev1.ResourceList{})
	defer patches.Reset()
	err := checkNodeGroupResource(0, set)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
	err = checkNodeGroupResource(0, set)
	convey.So(err, convey.ShouldBeNil)
}

func testUpdateAllocatedNodeRes() {
	outputCell := []gomonkey.OutputCell{
		{Values: gomonkey.Params{common.RespMsg{Status: common.FAIL, Msg: test.ErrTest.Error()}}, Times: 1},
		{Values: gomonkey.Params{common.RespMsg{Status: common.Success}}, Times: 1},
	}
	patches := gomonkey.ApplyFuncSeq(common.SendSyncMessageByRestful, outputCell).
		ApplyFuncReturn(getAppResReqs, corev1.ResourceList{})
	defer patches.Reset()
	err := updateAllocatedNodeRes(&appv1.DaemonSet{}, 0, false)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
	err = updateAllocatedNodeRes(&appv1.DaemonSet{}, 0, false)
	convey.So(err, convey.ShouldBeNil)
}

var (
	failedResp         = common.RespMsg{Status: common.ErrorParseBody, Msg: test.ErrTest.Error()}
	errorMarshalResp   = common.RespMsg{Status: common.Success, Data: func() {}}
	errorUnmarshalResp = common.RespMsg{Status: common.Success, Data: "test error type string"}
)

func testGetNodeGroupInfos() {
	testResp := types.InnerGetNodeGroupInfosResp{NodeGroupInfos: []types.NodeGroupInfo{{
		NodeGroupID:   1,
		NodeGroupName: "test-group-name",
	}}}
	outputCell := []gomonkey.OutputCell{
		{Values: gomonkey.Params{failedResp}, Times: 1},
		{Values: gomonkey.Params{common.RespMsg{Status: common.Success, Data: testResp}}, Times: 1},
	}
	patches := gomonkey.ApplyFuncSeq(common.SendSyncMessageByRestful, outputCell)
	defer patches.Reset()
	_, err := getNodeGroupInfos([]uint64{1})
	convey.So(err, convey.ShouldResemble, test.ErrTest)
	nodeGroupInfos, err := getNodeGroupInfos([]uint64{1})
	convey.So(err, convey.ShouldBeNil)
	convey.So(nodeGroupInfos, convey.ShouldResemble, testResp.NodeGroupInfos)
}

func testGetNodeInfoByUniqueName() {
	testPod := &corev1.Pod{}
	id, name, err := getNodeInfoByUniqueName(testPod)
	convey.So(err, convey.ShouldBeNil)
	convey.So(id, convey.ShouldEqual, 0)
	convey.So(name, convey.ShouldEqual, "")

	testPod.Spec.NodeName = "test-node-name"
	testResp := types.InnerGetNodeInfoByNameResp{
		NodeName: "test-name",
		NodeID:   1,
	}
	outputCell := []gomonkey.OutputCell{
		{Values: gomonkey.Params{errorMarshalResp}, Times: 1},
		{Values: gomonkey.Params{common.RespMsg{Status: common.Success, Data: testResp}}, Times: 1},
	}
	patches := gomonkey.ApplyFuncSeq(common.SendSyncMessageByRestful, outputCell)
	defer patches.Reset()
	_, _, err = getNodeInfoByUniqueName(testPod)
	convey.So(err, convey.ShouldResemble, errors.New("marshal internal response error"))
	id, name, err = getNodeInfoByUniqueName(testPod)
	convey.So(err, convey.ShouldBeNil)
	convey.So(id, convey.ShouldEqual, testResp.NodeID)
	convey.So(name, convey.ShouldEqual, testResp.NodeName)
}

func testGetNodeStatus() {
	status, err := getNodeStatus("")
	convey.So(status, convey.ShouldEqual, nodeStatusUnknown)
	convey.So(err, convey.ShouldBeNil)

	testResp := types.InnerGetNodeStatusResp{
		NodeStatus: nodeStatusReady,
	}
	outputCell := []gomonkey.OutputCell{
		{Values: gomonkey.Params{errorUnmarshalResp}, Times: 1},
		{Values: gomonkey.Params{common.RespMsg{Status: common.Success, Data: testResp}}, Times: 1},
	}
	patches := gomonkey.ApplyFuncSeq(common.SendSyncMessageByRestful, outputCell)
	defer patches.Reset()

	_, err = getNodeStatus("test-node-name")
	convey.So(err, convey.ShouldResemble, errors.New("unmarshal internal response error"))
	nodeStatus, err := getNodeStatus("test-node-name")
	convey.So(err, convey.ShouldBeNil)
	convey.So(nodeStatus, convey.ShouldEqual, testResp.NodeStatus)
}

func testGetAppResReqs() {
	ds := &appv1.DaemonSet{}
	container := corev1.Container{}
	container.Resources.Limits = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("200Mi"),
	}
	ds.Spec.Template.Spec.Containers = []corev1.Container{container, container}
	reqs := getAppResReqs(ds)
	exceptedCpu := resource.MustParse("2")
	exceptedMemory := resource.MustParse("400Mi")

	convey.So(reqs.Cpu().Value(), convey.ShouldResemble, exceptedCpu.Value())
	convey.So(reqs.Memory().Value(), convey.ShouldResemble, exceptedMemory.Value())
}
