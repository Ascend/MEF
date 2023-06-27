// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager test about node group
package nodemanager

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

func TestCreateNodeGroup(t *testing.T) {
	convey.Convey("createNodeGroup should be success", t, testCreateNodeGroup)
	convey.Convey("createNodeGroup should be failed, input error", t, testCreateNodeGroupErrInput)
	convey.Convey("createNodeGroup should be failed, param error", t, testCreateNodeGroupErrParam)
	convey.Convey("createNodeGroup should be failed, the table num has reached the maximum", t, testCreateGroupMaxCount)
	convey.Convey("createNodeGroup should be failed, group name is duplicate", t, testCreateGroupNameDuplicate)
	convey.Convey("createNodeGroup should be failed, create error", t, testCreateNodeGroupErrCreate)
}

func testCreateNodeGroup() {
	group := &NodeGroup{
		Description: "test-group-description-1",
		GroupName:   "test_group_name_1",
	}

	args := fmt.Sprintf(`
			{
  			"nodeGroupName": "%s",
  			"description": "%s"
			}`, group.GroupName, group.Description)
	resp := createNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testCreateNodeGroupErrInput() {
	resp := createNodeGroup("")
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testCreateNodeGroupErrParam() {
	group := &NodeGroup{
		Description: "test-group-description-2",
		GroupName:   "test_group_name_2",
	}

	convey.Convey("groupName is not exist", func() {
		args := fmt.Sprintf(`{"description": "%s"}`, group.Description)
		resp := createNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("description is not exist", func() {
		args := fmt.Sprintf(`{"nodeGroupName": "%s"}`, group.GroupName)
		resp := createNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func testCreateGroupMaxCount() {
	group := &NodeGroup{
		Description: "test-group-description-3",
		GroupName:   "test_group_name_3",
	}

	const maxTableCount = 1024
	var p1 = gomonkey.ApplyFunc(GetTableCount,
		func(tb interface{}) (int, error) {
			return maxTableCount, testErr
		})
	defer p1.Reset()
	args := fmt.Sprintf(`
			{
  			"nodeGroupName": "%s",
  			"description": "%s"
			}`, group.GroupName, group.Description)
	resp := createNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCheckNodeMrgSize)
}

func testCreateGroupNameDuplicate() {
	group := &NodeGroup{
		Description: "test-group-description-4",
		GroupName:   "test_group_name_4",
	}
	args := fmt.Sprintf(`{"nodeGroupName": "%s"}`, group.GroupName)
	_ = createNodeGroup(args)
	resp := createNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorNodeMrgDuplicate)
}

func testCreateNodeGroupErrCreate() {
	group := &NodeGroup{
		Description: "test-group-description-5",
		GroupName:   "test_group_name_5",
	}
	args := fmt.Sprintf(`{"nodeGroupName": "%s"}`, group.GroupName)

	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "createNodeGroup",
		func(n *NodeServiceImpl, nodeGroup *NodeGroup) error {
			return testErr
		})
	defer p1.Reset()
	resp := createNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCreateNodeGroup)
}

func TestGroupStatistics(t *testing.T) {
	convey.Convey("getNodeGroupStatistics should be success", t, testGetGroupStat)
	convey.Convey("getNodeGroupStatistics should be failed, get count error", t, testGetGroupStatErrGetCount)
}

func testGetGroupStat() {
	resp := getNodeGroupStatistics(``)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testGetGroupStatErrGetCount() {
	var p1 = gomonkey.ApplyFunc(GetTableCount,
		func(tb interface{}) (int, error) {
			return 0, testErr
		})
	defer p1.Reset()
	resp := getNodeGroupStatistics(``)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCountNodeGroup)
}

func TestGetGroupDetail(t *testing.T) {
	convey.Convey("getNodeGroupDetail should be success", t, testGetNodeGroupDetail)
	convey.Convey("getNodeGroupDetail should be failed, id is string", t, testGetNodeGroupDetailErrInput)
	convey.Convey("getNodeGroupDetail should be failed, param error", t, testGetNodeGroupDetailErrParam)
	convey.Convey("getNodeGroupDetail should be failed, get group error", t, testGetNodeGroupDetailErrGetGroup)
	convey.Convey("getNodeGroupDetail should be failed, list relations error", t, testGetDetailErrListRelations)
	convey.Convey("getNodeGroupDetail should be failed, get node by id error", t, testGetGroupDetailErrGetNodeById)
}

func testGetNodeGroupDetail() {
	group := &NodeGroup{
		GroupName: "test_group_name_6",
		CreatedAt: time.Now().Format(TimeFormat),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	convey.So(env.createGroup(group), convey.ShouldBeNil)
	node := &NodeInfo{
		Description:  "test-node-description-10",
		NodeName:     "test-node-name-10",
		UniqueName:   "test-node-unique-name-10",
		SerialNumber: "test-node-serial-number-10",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	convey.So(env.createNode(node), convey.ShouldBeNil)
	relation := &NodeRelation{
		NodeID:    node.ID,
		GroupID:   group.ID,
		CreatedAt: time.Now().Format(TimeFormat),
	}
	convey.So(env.createRelation(relation), convey.ShouldBeNil)

	resp := getNodeGroupDetail(group.ID)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	groupDetail, ok := resp.Data.(NodeGroupDetail)
	convey.So(ok, convey.ShouldBeTrue)
	convey.So(groupDetail.NodeGroup, convey.ShouldResemble, *group)
}

func testGetNodeGroupDetailErrInput() {
	args := `{"id": "1"}`
	resp := getNodeGroupDetail(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testGetNodeGroupDetailErrParam() {
	resp := getNodeGroupDetail(uint64(0))
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testGetNodeGroupDetailErrGetGroup() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getNodeGroupByID",
		func(n *NodeServiceImpl, groupID uint64) (*NodeGroup, error) {
			return nil, testErr
		})
	defer p1.Reset()
	resp := getNodeGroupDetail(uint64(1))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNodeGroup)
}

func testGetDetailErrListRelations() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listNodeRelationsByGroupId",
		func(n *NodeServiceImpl, groupID uint64) (*[]NodeRelation, error) {
			return nil, testErr
		})
	defer p1.Reset()
	resp := getNodeGroupDetail(uint64(1))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNodeGroup)
}

func testGetGroupDetailErrGetNodeById() {
	group := &NodeGroup{
		GroupName: "test_group_name_7",
		CreatedAt: time.Now().Format(TimeFormat),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	convey.So(env.createGroup(group), convey.ShouldBeNil)

	node := &NodeInfo{
		Description:  "test-node-description-11",
		NodeName:     "test-node-name-11",
		UniqueName:   "test-node-unique-name-11",
		SerialNumber: "test-node-serial-number-11",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	convey.So(env.createNode(node), convey.ShouldBeNil)

	relation := &NodeRelation{
		NodeID:    node.ID,
		GroupID:   group.ID,
		CreatedAt: time.Now().Format(TimeFormat),
	}
	convey.So(env.createRelation(relation), convey.ShouldBeNil)

	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getNodeByID",
		func(uint64) (*NodeInfo, error) {
			return nil, testErr
		})
	defer p1.Reset()
	resp := getNodeGroupDetail(group.ID)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNodeGroup)
}

func TestModifyGroup(t *testing.T) {
	convey.Convey("modifyGroup should be success", t, testModifyGroup)
	convey.Convey("modifyGroup should be failed, input error", t, testModifyGroupErrInput)
	convey.Convey("modifyGroup should be failed, param error", t, testModifyGroupErrParam)
	convey.Convey("modifyGroup should be failed, update error", t, testModifyGroupErrUpdate)
}

func testModifyGroup() {
	group := &NodeGroup{
		Description: "test-group-description-8",
		GroupName:   "test_group_name_8",
	}
	res := env.createGroup(group)
	convey.So(res, convey.ShouldBeNil)

	args := fmt.Sprintf(`
			{
				"groupID": %d,
				"nodeGroupName": "%s",
				"description": "%s"
			}`, group.ID, group.GroupName, group.Description)
	resp := modifyNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	verifyRes := env.verifyDbNodeGroup(group, "UpdatedAt")
	convey.So(verifyRes, convey.ShouldBeNil)
}

func testModifyGroupErrInput() {
	resp := modifyNodeGroup(``)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testModifyGroupErrParam() {
	args := `
{
	"nodeGroupName": "test_group_name_9",
	"description": "test-group-description-9"
}`
	resp := modifyNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testModifyGroupErrUpdate() {
	group := &NodeGroup{
		Description: "test-group-description-10",
		GroupName:   "test_group_name_10",
	}
	res := env.createGroup(group)
	convey.So(res, convey.ShouldBeNil)

	args := fmt.Sprintf(`
			{
				"groupID": %d,
				"nodeGroupName": "%s",
				"description": "%s"
			}`, group.ID, group.GroupName, group.Description)

	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "updateGroup",
		func(n *NodeServiceImpl, id uint64, columns map[string]interface{}) (int64, error) {
			return 0, testErr
		})
	defer p1.Reset()
	resp := modifyNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorModifyNodeGroup)
}

func TestBatchDeleteGroup(t *testing.T) {
	convey.Convey("batchDeleteNodeGroup should be success", t, testBatchDeleteNodeGroup)
	convey.Convey("batchDeleteNodeGroup should be failed", t, testBatchDeleteNodeGroupErr)
}

func testBatchDeleteNodeGroup() {
	group := &NodeGroup{
		Description: "test-group-description-11",
		GroupName:   "test_group_name_11",
	}
	res := env.createGroup(group)
	convey.So(res, convey.ShouldBeNil)

	var p1 = gomonkey.ApplyFunc(common.SendSyncMessageByRestful,
		func(input interface{}, router *common.Router, timeout time.Duration) common.RespMsg {
			var rsp common.RespMsg
			counts := map[uint64]int64{group.ID: 0}
			rsp.Status = common.Success
			rsp.Data = counts
			return rsp
		})
	defer p1.Reset()

	args := fmt.Sprintf(`{"groupIDs": [%d]}`, group.ID)
	resp := batchDeleteNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	verifyRes := env.verifyDbNodeGroup(group)
	convey.So(verifyRes, convey.ShouldEqual, gorm.ErrRecordNotFound)
}

func testBatchDeleteNodeGroupErr() {
	convey.Convey("bad id type", func() {
		args := `{"groupIDs": ["1"]}`
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})
	convey.Convey("duplicate id", func() {
		args := `{"groupIDs": [1, 1]}`
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})
	convey.Convey("empty list", func() {
		args := `{"groupIDs": []}`
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})
	convey.Convey("group id is not exist", func() {
		args := `{"groupIDs": [20]}`
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteNodeGroup)
	})
}

func TestListNodeGroup(t *testing.T) {
	convey.Convey("listNodeGroup should be success", t, testListNodeGroup)
	convey.Convey("listNodeGroup should be failed, input error", t, testListNodeGroupErrInput)
	convey.Convey("listNodeGroup should be failed, param error", t, testListNodeGroupErrParam)
	convey.Convey("listNodeGroup should be failed, count group error", t, testListNodeGroupErrCountGroup)
	convey.Convey("listNodeGroup should be failed, get group error", t, testListNodeGroupErrGetGroup)
	convey.Convey("listNodeGroup should be failed, list relations error", t, testListNodeGroupErrListRelations)
}

func testListNodeGroup() {
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	resp := listNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testListNodeGroupErrInput() {
	resp := listNodeGroup("")
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testListNodeGroupErrParam() {
	args := types.ListReq{PageNum: 1, PageSize: errPageSize}
	resp := listNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testListNodeGroupErrCountGroup() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "countNodeGroupsByName",
		func(n *NodeServiceImpl, nodeGroup string) (int64, error) {
			return 0, testErr
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	resp := listNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNodeGroups)
}

func testListNodeGroupErrGetGroup() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getNodeGroupsByName",
		func(n *NodeServiceImpl, pageNum, pageSize uint64, nodeGroup string) (*[]NodeGroup, error) {
			return nil, testErr
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	resp := listNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNodeGroups)
}

func testListNodeGroupErrListRelations() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listNodeRelationsByGroupId",
		func(n *NodeServiceImpl, groupID uint64) (*[]NodeRelation, error) {
			return nil, testErr
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	resp := listNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNodeGroups)
}

func TestAddNodeRelation(t *testing.T) {
	convey.Convey("addNodeRelation should be success", t, testAddNodeRelation)
	convey.Convey("addNodeRelation should be failed, input error", t, testAddNodeRelationErrInput)
	convey.Convey("addNodeRelation should be failed, param error", t, testAddNodeRelationErrParam)
	convey.Convey("addNodeRelation should be failed, add error", t, testAddNodeRelationErrAdd)
}

func testAddNodeRelation() {
	node := &NodeInfo{
		Description:  "test-node-description-12",
		NodeName:     "test-node-name-12",
		UniqueName:   "test-node-unique-name-12",
		SerialNumber: "test-node-serial-number-12",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		Description: "test-group-description-12",
		GroupName:   "test_group_name_12",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	resNode := env.createNode(node)
	convey.So(resNode, convey.ShouldBeNil)
	resGroup := env.createGroup(group)
	convey.So(resGroup, convey.ShouldBeNil)

	args := fmt.Sprintf(`{"groupID": %d, "nodeIDs": [%d]}`, group.ID, node.ID)
	resp := addNodeRelation(args)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	relation := &NodeRelation{NodeID: node.ID, GroupID: group.ID}
	verifyRes := env.verifyDbNodeRelation(relation, "CreatedAt")
	convey.So(verifyRes, convey.ShouldBeNil)
}

func testAddNodeRelationErrInput() {
	resp := addNodeRelation(``)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testAddNodeRelationErrParam() {
	args := `{"nodeIDs": [1]}`
	resp := addNodeRelation(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testAddNodeRelationErrAdd() {
	convey.Convey("group id is not exist", func() {
		args := `{"groupID": 20, "nodeIDs": [1]}`
		resp := addNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorAddNodeToGroup)
	})
	convey.Convey("nodeIDs is not exist", func() {
		args := `{"groupID": 1, "nodeIDs": [50]}`
		resp := addNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorAddNodeToGroup)
	})
	convey.Convey("get available resource error", func() {
		var c *nodeSyncImpl
		var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "GetAvailableResource",
			func(n *nodeSyncImpl, hostname string) (*NodeResource, error) {
				return nil, testErr
			})
		defer p1.Reset()
		args := `{"groupID": 1, "nodeIDs": [1]}`
		resp := addNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorAddNodeToGroup)
	})
	convey.Convey("get managed node by id error", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getManagedNodeByID",
			func(n *NodeServiceImpl, nodeID uint64) (*NodeInfo, error) {
				return nil, testErr
			})
		defer p1.Reset()
		args := `{"groupID": 1, "nodeIDs": [1]}`
		resp := addNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorAddNodeToGroup)
	})
	convey.Convey("add node to group error", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "addNodeToGroup",
			func(n *NodeServiceImpl, relation *NodeRelation, uniqueName string) error {
				return testErr
			})
		defer p1.Reset()
		args := `{"groupID": 1, "nodeIDs": [1]}`
		resp := addNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorAddNodeToGroup)
	})
}

