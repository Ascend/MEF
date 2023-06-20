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

func TestCreateGroup(t *testing.T) {
	convey.Convey("createNodeGroup should be success", t, testCreateNodeGroup)
	convey.Convey("createNodeGroup should be failed, input error", t, testCreateNodeGroupErrInput)
	convey.Convey("createNodeGroup should be failed, param error", t, testCreateNodeGroupErrParam)
	convey.Convey("createNodeGroup should be failed, the table num has reached the maximum", t, testCreateGroupMaxCount)
	convey.Convey("createNodeGroup should be failed, group name is duplicate", t, testCreateGroupNameDuplicate)
	convey.Convey("createNodeGroup should be failed, create error", t, testCreateNodeGroupErrCreate)
}

func testCreateNodeGroup() {
	group := &NodeGroup{
		Description: "test-create-group-1-description",
		GroupName:   "test_create_group_1_name",
	}

	args := fmt.Sprintf(`
			{
  			"nodeGroupName": "%s",
  			"description": "%s"
			}`, group.GroupName, group.Description)
	resp := createNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	verifyRes := env.verifyDbNodeGroup(group, "ID", "UpdatedAt", "CreatedAt")
	convey.So(verifyRes, convey.ShouldBeNil)
}

func testCreateNodeGroupErrInput() {
	resp := createNodeGroup("")
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testCreateNodeGroupErrParam() {
	group := &NodeGroup{
		Description: "test-create-group-2-description",
		GroupName:   "test_create_group_2_name",
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
		Description: "test-create-group-3-description",
		GroupName:   "test_create_group_3_name",
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
		Description: "test-create-group-4-description",
		GroupName:   "test_create_group_4_name",
	}
	args := fmt.Sprintf(`{"nodeGroupName": "%s"}`, group.GroupName)
	_ = createNodeGroup(args)
	resp := createNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorNodeMrgDuplicate)
}

func testCreateNodeGroupErrCreate() {
	group := &NodeGroup{
		Description: "test-create-group-5-description",
		GroupName:   "test_create_group_5_name",
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
		GroupName: "test_get_group_detail_1_name",
		CreatedAt: time.Now().Format(TimeFormat),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	convey.So(env.createGroup(group), convey.ShouldBeNil)
	node := &NodeInfo{
		NodeName:   "test-get-group-detail-1-node-name",
		UniqueName: "test-get-group-detail-1-unique-name",
		IP:         "0.0.0.0",
		IsManaged:  true,
		CreatedAt:  time.Now().Format(TimeFormat),
		UpdatedAt:  time.Now().Format(TimeFormat),
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
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getNodeByID",
		func(uint64) (*NodeInfo, error) {
			return nil, testErr
		})
	defer p1.Reset()
	resp := getNodeGroupDetail(uint64(1))
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
		Description: "test-modify-group-1-description",
		GroupName:   "test_modify_group_1_name",
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
	"nodeGroupName": "test_modify_group_2_name",
	"description": "test-modify-group-2-description"
}`
	resp := modifyNodeGroup(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testModifyGroupErrUpdate() {
	group := &NodeGroup{
		Description: "test-modify-group-2-description",
		GroupName:   "test_modify_group_2_name",
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
		Description: "test-batch-delete-group-1-description",
		GroupName:   "test_batch_delete_group_1_name",
	}
	res := env.createGroup(group)
	convey.So(res, convey.ShouldBeNil)

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

func TestListEdgeNodeGroup(t *testing.T) {
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
	const errorPageSize = 200
	args := types.ListReq{PageNum: 1, PageSize: errorPageSize}
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
		Description:  "test-add-relation-1-description",
		NodeName:     "test-add-relation-1-name",
		UniqueName:   "test-add-relation-1-description-unique-name",
		SerialNumber: "test-add-relation-1-serial-number",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		Description: "test-add-relation-1-description",
		GroupName:   "test_add_relation_1_name",
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
		args := `{"groupID": 1, "nodeIDs": [20]}`
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
		NodeName:     "test-delete-node-from-group-1-name",
		UniqueName:   "test-delete-node-from-group-1-unique-name",
		SerialNumber: "test-delete-node-from-group-1-serial-number",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		GroupName: "test_delete_node_from_group_1_name",
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
