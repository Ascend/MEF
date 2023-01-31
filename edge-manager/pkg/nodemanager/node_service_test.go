// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"

	"edge-manager/pkg/database"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/util"
	"huawei.com/mindxedge/base/common"
)

const (
	memoryDsn           = ":memory:?cache=shared"
	defaultPageSize     = 20
	shuffledNumberCount = 10000
)

var (
	env environment
)

type environment struct {
	shuffledNumbers     []int
	shuffledNumbersLock sync.Mutex
	patches             *Patches
}

func (e *environment) setup() error {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := common.InitHwlogger(logConfig, logConfig); err != nil {
		return err
	}
	db, err := gorm.Open(sqlite.Open(memoryDsn))
	if err != nil {
		return err
	}
	if err := e.setupTables(db); err != nil {
		return err
	}
	e.patches = e.setupGoMonkeyPatches(db)
	return nil
}

func (e *environment) teardown() {
	e.patches.Reset()
}

func (e *environment) setupTables(db *gorm.DB) error {
	if err := db.AutoMigrate(&NodeInfo{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&NodeRelation{}); err != nil {
		return err
	}
	if err := db.AutoMigrate(&NodeGroup{}); err != nil {
		return err
	}
	return nil
}

func (e *environment) setupGoMonkeyPatches(db *gorm.DB) *Patches {
	service := &nodeStatusServiceImpl{}
	client := &kubeclient.Client{}
	return ApplyFuncReturn(database.GetDb, db).
		ApplyFuncReturn(NodeStatusServiceInstance, service).
		ApplyMethodReturn(service, "ListNodeStatus", map[string]string{}).
		ApplyMethodReturn(service, "GetNodeStatus", statusOffline, nil).
		ApplyMethodReturn(service, "GetAllocatableResource", &NodeResource{}, nil).
		ApplyMethodReturn(service, "GetAllocatableNpu", int64(0), nil).
		ApplyFuncReturn(kubeclient.GetKubeClient, client).
		ApplyPrivateMethod(client, "patchNode", e.mockPatchNode).
		ApplyMethodReturn(client, "ListNode", &v1.NodeList{}, nil).
		ApplyMethodReturn(client, "GetNode", &v1.Node{}, nil).
		ApplyMethodReturn(client, "DeleteNode", nil).
		ApplyFuncReturn(getAppInstanceCountByGroupId, int64(0), nil)
}

func (e *environment) mockPatchNode(string, []map[string]interface{}) (*v1.Node, error) {
	return &v1.Node{}, nil
}

func (e *environment) createNode(node *NodeInfo) error {
	return NodeServiceInstance().createNode(node)
}

func (e *environment) verifyDbNodeInfo(node *NodeInfo, ignoredFields ...string) error {
	var (
		dbNode *NodeInfo
		err    error
	)
	if node.ID == 0 {
		dbNode, err = NodeServiceInstance().getNodeByUniqueName(node.UniqueName)
	} else {
		dbNode, err = NodeServiceInstance().getNodeByID(node.ID)
	}
	if err != nil {
		return err
	}
	if !env.compareStruct(*node, *dbNode, ignoredFields...) {
		return errors.New("node not equal")
	}
	return nil
}

func (e *environment) createGroup(group *NodeGroup) error {
	return NodeServiceInstance().createNodeGroup(group)
}

func (e *environment) verifyDbNodeGroup(group *NodeGroup, ignoredFields ...string) error {
	var (
		dbGroup  *NodeGroup
		dbGroups *[]NodeGroup
		err      error
	)
	if group.ID == 0 {
		dbGroups, err = NodeServiceInstance().getNodeGroupsByName(1, defaultPageSize, group.GroupName)
		for _, group := range *dbGroups {
			if group.GroupName == group.GroupName {
				dbGroup = &group
				break
			}
		}
	} else {
		dbGroup, err = NodeServiceInstance().getNodeGroupByID(group.ID)
	}
	if dbGroup == nil {
		return gorm.ErrRecordNotFound
	}
	if err != nil {
		return err
	}
	if !env.compareStruct(*group, *dbGroup, ignoredFields...) {
		return errors.New("group not equal")
	}
	return nil
}

func (e *environment) createRelation(relation *NodeRelation) error {
	return NodeServiceInstance().addNodeToGroup(&[]NodeRelation{*relation})
}

func (e *environment) verifyDbNodeRelation(relation *NodeRelation, ignoredFields ...string) error {
	dbRelations, err := NodeServiceInstance().getRelationsByNodeID(relation.NodeID)
	var dbRelation *NodeRelation
	for _, r := range *dbRelations {
		if r.GroupID == relation.GroupID {
			dbRelation = &r
			break
		}
	}
	if dbRelation == nil {
		return gorm.ErrRecordNotFound
	}
	if err != nil {
		return err
	}
	if !env.compareStruct(*relation, *dbRelation, ignoredFields...) {
		return errors.New("node not equal")
	}
	return nil
}

func (e *environment) compareStruct(a, b interface{}, ignoredFields ...string) bool {
	aType := reflect.TypeOf(a)
	bType := reflect.TypeOf(b)
	if aType != bType {
		return false
	}
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	for i := 0; i < aType.NumField(); i++ {
		fieldName := aType.Field(i).Name
		shouldIgnore := false
		for _, ignoredFieldName := range ignoredFields {
			if fieldName == ignoredFieldName {
				shouldIgnore = true
				break
			}
		}
		if shouldIgnore {
			continue
		}
		aFieldValue := aValue.Field(i).Interface()
		bFieldValue := bValue.Field(i).Interface()
		if aFieldValue != bFieldValue {
			return false
		}
	}
	return true
}

func (e *environment) randomize(pointers ...interface{}) {
	replacements := map[string]string{
		"#{random}": fmt.Sprintf("%04d", e.nextRandomInt()),
	}
	for _, pointer := range pointers {
		e.randomizeInternal(pointer, replacements)
	}
}

func (e *environment) nextRandomInt() int {
	e.shuffledNumbersLock.Lock()
	if len(e.shuffledNumbers) == 0 {
		e.shuffledNumbers = make([]int, shuffledNumberCount)
		for i := 0; i < shuffledNumberCount; i++ {
			e.shuffledNumbers[i] = i
		}
		rand.Shuffle(shuffledNumberCount, func(i, j int) {
			temp := e.shuffledNumbers[i]
			e.shuffledNumbers[i] = e.shuffledNumbers[j]
			e.shuffledNumbers[j] = temp
		})
	}
	randInt := e.shuffledNumbers[len(e.shuffledNumbers)-1]
	e.shuffledNumbers = e.shuffledNumbers[0 : len(e.shuffledNumbers)-1]
	e.shuffledNumbersLock.Unlock()
	return randInt
}

func (e *environment) randomizeInternal(pointer interface{}, replacements map[string]string) {
	if replacements == nil {
		return
	}
	ptrValue := reflect.ValueOf(pointer)
	if ptrValue.Kind() != reflect.Ptr || ptrValue.IsNil() {
		return
	}
	objValue := ptrValue.Elem()
	switch objValue.Kind() {
	case reflect.String:
		replacedStr := objValue.String()
		for oldStr, newStr := range replacements {
			replacedStr = strings.ReplaceAll(replacedStr, oldStr, newStr)
		}
		objValue.Set(reflect.ValueOf(replacedStr))
	case reflect.Struct:
		for i := 0; i < objValue.NumField(); i++ {
			e.randomizeInternal(objValue.Field(i).Addr().Interface(), replacements)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < objValue.Len(); i++ {
			e.randomizeInternal(objValue.Index(i).Addr().Interface(), replacements)
		}
	default:
	}
}

func TestMain(m *testing.M) {
	env = environment{}
	if err := env.setup(); err != nil {
		fmt.Printf("failed to setup test environment, reason: %v", err)
		return
	}
	defer env.teardown()
	code := m.Run()
	hwlog.RunLog.Infof("test complete, exitCode=%d", code)
}

func TestCreateNode(t *testing.T) {
	Convey("createNode functional test", t, createNodeFunctionalTest)
	Convey("createNode validation test", t, createNodeValidationTest)
}

func createNodeFunctionalTest() {
	node := &NodeInfo{
		Description: "test-create-node-1-description",
		NodeName:    "test-create-node-1-name",
		UniqueName:  "test-create-node-1-unique-name",
	}

	Convey("normal input", func() {
		args := fmt.Sprintf(`
			{
    			"nodeName": "%s",
    			"uniqueName": "%s",
    			"description": "%s"
			}`, node.NodeName, node.UniqueName, node.Description)
		resp := createNode(args)
		So(resp.Status, ShouldEqual, common.Success)
		So(env.verifyDbNodeInfo(node, "ID", "UpdatedAt", "CreatedAt", "IsManaged"), ShouldBeNil)
	})
}

func createNodeValidationTest() {
	node := &NodeInfo{
		NodeName:    "test-create-node-2-name",
		UniqueName:  "test-create-node-2-unique-name",
		Description: "test-create-node-2-description",
	}

	Convey("empty description", func() {
		node.Description = ""
		args := fmt.Sprintf(`
			{
    			"nodeName": "%s",
    			"uniqueName": "%s"
			}`, node.NodeName, node.UniqueName)
		resp := createNode(args)
		So(resp.Status, ShouldEqual, common.Success)
		So(env.verifyDbNodeInfo(node, "ID", "UpdatedAt", "CreatedAt", "IsManaged"), ShouldBeNil)
	})

	Convey("empty uniqueName", func() {
		args := fmt.Sprintf(`
			{
    			"nodeName": "%s",
    			"description": "%s"
			}`, node.NodeName, node.Description)
		resp := createNode(args)
		So(resp.Status, ShouldNotEqual, common.Success)
	})
}

func TestGetNodeDetail(t *testing.T) {
	Convey("getNodeDetail functional test", t, getNodeDetailFunctionalTest)
	Convey("getNodeDetail validation test", t, getNodeDetailValidationTest)
}

func getNodeDetailFunctionalTest() {
	node := &NodeInfo{
		Description: "test-get-node-detail-1-description",
		NodeName:    "test-get-node-detail-1-name",
		UniqueName:  "test-get-node-detail-1-unique-name",
		IP:          "0.0.0.0",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	env.randomize(node)
	So(env.createNode(node), ShouldBeNil)

	Convey("normal input", func() {
		resp := getNodeDetail(node.ID)
		So(resp.Status, ShouldEqual, common.Success)
		nodeInfoDetail, ok := resp.Data.(NodeInfoDetail)
		So(ok, ShouldBeTrue)
		So(nodeInfoDetail.NodeInfoEx.NodeInfo, ShouldResemble, *node)
	})
}

func getNodeDetailValidationTest() {
	Convey("bad id type", func() {
		args := `{"id": "1"}`
		resp := getNodeDetail(args)
		So(resp.Status, ShouldNotEqual, common.Success)
	})
}

func TestModifyNode(t *testing.T) {
	Convey("modifyNode functional test", t, modifyNodeFunctionalTest)
	Convey("modifyNode validation test", t, modifyNodeValidationTest)
}

func modifyNodeFunctionalTest() {
	node := &NodeInfo{
		Description: "test-modify-node-1-description",
		NodeName:    "test-modify-node-1-name",
		UniqueName:  "test-modify-node-1-unique-name",
		IsManaged:   true,
		IP:          "0.0.0.0",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	So(env.createNode(node), ShouldBeNil)
	node.Description = "test-modify-node-1-description-modified"
	node.NodeName = "test-modify-node-1-name-modified"

	Convey("normal input", func() {
		args := fmt.Sprintf(`
			{
    			"description": "%s",
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.Description, node.NodeName, node.ID)
		resp := modifyNode(args)
		So(resp.Status, ShouldEqual, common.Success)
		So(env.verifyDbNodeInfo(node, "UpdatedAt"), ShouldBeNil)
	})
}

func modifyNodeValidationTest() {
	node := &NodeInfo{
		Description: "test-modify-node-2-#{random}-description",
		NodeName:    "test-modify-node-2-#{random}-name",
		UniqueName:  "test-modify-node-2-#{random}-unique-name",
		IP:          "0.0.0.0",
		IsManaged:   true,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	env.randomize(node)

	Convey("empty description", func() {
		So(env.createNode(node), ShouldBeNil)
		node.Description = ""
		args := fmt.Sprintf(`
			{
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.NodeName, node.ID)
		resp := modifyNode(args)
		So(resp.Status, ShouldEqual, common.Success)
		So(env.verifyDbNodeInfo(node, "UpdatedAt"), ShouldBeNil)
	})

	Convey("modify unmanaged node", func() {
		node.IsManaged = false
		So(env.createNode(node), ShouldBeNil)
		args := fmt.Sprintf(`
			{
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.NodeName, node.ID)
		resp := modifyNode(args)
		So(resp.Status, ShouldNotEqual, common.Success)
	})
}

func TestGetNodeStatistics(t *testing.T) {
	Convey("getNodeStatistics functional test", t, getNodeStatisticsFunctionalTest)
}

func getNodeStatisticsFunctionalTest() {
	Convey("normal input", func() {
		resp := getNodeStatistics(``)
		So(resp.Status, ShouldEqual, common.Success)
	})
}

func TestGroupStatistics(t *testing.T) {
	Convey("getGroupNodeStatistics functional test", t, groupStatisticsFunctionalTest)
}

func groupStatisticsFunctionalTest() {
	Convey("normal input", func() {
		resp := getGroupNodeStatistics(``)
		So(resp.Status, ShouldEqual, common.Success)
	})
}

func TestCreateGroup(t *testing.T) {
	Convey("createGroup functional test", t, createGroupFunctionalTest)
	Convey("createGroup validation test", t, createGroupValidationTest)
}

func createGroupFunctionalTest() {
	group := &NodeGroup{
		Description: "test-create-group-1-description",
		GroupName:   "test_create_group_1_name",
	}

	Convey("normal input", func() {
		args := fmt.Sprintf(`
			{
    			"nodeGroupName": "%s",
    			"description": "%s"
			}`, group.GroupName, group.Description)
		resp := createGroup(args)
		So(resp.Status, ShouldEqual, common.Success)
		So(env.verifyDbNodeGroup(group, "ID", "UpdatedAt", "CreatedAt"), ShouldBeNil)
	})
}

func createGroupValidationTest() {
	group := &NodeGroup{
		Description: "test-create-group-2-description",
		GroupName:   "test_create_group_2_name",
	}

	Convey("description not present", func() {
		args := fmt.Sprintf(`{"nodeGroupName": "%s"}`, group.GroupName)
		resp := createGroup(args)
		So(resp.Status, ShouldEqual, common.Success)
	})

	Convey("groupName not present", func() {
		args := fmt.Sprintf(`{"description": "%s"}`, group.Description)
		resp := createGroup(args)
		So(resp.Status, ShouldEqual, common.ErrorParamInvalid)
	})
}

func TestGetGroupDetail(t *testing.T) {
	Convey("getEdgeNodeGroupDetail functional test", t, getGroupDetailFunctionalTest)
	Convey("getEdgeNodeGroupDetail validation test", t, getGroupDetailValidationTest)
}

func getGroupDetailFunctionalTest() {
	group := &NodeGroup{
		Description: "test-get-group-detail-1-description",
		GroupName:   "test_get_group_detail_1_name",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	So(env.createGroup(group), ShouldBeNil)

	Convey("normal input", func() {
		resp := getEdgeNodeGroupDetail(group.ID)
		So(resp.Status, ShouldEqual, common.Success)
		groupDetail, ok := resp.Data.(NodeGroupDetail)
		So(ok, ShouldBeTrue)
		So(groupDetail.NodeGroup, ShouldResemble, *group)
	})
}

func getGroupDetailValidationTest() {
	Convey("bad id type", func() {
		args := `{"id": "1"}`
		resp := getEdgeNodeGroupDetail(args)
		So(resp.Status, ShouldNotEqual, common.Success)
	})
}

func TestListManagedNode(t *testing.T) {
	Convey("listManagedNode functional test", t, litManagedNodeFunctionalTest)
}

func litManagedNodeFunctionalTest() {
	Convey("normal input", func() {
		args := util.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listManagedNode(args)
		So(resp.Status, ShouldEqual, common.Success)
	})
}

func TestAddUnManagedNode(t *testing.T) {
	Convey("addUnmanagedNode functional test", t, addUnManagedNodeFunctionalTest)
	Convey("addUnmanagedNode validation test", t, addUnManagedNodeValidationTest)
}

func addUnManagedNodeFunctionalTest() {
	node := &NodeInfo{
		Description: "test-adn-1-description",
		NodeName:    "test-adn-1-name",
		UniqueName:  "test-adn-1-unique-name",
		IP:          "0.0.0.0",
		IsManaged:   false,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		Description: "test-add-adn-1-description",
		GroupName:   "test_add_adn_1_name",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	So(env.createNode(node), ShouldBeNil)
	So(env.createGroup(group), ShouldBeNil)

	Convey("normal input", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
            "groupIDs": [%d],
            "nodeID": %d
			}`, node.NodeName, node.Description, group.ID, node.ID)
		resp := addUnManagedNode(args)
		So(resp.Status, ShouldEqual, common.Success)
		node.IsManaged = true
		So(env.verifyDbNodeInfo(node, "CreatedAt", "UpdatedAt"), ShouldBeNil)
		relation := &NodeRelation{NodeID: node.ID, GroupID: group.ID}
		So(env.verifyDbNodeRelation(relation, "CreatedAt"), ShouldBeNil)
	})
}

func addUnManagedNodeValidationTest() {
	node := &NodeInfo{
		Description: "test-adn-#{random}-description",
		NodeName:    "test-adn-#{random}-name",
		UniqueName:  "test-adn-#{random}-unique-name",
		IP:          "0.0.0.0",
		IsManaged:   false,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	env.randomize(node)
	So(env.createNode(node), ShouldBeNil)

	Convey("groupIDs not present", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
            "nodeID": %d
			}`, node.NodeName, node.Description, node.ID)
		resp := addUnManagedNode(args)
		So(resp.Status, ShouldEqual, common.Success)
	})

	Convey("groupIDs is empty", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
			"groupIDs": [],
            "nodeID": %d
			}`, node.NodeName, node.Description, node.ID)
		resp := addUnManagedNode(args)
		So(resp.Status, ShouldEqual, common.Success)
	})
}

func TestListUnManagedNode(t *testing.T) {
	Convey("listUnmanagedNode functional test", t, listUnManagedNodeFunctionalTest)
}

func listUnManagedNodeFunctionalTest() {
	Convey("normal input", func() {
		args := util.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listUnmanagedNode(args)
		So(resp.Status, ShouldEqual, common.Success)
	})
}

func TestListNode(t *testing.T) {
	Convey("listNode functional test", t, litNodeFunctionalTest)
}

func litNodeFunctionalTest() {
	Convey("normal input", func() {
		args := util.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listNode(args)
		So(resp.Status, ShouldEqual, common.Success)
	})
}

func TestBatchDeleteNode(t *testing.T) {
	Convey("batchDeleteNode functional test", t, batchDeleteNodeFunctionalTest)
	Convey("batchDeleteNode validation test", t, batchDeleteNodeValidationTest)
}

func batchDeleteNodeFunctionalTest() {
	node := &NodeInfo{
		Description: "test-batch-delete-node-1-description",
		NodeName:    "test-batch-delete-node-1-name",
		UniqueName:  "test-batch-delete-node-1-unique-name",
		IP:          "0.0.0.0",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	So(env.createNode(node), ShouldBeNil)

	Convey("normal input", func() {
		args := fmt.Sprintf(`[%d]`, node.ID)
		resp := batchDeleteNode(args)
		So(resp.Status, ShouldEqual, common.Success)
		deleteCount, ok := resp.Data.(int64)
		So(ok, ShouldBeTrue)
		So(deleteCount, ShouldEqual, int64(1))
		So(env.verifyDbNodeInfo(node), ShouldEqual, gorm.ErrRecordNotFound)
	})
}

func batchDeleteNodeValidationTest() {
	Convey("empty request", func() {
		args := ``
		resp := batchDeleteNode(args)
		So(resp.Status, ShouldNotEqual, common.Success)
	})
}

func TestBatchDeleteNodeRelation(t *testing.T) {
	Convey("batchDeleteNodeRelation functional test", t, batchDeleteNodeRelationFunctionalTest)
	Convey("batchDeleteNodeRelation validation test", t, batchDeleteNodeRelationValidationTest)
}

func batchDeleteNodeRelationFunctionalTest() {
	node := &NodeInfo{
		Description: "test-batch-delete-relation-1-description",
		NodeName:    "test-batch-delete-relation-1-name",
		UniqueName:  "test-batch-delete-relation-1-unique-name",
		IP:          "0.0.0.0",
		IsManaged:   true,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		Description: "test-batch-delete-relation-1-description",
		GroupName:   "test_batch_delete_relation_1_nme",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	So(env.createNode(node), ShouldBeNil)
	So(env.createGroup(group), ShouldBeNil)
	relation := &NodeRelation{
		NodeID:    node.ID,
		GroupID:   group.ID,
		CreatedAt: time.Now().Format(TimeFormat),
	}
	So(env.createRelation(relation), ShouldBeNil)

	Convey("normal input", func() {
		args := fmt.Sprintf(`[
            {
                "nodeID": %d,
                "groupID": %d
            }]`, node.ID, group.ID)
		resp := batchDeleteNodeRelation(args)
		So(resp.Status, ShouldEqual, common.Success)
		deleteCount, ok := resp.Data.(int64)
		So(ok, ShouldBeTrue)
		So(deleteCount, ShouldEqual, int64(1))
		So(env.verifyDbNodeRelation(relation), ShouldEqual, gorm.ErrRecordNotFound)
	})
}

func batchDeleteNodeRelationValidationTest() {
	Convey("nodeID not present", func() {
		args := `[{"groupID": 1}]`
		resp := batchDeleteNodeRelation(args)
		So(resp.Status, ShouldEqual, common.ErrorParamInvalid)
	})
	Convey("duplicate relations", func() {
		args := `[{"groupID":1, "nodeID":2},{"nodeID":2,"groupID":1}]`
		resp := batchDeleteNodeRelation(args)
		So(resp.Status, ShouldEqual, common.ErrorParamInvalid)
	})
}

func TestAddNodeRelation(t *testing.T) {
	Convey("addNodeRelation functional test", t, addNodeRelationFunctionalTest)
	Convey("addNodeRelation validation test", t, addNodeRelationValidationTest)
}

func addNodeRelationFunctionalTest() {
	node := &NodeInfo{
		Description: "test-add-relation-1-description",
		NodeName:    "test-add-relation-1-name",
		UniqueName:  "test-add-relation-1-description-unique-name",
		IP:          "0.0.0.0",
		IsManaged:   true,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		Description: "test-add-relation-1-description",
		GroupName:   "test_add_relation_1_name",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	So(env.createNode(node), ShouldBeNil)
	So(env.createGroup(group), ShouldBeNil)

	Convey("normal input", func() {
		args := fmt.Sprintf(`{"groupID": %d, "nodeIDs": [%d]}`, group.ID, node.ID)
		resp := addNodeRelation(args)
		So(resp.Status, ShouldEqual, common.Success)
		relation := &NodeRelation{NodeID: node.ID, GroupID: group.ID}
		So(env.verifyDbNodeRelation(relation, "CreatedAt"), ShouldBeNil)
	})
}

func addNodeRelationValidationTest() {
	Convey("groupID not present", func() {
		args := `{"nodeIDs": [1]}`
		resp := addNodeRelation(args)
		So(resp.Status, ShouldEqual, common.ErrorParamInvalid)
	})
}

func TestListEdgeNodeGroup(t *testing.T) {
	Convey("listEdgeNodeGroup functional test", t, listEdgeNodeGroupFunctionalTest)
}

func listEdgeNodeGroupFunctionalTest() {
	Convey("normal input", func() {
		args := util.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listEdgeNodeGroup(args)
		So(resp.Status, ShouldEqual, common.Success)
	})
}

func TestModifyGroup(t *testing.T) {
	Convey("modifyGroup functional test", t, modifyGroupFunctionalTest)
	Convey("modifyGroup validation test", t, modifyGroupValidationTest)
}

func modifyGroupFunctionalTest() {
	group := &NodeGroup{
		Description: "test-modify-group-1-description",
		GroupName:   "test_modify_group_1_n",
	}
	So(env.createGroup(group), ShouldBeNil)
	group.Description += "-m"
	group.GroupName += "_m"

	Convey("normal input", func() {
		args := fmt.Sprintf(`
				{
   					"groupID": %d,
   					"nodeGroupName": "%s",
   					"description": "%s"
				}`, group.ID, group.GroupName, group.Description)
		resp := modifyNodeGroup(args)
		So(resp.Status, ShouldEqual, common.Success)
		So(env.verifyDbNodeGroup(group, "UpdatedAt"), ShouldBeNil)
	})
}

func modifyGroupValidationTest() {
	group := &NodeGroup{
		Description: "test-modify-group-2-#{random}-description",
		GroupName:   "test_modify_group_2_#{random}_n",
	}
	env.randomize(group)
	So(env.createGroup(group), ShouldBeNil)

	Convey("empty description", func() {
		args := fmt.Sprintf(`
				{
   					"groupID": %d,
   					"nodeGroupName": "%s"
				}`, group.ID, group.GroupName)
		resp := modifyNodeGroup(args)
		So(resp.Status, ShouldEqual, common.Success)
	})

	Convey("empty groupID", func() {
		args := fmt.Sprintf(`
				{
   					"nodeGroupName": "%s",
   					"description": "%s"
				}`, group.GroupName, group.Description)
		resp := modifyNodeGroup(args)
		So(resp.Status, ShouldEqual, common.ErrorParamInvalid)
	})
}

func TestBatchDeleteGroup(t *testing.T) {
	Convey("batchDeleteNodeGroup functional test", t, batchDeleteGroupFunctionalTest)
	Convey("batchDeleteNodeGroup validation test", t, batchDeleteGroupValidationTest)
}

func batchDeleteGroupFunctionalTest() {
	group := &NodeGroup{
		Description: "test-batch-delete-group-1-description",
		GroupName:   "test_batch_delete_group_1_name",
	}
	So(env.createGroup(group), ShouldBeNil)

	Convey("normal input", func() {
		args := fmt.Sprintf(`{"groupIDs": [%d]}`, group.ID)
		resp := batchDeleteNodeGroup(args)
		So(resp.Status, ShouldEqual, common.Success)
		deleteIDs, ok := resp.Data.([]int64)
		So(ok, ShouldBeTrue)
		So(deleteIDs, ShouldResemble, []int64{group.ID})
		So(env.verifyDbNodeGroup(group), ShouldEqual, gorm.ErrRecordNotFound)
	})
}

func batchDeleteGroupValidationTest() {
	Convey("GroupIDs not present", func() {
		args := ``
		resp := batchDeleteNodeGroup(args)
		So(resp.Status, ShouldNotEqual, common.Success)
	})
	Convey("bad group id", func() {
		args := `{"groupIDs": [-1]}`
		resp := batchDeleteNodeGroup(args)
		So(resp.Status, ShouldEqual, common.ErrorParamInvalid)
	})
	Convey("bad id type", func() {
		args := `{"groupIDs": ["1"]}`
		resp := batchDeleteNodeGroup(args)
		So(resp.Status, ShouldNotEqual, common.Success)
	})
	Convey("duplicate id", func() {
		args := `{"groupIDs": [1, 1]}`
		resp := batchDeleteNodeGroup(args)
		So(resp.Status, ShouldEqual, common.ErrorParamInvalid)
	})
	Convey("empty list", func() {
		args := `{"groupIDs": []}`
		resp := batchDeleteNodeGroup(args)
		So(resp.Status, ShouldEqual, common.ErrorParamInvalid)
	})
}
