// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nodemanager for inner message test
package nodemanager

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"

	"edge-manager/pkg/types"
)

const (
	totalGroupNodeCount = 10
)

func TestInnerGetNodeInfoByUniqueName(t *testing.T) {
	convey.Convey("innerGetNodeInfoByUniqueName should be success", t, testInnerGetNodeInfoByUniqueName)
	convey.Convey("innerGetNodeInfoByUniqueName should be failed", t, testInnerGetNodeInfoByUniqueNameErr)
}

func testInnerGetNodeInfoByUniqueName() {
	node := &NodeInfo{
		Description:  "test-node-description-15",
		NodeName:     "test-node-name-15",
		UniqueName:   "test-node-unique-name-15",
		SerialNumber: "test-node-serial-number-15",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	resNode := env.createNode(node)
	convey.So(resNode, convey.ShouldBeNil)
	input := types.InnerGetNodeInfoByNameReq{UniqueName: node.UniqueName}
	msg := model.Message{}
	err := msg.FillContent(input)
	convey.So(err, convey.ShouldBeNil)
	res := innerGetNodeInfoByUniqueName(&msg)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testInnerGetNodeInfoByUniqueNameErr() {
	convey.Convey("input error", func() {
		res := innerGetNodeInfoByUniqueName(&model.Message{Content: []byte("")})
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("get node by id failed", func() {
		input := types.InnerGetNodeInfoByNameReq{UniqueName: ""}
		msg := model.Message{}
		err := msg.FillContent(input)
		convey.So(err, convey.ShouldBeNil)
		res := innerGetNodeInfoByUniqueName(&msg)
		convey.So(res.Status, convey.ShouldNotEqual, common.Success)
	})
}

func TestInnerGetNodeSoftwareInfo(t *testing.T) {
	convey.Convey("innerGetNodeSoftwareInfo should be success", t, testInnerGetNodeSoftwareInfo)
	convey.Convey("innerGetNodeSoftwareInfo should be failed", t, testInnerGetNodeSoftwareInfoErr)
}

func testInnerGetNodeSoftwareInfo() {
	sfwInfo := []types.SoftwareInfo{
		{
			Name:            "edgecore",
			Version:         "v1.12",
			InactiveVersion: "v1.12",
		}}

	sfwInfoRespByte, err := json.Marshal(sfwInfo)
	if err != nil {
		fmt.Printf("marshal software info resp failed, error: %v", err)
	}

	node := &NodeInfo{
		Description:  "test-node-description-16",
		NodeName:     "test-node-name-16",
		UniqueName:   "test-node-unique-name-16",
		SerialNumber: "test-node-serial-number-16",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
		SoftwareInfo: string(sfwInfoRespByte),
	}
	resNode := env.createNode(node)
	convey.So(resNode, convey.ShouldBeNil)

	input := types.InnerGetSfwInfoBySNReq{SerialNumber: node.SerialNumber}
	msg := model.Message{}
	err = msg.FillContent(input)
	convey.So(err, convey.ShouldBeNil)
	res := innerGetNodeSoftwareInfo(&msg)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testInnerGetNodeSoftwareInfoErr() {
	convey.Convey("input error", func() {
		res := innerGetNodeSoftwareInfo(&model.Message{Content: []byte("")})
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("get info failed", func() {
		input := types.InnerGetSfwInfoBySNReq{SerialNumber: "error sn"}
		msg := model.Message{}
		err := msg.FillContent(input)
		convey.So(err, convey.ShouldBeNil)
		res := innerGetNodeSoftwareInfo(&msg)
		convey.So(res.Msg, convey.ShouldEqual, "get node info by unique name failed")
	})

	convey.Convey("unmarshal failed", func() {
		node := &NodeInfo{
			Description:  "test-node-description-17",
			NodeName:     "test-node-name-17",
			UniqueName:   "test-node-unique-name-17",
			SerialNumber: "test-node-serial-number-17",
			IP:           "0.0.0.0",
			IsManaged:    true,
			CreatedAt:    time.Now().Format(TimeFormat),
			UpdatedAt:    time.Now().Format(TimeFormat),
			SoftwareInfo: "",
		}
		resNode := env.createNode(node)
		convey.So(resNode, convey.ShouldBeNil)
		input := types.InnerGetSfwInfoBySNReq{SerialNumber: node.SerialNumber}
		msg := model.Message{}
		err := msg.FillContent(input)
		convey.So(err, convey.ShouldBeNil)
		res := innerGetNodeSoftwareInfo(&msg)
		convey.So(res.Msg, convey.ShouldEqual, "get node info failed because unmarshal failed")
	})
}

func TestInnerGetNodeStatus(t *testing.T) {
	convey.Convey("innerGetNodeStatus should be success", t, testInnerGetNodeStatus)
	convey.Convey("innerGetNodeStatus should be failed", t, testInnerGetNodeStatusErr)
}

func testInnerGetNodeStatus() {
	node := &NodeInfo{
		Description:  "test-node-description-18",
		NodeName:     "test-node-name-18",
		UniqueName:   "test-node-unique-name-18",
		SerialNumber: "test-node-serial-number-18",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	resNode := env.createNode(node)
	convey.So(resNode, convey.ShouldBeNil)
	input := types.InnerGetNodeStatusReq{UniqueName: node.UniqueName}
	msg := model.Message{}
	err := msg.FillContent(input)
	convey.So(err, convey.ShouldBeNil)
	res := innerGetNodeStatus(&msg)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testInnerGetNodeStatusErr() {
	convey.Convey("input error", func() {
		res := innerGetNodeStatus(&model.Message{Content: []byte("")})
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("get node status error", func() {
		var c *nodeSyncImpl
		var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetK8sNodeStatus",
			func(s *nodeSyncImpl, hostname string) (string, error) {
				return statusOffline, test.ErrTest
			})
		defer p1.Reset()
		input := types.InnerGetNodeStatusReq{UniqueName: ""}
		msg := model.Message{}
		err := msg.FillContent(input)
		convey.So(err, convey.ShouldBeNil)
		res := innerGetNodeStatus(&msg)
		convey.So(res.Msg, convey.ShouldEqual, "internal get node status failed")
	})
}

func TestInnerGetNodeGroupInfosByIds(t *testing.T) {
	convey.Convey("innerGetNodeGroupInfosByIds should be success", t, testInnerGetNodeGroupInfosByIds)
	convey.Convey("innerGetNodeGroupInfosByIds should be failed", t, testInnerGetNodeGroupInfosByIdsErr)
}

func testInnerGetNodeGroupInfosByIds() {
	group := &NodeGroup{
		GroupName: "test_group_name_15",
		CreatedAt: time.Now().Format(TimeFormat),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	resGroup := env.createGroup(group)
	convey.So(resGroup, convey.ShouldBeNil)
	input := types.InnerGetNodeGroupInfosReq{NodeGroupIds: []uint64{group.ID}}
	msg := model.Message{}
	err := msg.FillContent(input)
	convey.So(err, convey.ShouldBeNil)
	res := innerGetNodeGroupInfosByIds(&msg)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testInnerGetNodeGroupInfosByIdsErr() {
	convey.Convey("input error", func() {
		res := innerGetNodeGroupInfosByIds(&model.Message{Content: []byte("")})
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("id does not exist", func() {
		input := types.InnerGetNodeGroupInfosReq{NodeGroupIds: []uint64{0}}
		msg := model.Message{}
		err := msg.FillContent(input)
		convey.So(err, convey.ShouldBeNil)
		res := innerGetNodeGroupInfosByIds(&msg)
		convey.So(res.Msg, convey.ShouldEqual, fmt.Sprintf("get node group info, id %v do no exist", 0))
	})

	convey.Convey("get node group by id error", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getNodeGroupByID",
			func(n *NodeServiceImpl, groupID uint64) (*NodeGroup, error) {
				return nil, test.ErrTest
			})
		defer p1.Reset()
		input := types.InnerGetNodeGroupInfosReq{NodeGroupIds: []uint64{0}}
		msg := model.Message{}
		err := msg.FillContent(input)
		convey.So(err, convey.ShouldBeNil)
		res := innerGetNodeGroupInfosByIds(&msg)
		convey.So(res.Msg, convey.ShouldEqual, fmt.Sprintf("get node group info id %v, db failed", 0))
	})
}

func TestInnerAllNodeInfos(t *testing.T) {
	convey.Convey("innerAllNodeInfos should be success", t, func() {
		res := innerAllNodeInfos(&model.Message{Content: []byte("")})
		convey.So(res.Status, convey.ShouldEqual, common.Success)
	})

	convey.Convey("innerAllNodeInfos should be failed", t, func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listNodes",
			func(n *NodeServiceImpl) (*[]NodeInfo, error) {
				return nil, test.ErrTest
			})
		defer p1.Reset()
		res := innerAllNodeInfos(&model.Message{Content: []byte("")})
		convey.So(res.Status, convey.ShouldEqual, "")
	})
}

func TestInnerCheckNodeGroupResReq(t *testing.T) {
	convey.Convey("innerCheckNodeGroupResReq and innerCheckNodeGroupResReq should be success", t, testNodeGroupResReq)
	convey.Convey("innerCheckNodeGroupResReq and innerCheckNodeGroupResReq should be failed", t, testNodeGroupResReqErr)
}

func testNodeGroupResReq() {
	input := types.InnerCheckNodeResReq{NodeGroupID: 1}
	msg := model.Message{}
	err := msg.FillContent(input)
	convey.So(err, convey.ShouldBeNil)
	res := innerCheckNodeGroupResReq(&msg)
	convey.So(res.Status, convey.ShouldEqual, common.Success)

	input2 := types.InnerGetNodesReq{NodeGroupID: 1}
	err = msg.FillContent(input2)
	convey.So(err, convey.ShouldBeNil)
	res2 := innerGetNodesByNodeGroupID(&msg)
	convey.So(res2.Status, convey.ShouldEqual, common.Success)
}

func testNodeGroupResReqErr() {
	convey.Convey("input error", func() {
		res := innerCheckNodeGroupResReq(&model.Message{Content: []byte("")})
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")

		res2 := innerCheckNodeGroupResReq(&model.Message{Content: []byte("")})
		convey.So(res2.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("check error", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listNodeRelationsByGroupId",
			func(n *NodeServiceImpl, groupId uint64) (*[]NodeRelation, error) {
				return nil, test.ErrTest
			})
		defer p1.Reset()
		input := types.InnerCheckNodeResReq{NodeGroupID: 1}
		msg := model.Message{}
		err := msg.FillContent(input)
		convey.So(err, convey.ShouldBeNil)
		res := innerCheckNodeGroupResReq(&msg)
		convey.So(res.Status, convey.ShouldEqual, "")

		input2 := types.InnerGetNodesReq{NodeGroupID: 1}
		err = msg.FillContent(input2)
		convey.So(err, convey.ShouldBeNil)
		res2 := innerGetNodesByNodeGroupID(&msg)
		convey.So(res2.Msg, convey.ShouldEqual, "inner message get node status failed")
	})
}

func TestInnerUpdateNodeGroupResReq(t *testing.T) {
	convey.Convey("innerUpdateNodeGroupResReq should be success", t, testInnerUpdateNodeGroupResReq)
	convey.Convey("innerUpdateNodeGroupResReq should be failed", t, testInnerUpdateNodeGroupResReqErr)
}

func testInnerUpdateNodeGroupResReq() {
	group := &NodeGroup{
		Description: "test-group-description-16",
		GroupName:   "test_group_name_16",
	}
	convey.So(env.createGroup(group), convey.ShouldBeNil)

	innerUpdateNodeResReq := types.InnerUpdateNodeResReq{
		NodeGroupID:  group.ID,
		ResourceReqs: nil,
		IsUndeploy:   false,
	}
	msg := model.Message{}
	err := msg.FillContent(innerUpdateNodeResReq)
	convey.So(err, convey.ShouldBeNil)
	resp := innerUpdateNodeGroupResReq(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testInnerUpdateNodeGroupResReqErr() {
	convey.Convey("input error", func() {
		resp := innerUpdateNodeGroupResReq(&model.Message{Content: []byte("")})
		convey.So(resp.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("update error", func() {
		innerUpdateNodeResReq := types.InnerUpdateNodeResReq{
			NodeGroupID:  10,
			ResourceReqs: nil,
			IsUndeploy:   false,
		}
		msg := model.Message{}
		err := msg.FillContent(innerUpdateNodeResReq)
		convey.So(err, convey.ShouldBeNil)
		resp := innerUpdateNodeGroupResReq(&msg)
		convey.So(resp.Msg, convey.ShouldEqual, fmt.Sprintf("get node group id [%d] resources request failed, "+
			"db error", innerUpdateNodeResReq.NodeGroupID))
	})
}

func TestInnerGetNodeSnsByGroupId(t *testing.T) {
	groupName := "groupName"
	convey.Convey("test get node ", t, func() {
		patch1 := gomonkey.ApplyPrivateMethod(&NodeServiceImpl{}, "getNodeByID", func(nodeId uint64) (*NodeInfo, error) {
			return &NodeInfo{
				SerialNumber: "sn-" + strconv.Itoa(int(nodeId)),
			}, nil
		})
		defer patch1.Reset()
		nodegroupRes := CreateNodeGroupReq{
			Description:   "description",
			NodeGroupName: &groupName,
		}
		bytes, err := json.Marshal(nodegroupRes)
		msg := model.Message{}
		err = msg.FillContent(bytes)
		convey.So(err, convey.ShouldBeNil)
		resp := createNodeGroup(&msg)
		convey.So(resp.Status == common.Success, convey.ShouldBeTrue)
		testGroupId, ok := resp.Data.(uint64)
		convey.So(ok, convey.ShouldBeTrue)
		for i := 1; i <= totalGroupNodeCount; i++ {
			relation := NodeRelation{
				GroupID:   testGroupId,
				NodeID:    uint64(i),
				CreatedAt: time.Now().String(),
			}
			err := test.MockGetDb().Model(NodeRelation{}).Create(relation).Error
			convey.So(err, convey.ShouldBeNil)
		}
		reqInput := requests.GetSnsReq{GroupId: testGroupId}
		bytes, err = json.Marshal(reqInput)
		convey.So(err, convey.ShouldBeNil)
		err = msg.FillContent(bytes)
		convey.So(err, convey.ShouldBeNil)
		resp = innerGetNodeSnsByGroupId(&msg)
		convey.So(resp.Status == common.Success, convey.ShouldBeTrue)
		sns, ok := resp.Data.([]string)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(sns) == totalGroupNodeCount, convey.ShouldBeTrue)
		for i := 1; i <= totalGroupNodeCount; i++ {
			test.MockGetDb().Model(NodeRelation{}).Where("group_id = ?", testGroupId).Delete(&NodeRelation{})
			convey.So(err, convey.ShouldBeNil)
		}
	})
}

func TestInnerGetNodeSnAndIpByID(t *testing.T) {
	convey.Convey("test innerGetNodeSnAndIpByID", t, func() {
		patch := gomonkey.ApplyPrivateMethod(&NodeServiceImpl{}, "getNodeByID", func(nodeId uint64) (*NodeInfo, error) {
			return &NodeInfo{
				ID:           nodeId,
				UniqueName:   strconv.Itoa(int(nodeId)),
				SerialNumber: strconv.Itoa(int(nodeId)),
				IP:           "10.10.10.10",
			}, nil
		})
		defer patch.Reset()
		req := types.InnerGetNodeInfosReq{
			NodeIds: []uint64{1},
		}
		msg := model.Message{}
		err := msg.FillContent(req)
		convey.So(err, convey.ShouldBeNil)
		resp := innerGetNodeSnAndIpByID(&msg)
		convey.So(resp.Status == common.Success, convey.ShouldBeTrue)
		info, ok := resp.Data.(types.InnerGetNodeInfosResp)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(info.NodeInfos) == 1, convey.ShouldBeTrue)
	})
}