func TestDeleteNodeFromGroup(t *testing.T) {
	convey.Convey("deleteNodeFromGroup should be success", t, testDeleteNodeFromGroup)
	convey.Convey("deleteNodeFromGroup should be failed", t, testDeleteNodeFromGroupErr)
}

func testDeleteNodeFromGroup() {
	node := &NodeInfo{
		Description:  "test-node-description-13",
		NodeName:     "test-node-name-13",
		UniqueName:   "test-node-unique-name-13",
		SerialNumber: "test-node-serial-number-13",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		GroupName: "test_group_name_13",
		CreatedAt: time.Now().Format(TimeFormat),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	resNode := env.createNode(node)
	convey.So(resNode, convey.ShouldBeNil)
	resGroup := env.createGroup(group)
	convey.So(resGroup, convey.ShouldBeNil)
	relation := &NodeRelation{
		NodeID:    node.ID,
		GroupID:   group.ID,
		CreatedAt: time.Now().Format(TimeFormat),
	}
	resRelation := env.createRelation(relation)
	convey.So(resRelation, convey.ShouldBeNil)

	args := fmt.Sprintf(`{
               "nodeIDs": [%d],
               "groupID": %d
           }`, node.ID, group.ID)
	resp := deleteNodeFromGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDeleteNodeFromGroupErr() {
	convey.Convey("bad input type", func() {
		args := `{"groupIDs": ["1"]}`
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})
	convey.Convey("empty nodeIDs", func() {
		args := `{"groupID": 1, "nodeIDs": []}`
		resp := deleteNodeFromGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})
	convey.Convey("group id is not exist", func() {
		args := `{"groupID": 1, "nodeIDs": [100]}`
		resp := deleteNodeFromGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteNodeFromGroup)
	})
}

func TestBatchDeleteNodeRelation(t *testing.T) {
	convey.Convey("batchDeleteNodeRelation should be success", t, testBatchDeleteNodeRelation)
	convey.Convey("batchDeleteNodeRelation should be failed", t, testBatchDeleteNodeRelationErr)
}

func testBatchDeleteNodeRelation() {
	node := &NodeInfo{
		Description:  "test-node-description-14",
		NodeName:     "test-node-name-14",
		UniqueName:   "test-node-unique-name-14",
		SerialNumber: "test-node-serial-number-14",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	resNode := env.createNode(node)
	convey.So(resNode, convey.ShouldBeNil)

	group := &NodeGroup{
		Description: "test-group-description-14",
		GroupName:   "test_group_name_14",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	resGroup := env.createGroup(group)
	convey.So(resGroup, convey.ShouldBeNil)

	relation := &NodeRelation{
		NodeID:    node.ID,
		GroupID:   group.ID,
		CreatedAt: time.Now().Format(TimeFormat),
	}
	resRelation := env.createRelation(relation)
	convey.So(resRelation, convey.ShouldBeNil)

	args := fmt.Sprintf(`[
            {
                "nodeID": %d,
                "groupID": %d
            }]`, node.ID, group.ID)
	resp := batchDeleteNodeRelation(args)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	verifyRes := env.verifyDbNodeRelation(relation)
	convey.So(verifyRes, convey.ShouldEqual, gorm.ErrRecordNotFound)
}

func testBatchDeleteNodeRelationErr() {
	convey.Convey("input error", func() {
		resp := batchDeleteNodeRelation(``)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("nodeID is not exist", func() {
		args := `[{"groupID": 1}]`
		resp := batchDeleteNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("duplicate relations", func() {
		args := `[{"groupID":1, "nodeID":2},{"nodeID":2,"groupID":1}]`
		resp := batchDeleteNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("delete error", func() {
		args := `[{"groupID":1, "nodeID":2}]`
		resp := batchDeleteNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteNodeFromGroup)
	})
}
