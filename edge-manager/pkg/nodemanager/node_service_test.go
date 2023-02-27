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

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/database"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
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
	patches             *gomonkey.Patches
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

func (e *environment) setupGoMonkeyPatches(db *gorm.DB) *gomonkey.Patches {
	service := &nodeSyncImpl{}
	client := &kubeclient.Client{}
	return gomonkey.ApplyFuncReturn(database.GetDb, db).
		ApplyFuncReturn(NodeSyncInstance, service).
		ApplyMethodReturn(service, "ListNodeStatus", map[string]string{}).
		ApplyMethodReturn(service, "GetNodeStatus", statusOffline, nil).
		ApplyMethodReturn(service, "GetAllocatableResource", &NodeResource{}, nil).
		ApplyMethodReturn(service, "GetAvailableResource", &NodeResource{}, nil).
		ApplyFuncReturn(kubeclient.GetKubeClient, client).
		ApplyPrivateMethod(client, "patchNode", e.mockPatchNode).
		ApplyMethodReturn(client, "ListNode", &v1.NodeList{}, nil).
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
	method := gomonkey.ApplyMethodFunc(kubeclient.GetKubeClient(), "AddNodeLabels",
		func(string, map[string]string) (*v1.Node, error) {
			return nil, nil
		})
	defer method.Reset()
	return NodeServiceInstance().addNodeToGroup(relation, "")
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

func TestGetNodeDetail(t *testing.T) {
	convey.Convey("getNodeDetail functional test", t, getNodeDetailFunctionalTest)
	convey.Convey("getNodeDetail validation test", t, getNodeDetailValidationTest)
}

func getNodeDetailFunctionalTest() {
	node := &NodeInfo{
		Description:  "test-get-node-detail-1-description",
		NodeName:     "test-get-node-detail-1-name",
		UniqueName:   "test-get-node-detail-1-unique-name",
		SerialNumber: "test-get-node-detail-1-serial-number",
		IP:           "0.0.0.0",
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		resp := getNodeDetail(node.ID)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		nodeInfoDetail, ok := resp.Data.(NodeInfoDetail)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(nodeInfoDetail.NodeInfoEx.NodeInfo, convey.ShouldResemble, *node)
	})
}

func getNodeDetailValidationTest() {
	convey.Convey("bad id type", func() {
		args := `{"id": "1"}`
		resp := getNodeDetail(args)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	})
}

func TestModifyNode(t *testing.T) {
	convey.Convey("modifyNode functional test", t, modifyNodeFunctionalTest)
	convey.Convey("modifyNode validation test", t, modifyNodeValidationTest)
}

