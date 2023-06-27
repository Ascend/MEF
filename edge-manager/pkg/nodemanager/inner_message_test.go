// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package nodemanager for inner message test
package nodemanager

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
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
	res := innerGetNodeInfoByUniqueName(input)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testInnerGetNodeInfoByUniqueNameErr() {
	convey.Convey("input error", func() {
		input := ""
		res := innerGetNodeInfoByUniqueName(input)
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("get node by id failed", func() {
		input := types.InnerGetNodeInfoByNameReq{UniqueName: ""}
		res := innerGetNodeInfoByUniqueName(input)
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
	res := innerGetNodeSoftwareInfo(input)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testInnerGetNodeSoftwareInfoErr() {
	convey.Convey("input error", func() {
		input := ""
		res := innerGetNodeSoftwareInfo(input)
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("get info failed", func() {
		input := types.InnerGetSfwInfoBySNReq{SerialNumber: "error sn"}
		res := innerGetNodeSoftwareInfo(input)
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
		res := innerGetNodeSoftwareInfo(input)
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
	res := innerGetNodeStatus(input)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testInnerGetNodeStatusErr() {
	convey.Convey("input error", func() {
		input := ""
		res := innerGetNodeStatus(input)
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("get node status error", func() {
		var c *nodeSyncImpl
		var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetNodeStatus",
			func(s *nodeSyncImpl, hostname string) (string, error) {
				return statusOffline, testErr
			})
		defer p1.Reset()
		input := types.InnerGetNodeStatusReq{UniqueName: ""}
		res := innerGetNodeStatus(input)
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
	res := innerGetNodeGroupInfosByIds(input)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testInnerGetNodeGroupInfosByIdsErr() {
	convey.Convey("input error", func() {
		input := ""
		res := innerGetNodeGroupInfosByIds(input)
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("id does not exist", func() {
		input := types.InnerGetNodeGroupInfosReq{NodeGroupIds: []uint64{0}}
		res := innerGetNodeGroupInfosByIds(input)
		convey.So(res.Msg, convey.ShouldEqual, fmt.Sprintf("get node group info, id %v do no exist", 0))
	})

	convey.Convey("get node group by id error", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getNodeGroupByID",
			func(n *NodeServiceImpl, groupID uint64) (*NodeGroup, error) {
				return nil, testErr
			})
		defer p1.Reset()
		input := types.InnerGetNodeGroupInfosReq{NodeGroupIds: []uint64{0}}
		res := innerGetNodeGroupInfosByIds(input)
		convey.So(res.Msg, convey.ShouldEqual, fmt.Sprintf("get node group info id %v, db failed", 0))
	})
}

func TestInnerAllNodeInfos(t *testing.T) {
	convey.Convey("innerAllNodeInfos should be success", t, func() {
		res := innerAllNodeInfos("")
		convey.So(res.Status, convey.ShouldEqual, common.Success)
	})

	convey.Convey("innerAllNodeInfos should be failed", t, func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listNodes",
			func(n *NodeServiceImpl) (*[]NodeInfo, error) {
				return nil, testErr
			})
		defer p1.Reset()
		res := innerAllNodeInfos("")
		convey.So(res.Status, convey.ShouldEqual, "")
	})
}

func TestInnerCheckNodeGroupResReq(t *testing.T) {
	convey.Convey("innerCheckNodeGroupResReq and innerCheckNodeGroupResReq should be success", t, testNodeGroupResReq)
	convey.Convey("innerCheckNodeGroupResReq and innerCheckNodeGroupResReq should be failed", t, testNodeGroupResReqErr)
}

func testNodeGroupResReq() {
	input := types.InnerCheckNodeResReq{NodeGroupID: 1}
	res := innerCheckNodeGroupResReq(input)
	convey.So(res.Status, convey.ShouldEqual, common.Success)

	input2 := types.InnerGetNodesReq{NodeGroupID: 1}
	res2 := innerGetNodesByNodeGroupID(input2)
	convey.So(res2.Status, convey.ShouldEqual, common.Success)
}

func testNodeGroupResReqErr() {
	convey.Convey("input error", func() {
		res := innerCheckNodeGroupResReq("")
		convey.So(res.Msg, convey.ShouldEqual, "parse inner message content failed")

		res2 := innerCheckNodeGroupResReq("")
		convey.So(res2.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("check error", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listNodeRelationsByGroupId",
			func(n *NodeServiceImpl, groupId uint64) (*[]NodeRelation, error) {
				return nil, testErr
			})
		defer p1.Reset()
		input := types.InnerCheckNodeResReq{NodeGroupID: 1}
		res := innerCheckNodeGroupResReq(input)
		convey.So(res.Msg, convey.ShouldEqual, fmt.Sprintf("get node relations by group id [%d] error", 1))

		input2 := types.InnerGetNodesReq{NodeGroupID: 1}
		res2 := innerGetNodesByNodeGroupID(input2)
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
	resp := innerUpdateNodeGroupResReq(innerUpdateNodeResReq)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testInnerUpdateNodeGroupResReqErr() {
	convey.Convey("input error", func() {
		resp := innerUpdateNodeGroupResReq("")
		convey.So(resp.Msg, convey.ShouldEqual, "parse inner message content failed")
	})

	convey.Convey("update error", func() {
		innerUpdateNodeResReq := types.InnerUpdateNodeResReq{
			NodeGroupID:  10,
			ResourceReqs: nil,
			IsUndeploy:   false,
		}
		resp := innerUpdateNodeGroupResReq(innerUpdateNodeResReq)
		convey.So(resp.Msg, convey.ShouldEqual, fmt.Sprintf("get node group id [%d] resources request failed, "+
			"db error", innerUpdateNodeResReq.NodeGroupID))
	})
}