func modifyNodeFunctionalTest() {
	node := &NodeInfo{
		Description:  "test-modify-node-1-description",
		NodeName:     "test-modify-node-1-name",
		UniqueName:   "test-modify-node-1-unique-name",
		SerialNumber: "test-modify-node-1-serial-number",
		IsManaged:    true,
		IP:           "0.0.0.0",
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)
	node.Description = "test-modify-node-1-description-modified"
	node.NodeName = "test-modify-node-1-name-modified"

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`
			{
    			"description": "%s",
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.Description, node.NodeName, node.ID)
		resp := modifyNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node, "UpdatedAt")
		convey.So(verifyRes, convey.ShouldBeNil)
	})
}

func modifyNodeValidationTest() {
	node := &NodeInfo{
		Description:  "test-modify-node-2-#{random}-description",
		NodeName:     "test-modify-node-2-#{random}-name",
		UniqueName:   "test-modify-node-2-#{random}-unique-name",
		SerialNumber: "test-modify-node-2-#{random}-serial-number",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	env.randomize(node)

	convey.Convey("empty description", func() {
		nodeRes := env.createNode(node)
		convey.So(nodeRes, convey.ShouldBeNil)
		node.Description = ""
		args := fmt.Sprintf(`
			{
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.NodeName, node.ID)
		resp := modifyNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node, "UpdatedAt")
		convey.So(verifyRes, convey.ShouldBeNil)
	})

	convey.Convey("modify unmanaged node", func() {
		node.IsManaged = false
		nodeRes := env.createNode(node)
		convey.So(nodeRes, convey.ShouldBeNil)
		args := fmt.Sprintf(`
			{
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.NodeName, node.ID)
		resp := modifyNode(args)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	})
}

func TestGetNodeStatistics(t *testing.T) {
	convey.Convey("getNodeStatistics functional test", t, getNodeStatisticsFunctionalTest)
}

func getNodeStatisticsFunctionalTest() {
	convey.Convey("normal input", func() {
		resp := getNodeStatistics(``)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func TestGroupStatistics(t *testing.T) {
	convey.Convey("getGroupNodeStatistics functional test", t, groupStatisticsFunctionalTest)
}

func groupStatisticsFunctionalTest() {
	convey.Convey("normal input", func() {
		resp := getGroupNodeStatistics(``)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func TestCreateGroup(t *testing.T) {
	convey.Convey("createGroup functional test", t, createGroupFunctionalTest)
	convey.Convey("createGroup validation test", t, createGroupValidationTest)
}

func createGroupFunctionalTest() {
	group := &NodeGroup{
		Description: "test-create-group-1-description",
		GroupName:   "test_create_group_1_name",
	}

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`
			{
    			"nodeGroupName": "%s",
    			"description": "%s"
			}`, group.GroupName, group.Description)
		resp := createGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeGroup(group, "ID", "UpdatedAt", "CreatedAt")
		convey.So(verifyRes, convey.ShouldBeNil)
	})
}

func createGroupValidationTest() {
	group := &NodeGroup{
		Description: "test-create-group-2-description",
		GroupName:   "test_create_group_2_name",
	}

	convey.Convey("description not present", func() {
		args := fmt.Sprintf(`{"nodeGroupName": "%s"}`, group.GroupName)
		resp := createGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})

	convey.Convey("groupName not present", func() {
		args := fmt.Sprintf(`{"description": "%s"}`, group.Description)
		resp := createGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})
	convey.Convey("groupName duplicate", func() {
		args := fmt.Sprintf(`{"nodeGroupName": "%s"}`, group.GroupName)
		_ = createGroup(args)
		resp := createGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorNodeMrgDuplicate)
	})
}

func TestGetGroupDetail(t *testing.T) {
	convey.Convey("getEdgeNodeGroupDetail functional test", t, getGroupDetailFunctionalTest)
	convey.Convey("getEdgeNodeGroupDetail validation test", t, getGroupDetailValidationTest)
	convey.Convey("getEdgeNodeGroupDetail id not exit", t, getGroupDetailValidationTest2)
}

func getGroupDetailFunctionalTest() {
	group := &NodeGroup{
		GroupName: "test_get_group_detail_1_name",
		CreatedAt: time.Now().Format(TimeFormat),
		UpdatedAt: time.Now().Format(TimeFormat),
	}
	convey.So(env.createGroup(group), convey.ShouldBeNil)
	node := &NodeInfo{
		NodeName:   "test-get-group-detail-1-node",
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
	convey.Convey("normal input", func() {
		resp := getEdgeNodeGroupDetail(group.ID)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		groupDetail, ok := resp.Data.(NodeGroupDetail)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(groupDetail.NodeGroup, convey.ShouldResemble, *group)
	})
}

func getGroupDetailValidationTest() {
	convey.Convey("bad id type", func() {
		args := `{"id": "1"}`
		resp := getEdgeNodeGroupDetail(args)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	})
}

func getGroupDetailValidationTest2() {
	convey.Convey("id not exit", func() {
		resp := getEdgeNodeGroupDetail(uint64(0))
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	})
}

func TestListManagedNode(t *testing.T) {
	convey.Convey("listManagedNode functional test", t, listManagedNodeFunctionalTest)
	convey.Convey("listManagedNode functional test", t, listManagedNodeTest1)
}

func listManagedNodeFunctionalTest() {
	convey.Convey("normal input", func() {
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func listManagedNodeTest1() {
	convey.Convey("error input", func() {
		args := ""
		resp := listManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
	})
}

func TestAddUnManagedNode(t *testing.T) {
	convey.Convey("addUnmanagedNode functional test", t, addUnManagedNodeFunctionalTest)
	convey.Convey("addUnmanagedNode validation test", t, addUnManagedNodeValidationTest)
}

func addUnManagedNodeFunctionalTest() {
	node := &NodeInfo{
		Description:  "test-adn-1-description",
		NodeName:     "test-adn-1-name",
		UniqueName:   "test-adn-1-unique-name",
		SerialNumber: "test-adn-1-serial-number",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		Description: "test-add-adn-1-description",
		GroupName:   "test_add_adn_1_name",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	resNode := env.createNode(node)
	convey.So(resNode, convey.ShouldBeNil)
	resGroup := env.createGroup(group)
	convey.So(resGroup, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
            "groupIDs": [%d],
            "nodeID": %d
			}`, node.NodeName, node.Description, group.ID, node.ID)
		resp := addUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		node.IsManaged = true
		convey.So(env.verifyDbNodeInfo(node, "CreatedAt", "UpdatedAt"), convey.ShouldBeNil)
		relation := &NodeRelation{NodeID: node.ID, GroupID: group.ID}
		convey.So(env.verifyDbNodeRelation(relation, "CreatedAt"), convey.ShouldBeNil)
	})
}

func addUnManagedNodeValidationTest() {
	node := &NodeInfo{
		Description:  "test-adn-#{random}-description",
		NodeName:     "test-adn-#{random}-name",
		UniqueName:   "test-adn-#{random}-unique-name",
		SerialNumber: "test-adn-#{random}-serial-umber",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	env.randomize(node)
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("groupIDs not present", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
            "nodeID": %d
			}`, node.NodeName, node.Description, node.ID)
		resp := addUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})

	convey.Convey("groupIDs is empty", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
			"groupIDs": [],
            "nodeID": %d
			}`, node.NodeName, node.Description, node.ID)
		resp := addUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func TestListUnManagedNode(t *testing.T) {
	convey.Convey("listUnmanagedNode functional test", t, listUnManagedNodeFunctionalTest)
	convey.Convey("listUnmanagedNode param error", t, listUnManagedNodeFunctionalError)
}

func listUnManagedNodeFunctionalTest() {
	convey.Convey("normal input", func() {
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listUnmanagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func listUnManagedNodeFunctionalError() {
	convey.Convey("error input", func() {
		args := ""
		resp := listUnmanagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})
}

func TestListNode(t *testing.T) {
	convey.Convey("listNode functional test", t, listNodeFunctionalTest)
	convey.Convey("listNode error input", t, listNodeTest1)
}

func listNodeFunctionalTest() {
	convey.Convey("normal input", func() {
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func listNodeTest1() {
	convey.Convey("error input", func() {
		args := ""
		resp := listNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
	})
}

func TestBatchDeleteNode(t *testing.T) {
	convey.Convey("batchDeleteNode functional test", t, batchDeleteNodeFunctionalTest)
	convey.Convey("batchDeleteNode validation test", t, batchDeleteNodeValidationTest)
}

func batchDeleteNodeFunctionalTest() {
	node := &NodeInfo{
		Description:  "test-batch-delete-node-1-description",
		NodeName:     "test-batch-delete-node-1-name",
		UniqueName:   "test-batch-delete-node-1-unique-name",
		SerialNumber: "test-batch-delete-node-1-serial-number",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`{"nodeIDs": [%d]}`, node.ID)
		resp := batchDeleteNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node)
		convey.So(verifyRes, convey.ShouldNotEqual, "record not found")
	})
}

func batchDeleteNodeValidationTest() {
	convey.Convey("empty request", func() {
		args := ``
		resp := batchDeleteNode(args)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	})
}

func TestBatchDeleteNodeRelation(t *testing.T) {
	convey.Convey("batchDeleteNodeRelation functional test", t, batchDeleteNodeRelationFunctionalTest)
	convey.Convey("batchDeleteNodeRelation validation test", t, batchDeleteNodeRelationValidationTest)
}

func batchDeleteNodeRelationFunctionalTest() {
	node := &NodeInfo{
		Description:  "test-batch-delete-relation-1-description",
		NodeName:     "test-batch-delete-relation-1-name",
		UniqueName:   "test-batch-delete-relation-1-unique-name",
		SerialNumber: "test-batch-delete-relation-1-serial-number",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		Description: "test-batch-delete-relation-1-description",
		GroupName:   "test_batch_delete_relation_1_nme",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
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

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`[
            {
                "nodeID": %d,
                "groupID": %d
            }]`, node.ID, group.ID)
		resp := batchDeleteNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeRelation(relation)
		convey.So(verifyRes, convey.ShouldEqual, gorm.ErrRecordNotFound)
	})
}

func batchDeleteNodeRelationValidationTest() {
	convey.Convey("nodeID not present", func() {
		args := `[{"groupID": 1}]`
		resp := batchDeleteNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})
	convey.Convey("duplicate relations", func() {
		args := `[{"groupID":1, "nodeID":2},{"nodeID":2,"groupID":1}]`
		resp := batchDeleteNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})
}

func TestAddNodeRelation(t *testing.T) {
	convey.Convey("addNodeRelation functional test", t, addNodeRelationFunctionalTest)
	convey.Convey("addNodeRelation validation test", t, addNodeRelationValidationTest)
}

func addNodeRelationFunctionalTest() {
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

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`{"groupID": %d, "nodeIDs": [%d]}`, group.ID, node.ID)
		resp := addNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		relation := &NodeRelation{NodeID: node.ID, GroupID: group.ID}
		verifyRes := env.verifyDbNodeRelation(relation, "CreatedAt")
		convey.So(verifyRes, convey.ShouldBeNil)
	})
}

func addNodeRelationValidationTest() {
	convey.Convey("groupID not present", func() {
		args := `{"nodeIDs": [1]}`
		resp := addNodeRelation(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})
}

func TestListEdgeNodeGroup(t *testing.T) {
	convey.Convey("listEdgeNodeGroup functional test", t, listEdgeNodeGroupFunctionalTest)
	convey.Convey("listEdgeNodeGroup error test", t, listEdgeNodeGroupTest1)
}

func listEdgeNodeGroupFunctionalTest() {
	convey.Convey("normal input", func() {
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listEdgeNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func listEdgeNodeGroupTest1() {
	convey.Convey("error input", func() {
		args := ""
		resp := listEdgeNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
	})
}

func TestModifyGroup(t *testing.T) {
	convey.Convey("modifyGroup functional test", t, modifyGroupFunctionalTest)
	convey.Convey("modifyGroup validation test", t, modifyGroupValidationTest)
}

func modifyGroupFunctionalTest() {
	group := &NodeGroup{
		Description: "test-modify-group-1-description",
		GroupName:   "test_modify_group_1_n",
	}
	res := env.createGroup(group)
	convey.So(res, convey.ShouldBeNil)
	group.Description += "-m"
	group.GroupName += "_m"

	convey.Convey("normal input", func() {
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
	})
}

func modifyGroupValidationTest() {
	group := &NodeGroup{
		Description: "test-modify-group-2-#{random}-description",
		GroupName:   "test_modify_group_2_#{random}_n",
	}
	env.randomize(group)
	res := env.createGroup(group)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("empty description", func() {
		args := fmt.Sprintf(`
				{
   					"groupID": %d,
   					"nodeGroupName": "%s"
				}`, group.ID, group.GroupName)
		resp := modifyNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})

	convey.Convey("empty groupID", func() {
		args := fmt.Sprintf(`
				{
   					"nodeGroupName": "%s",
   					"description": "%s"
				}`, group.GroupName, group.Description)
		resp := modifyNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})
}

func TestBatchDeleteGroup(t *testing.T) {
	convey.Convey("batchDeleteNodeGroup functional test", t, batchDeleteGroupFunctionalTest)
	convey.Convey("batchDeleteNodeGroup validation test", t, batchDeleteGroupValidationTest)
}

func batchDeleteGroupFunctionalTest() {
	group := &NodeGroup{
		Description: "test-batch-delete-group-1-description",
		GroupName:   "test_batch_delete_group_1_name",
	}
	res := env.createGroup(group)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`{"groupIDs": [%d]}`, group.ID)
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeGroup(group)
		convey.So(verifyRes, convey.ShouldEqual, gorm.ErrRecordNotFound)
	})
}

func batchDeleteGroupValidationTest() {
	convey.Convey("GroupIDs not present", func() {
		args := ``
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
	})
	convey.Convey("bad group id", func() {
		args := `{"groupIDs": [-1]}`
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})
	convey.Convey("bad id type", func() {
		args := `{"groupIDs": ["1"]}`
		resp := batchDeleteNodeGroup(args)
		convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
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
}

func TestDeleteNodeFromGroup(t *testing.T) {
	convey.Convey("batchDeleteNodeGroup functional test", t, testDeleteNodeFromGroup)
	convey.Convey("batchDeleteNodeGroup validation test", t, deleteNodeFromGroupValidation)
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
	convey.Convey("test deleteNodeFromGroup success", func() {
		args := fmt.Sprintf(`{
                "nodeIDs": [%d],
                "groupID": %d
            }`, node.ID, group.ID)
		resp := deleteNodeFromGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func deleteNodeFromGroupValidation() {
	convey.Convey("test deleteNodeFromGroup param error", func() {
		args := fmt.Sprintf(`{
                "nodeIDs": %d,
                "groupID": %d
            }`, 1, 1)
		resp := deleteNodeFromGroup(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})
}

func TestInnerGetNodeInfoByUniqueName(t *testing.T) {
	convey.Convey("InnerGetNodeInfoByUniqueName functional test", t, func() {
		convey.Convey("innerGetNodeInfoByUniqueName success", func() {
			node := &NodeInfo{
				NodeName:     "test-inner-node1",
				UniqueName:   "test-inner-node-unique-name1",
				SerialNumber: "test-inner-node-serial-number1",
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
		})
		convey.Convey("innerGetNodeInfoByUniqueName param error", func() {
			input := ""
			res := innerGetNodeInfoByUniqueName(input)
			convey.So(res.Status, convey.ShouldNotEqual, common.Success)
		})
	})
}

func TestInnerGetNodeStatus(t *testing.T) {
	convey.Convey("InnerGetNodeStatus functional test", t, func() {
		convey.Convey("innerGetNodeInfoByUniqueName success", func() {
			node := &NodeInfo{
				NodeName:     "test-inner-node2",
				UniqueName:   "test-inner-node-unique-name2",
				SerialNumber: "test-inner-node-unique-serial-number2",
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
		})
		convey.Convey("innerGetNodeInfoByUniqueName param error", func() {
			input := ""
			res := innerGetNodeStatus(input)
			convey.So(res.Status, convey.ShouldNotEqual, common.Success)
		})
	})
}

func TestInnerGetNodeGroupInfosByIds(t *testing.T) {
	convey.Convey("InnerGetNodeGroupInfosByIds functional test", t, func() {
		convey.Convey("innerGetNodeInfoByUniqueName success", func() {
			group := &NodeGroup{
				GroupName: "test_inner_node_group_2_name",
				CreatedAt: time.Now().Format(TimeFormat),
				UpdatedAt: time.Now().Format(TimeFormat),
			}
			resGroup := env.createGroup(group)
			convey.So(resGroup, convey.ShouldBeNil)
			input := types.InnerGetNodeGroupInfosReq{NodeGroupIds: []uint64{group.ID}}
			res := innerGetNodeGroupInfosByIds(input)
			convey.So(res.Status, convey.ShouldEqual, common.Success)
		})
		convey.Convey("innerGetNodeInfoByUniqueName param error", func() {
			input := types.InnerGetNodeGroupInfosReq{NodeGroupIds: []uint64{0}}
			res := innerGetNodeGroupInfosByIds(input)
			convey.So(res.Status, convey.ShouldNotEqual, common.Success)
		})
	})
}
