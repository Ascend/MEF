// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager test about node
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
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"

	"edge-manager/pkg/database"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

const (
	memoryDsn           = ":memory:?cache=shared"
	defaultPageSize     = 20
	shuffledNumberCount = 10000
	errPageSize         = 200
)

var (
	env     environment
	testErr = errors.New("test error")
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
	if err = e.setupTables(db); err != nil {
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
	convey.Convey("getNodeDetail should be success", t, testGetNodeDetail)
	convey.Convey("getNodeDetail should be failed, input error", t, testGetNodeDetailErrInput)
	convey.Convey("getNodeDetail should be failed, param error", t, testGetNodeDetailErrParam)
	convey.Convey("getNodeDetail should be failed, id error", t, testGetNodeDetailErrGetNodeByID)
	convey.Convey("getNodeDetail should be failed, eval node group error", t, testGetNodeDetailErrEvalNodeGroup)
}

func testGetNodeDetail() {
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

func testGetNodeDetailErrInput() {
	resp := getNodeDetail("1")
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testGetNodeDetailErrParam() {
	var c *checker.UintChecker
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "Check",
		func(n *checker.UintChecker, data interface{}) checker.CheckResult {
			checkRes := checker.CheckResult{
				Result: false,
				Reason: "",
			}
			return checkRes
		})
	defer p1.Reset()
	resp := getNodeDetail(uint64(0))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testGetNodeDetailErrGetNodeByID() {
	resp := getNodeDetail(uint64(20))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNode)
}

func testGetNodeDetailErrEvalNodeGroup() {
	convey.Convey("error get relations by id", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getRelationsByNodeID",
			func(n *NodeServiceImpl, id uint64) (*[]NodeRelation, error) {
				return nil, testErr
			})
		defer p1.Reset()
		resp := getNodeDetail(uint64(1))
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNode)
	})
}

func TestModifyNode(t *testing.T) {
	convey.Convey("modifyNode should be success", t, testModifyNode)
	convey.Convey("modifyNode should be failed", t, testModifyNodeErr)
}

func testModifyNode() {
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

func testModifyNodeErr() {
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

	convey.Convey("input error", func() {
		resp := modifyNode(``)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("param error", func() {
		nodeRes := env.createNode(node)
		convey.So(nodeRes, convey.ShouldBeNil)
		node.Description = ""
		args := fmt.Sprintf(`
			{
    			"description": "%s",
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.Description, "", node.ID)
		resp := modifyNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("empty description", func() {
		nodeRes := env.createNode(node)
		convey.So(nodeRes, convey.ShouldBeNil)
		node.Description = ""
		args := fmt.Sprintf(`{"nodeName": "%s", "nodeID": %d}`, node.NodeName, node.ID)
		resp := modifyNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node, "UpdatedAt")
		convey.So(verifyRes, convey.ShouldBeNil)
	})

	convey.Convey("duplicate node name", func() {
		nodeRes := env.createNode(node)
		convey.So(nodeRes, convey.ShouldBeNil)
		args := fmt.Sprintf(`
			{
    			"description": "%s",
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.Description, "test-modify-node-1-name-modified", node.ID)
		resp := modifyNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorNodeMrgDuplicate)
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
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorModifyNode)
	})
}

func TestGetNodeStatistics(t *testing.T) {
	convey.Convey("getNodeStatistics should be success", t, testGetNodeStatistics)
	convey.Convey("getNodeStatistics should be failed, list node error", t, testGetNodeStatisticsErr)
}

func testGetNodeStatistics() {
	convey.Convey("normal input", func() {
		resp := getNodeStatistics(``)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func testGetNodeStatisticsErr() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listNodes",
		func(n *NodeServiceImpl) (*[]NodeInfo, error) {
			return nil, testErr
		})
	defer p1.Reset()
	resp := getNodeStatistics(``)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCountNodeByStatus)
}

func TestAddUnManagedNode(t *testing.T) {
	convey.Convey("addUnmanagedNode should be success", t, testAddUnManagedNode)
	convey.Convey("addUnmanagedNode should be failed", t, testAddUnManagedNodeErr)
}

func testAddUnManagedNode() {
	node := &NodeInfo{
		Description:  "test-add-node-1-description",
		NodeName:     "test-add-node-1-name",
		UniqueName:   "test-add-node-1-unique-name",
		SerialNumber: "test-add-node-1-serial-number",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	group := &NodeGroup{
		Description: "test-add-1-description",
		GroupName:   "test_add_1_name",
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

func testAddUnManagedNodeErr() {
	node := &NodeInfo{
		Description:  "test-add-#{random}-description",
		NodeName:     "test-add-#{random}-name",
		UniqueName:   "test-add-#{random}-unique-name",
		SerialNumber: "test-add-#{random}-serial-umber",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	env.randomize(node)
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("input error", func() {
		args := ``
		resp := addUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("invalid param: node name is not exist", func() {
		args := fmt.Sprintf(`{
			"name": "",
            "description": "%s",
            "groupIDs": [],
            "nodeID": %d
			}`, node.Description, node.ID)
		resp := addUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("groupIDs is not exist", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
            "nodeID": %d
			}`, node.NodeName, node.Description, uint64(20))
		resp := addUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorAddUnManagedNode)
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

func TestListManagedAndUnmanagedNode(t *testing.T) {
	convey.Convey("listManagedNode and listUnmanagedNode should be success", t, testListManagedAndUnmanagedNode)
	convey.Convey("listManagedNode and listUnmanagedNode failed, input error", t, testListManagedAndUnmanagedErrInput)
	convey.Convey("listManagedNode and listUnmanagedNode failed, param error", t, testListManagedAndUnmanagedErrParam)
	convey.Convey("listManagedNode and listUnmanagedNode failed, count error", t, testListManagedAndUnmanagedErrCount)
	convey.Convey("listManagedNode and listUnmanagedNode failed, list node error", t, testListManagedAndUnmanagedErrList)
	convey.Convey("listManagedNode should be failed, eval node group error", t, testListManagedNodeErrEvalNodeGroup)
}

func testListManagedAndUnmanagedNode() {
	convey.Convey("normal input", func() {
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		resp = listUnmanagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func testListManagedAndUnmanagedErrInput() {
	args := ""
	resp := listManagedNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
	resp = listUnmanagedNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testListManagedAndUnmanagedErrParam() {
	args := types.ListReq{PageNum: 1, PageSize: errPageSize}
	resp := listManagedNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	resp = listUnmanagedNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testListManagedAndUnmanagedErrCount() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "countNodesByName",
		func(n *NodeServiceImpl, name string, isManaged int) (int64, error) {
			return 0, testErr
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	resp := listManagedNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
	resp = listUnmanagedNode(args)
	convey.So(resp.Msg, convey.ShouldEqual, "count node failed")
}

func testListManagedAndUnmanagedErrList() {
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}

	var c1 *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c1), "listManagedNodesByName",
		func(n *NodeServiceImpl, page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
			return nil, testErr
		})
	defer p1.Reset()

	resp := listManagedNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)

	var c2 *NodeServiceImpl
	var p2 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c2), "listUnManagedNodesByName",
		func(n *NodeServiceImpl, page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
			return nil, testErr
		})
	defer p2.Reset()
	resp = listUnmanagedNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListUnManagedNode)
}

func testListManagedNodeErrEvalNodeGroup() {
	convey.Convey("error get relations by id", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getRelationsByNodeID",
			func(n *NodeServiceImpl, id uint64) (*[]NodeRelation, error) {
				return nil, testErr
			})
		defer p1.Reset()
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
	})
}

func TestListNode(t *testing.T) {
	convey.Convey("listNode should be success", t, testListNode)
	convey.Convey("listNode should be failed, input error", t, testListNodeErrInput)
	convey.Convey("listNode should be failed, param error", t, testListNodeErrParam)
	convey.Convey("listNode should be failed, count nodes error", t, testListNodeErrCount)
	convey.Convey("listNode should be failed, list nodes error", t, testListNodeErrList)
	convey.Convey("listNode should be failed, eval node group error", t, testListNodeErrEvalNodeGroup)
}

func testListNode() {
	convey.Convey("normal input", func() {
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func testListNodeErrInput() {
	args := ""
	resp := listNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testListNodeErrParam() {
	args := types.ListReq{PageNum: 1, PageSize: errPageSize}
	resp := listNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testListNodeErrCount() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "countAllNodesByName",
		func(n *NodeServiceImpl, name string) (int64, error) {
			return 0, testErr
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	resp := listNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
}

func testListNodeErrList() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listAllNodesByName",
		func(n *NodeServiceImpl, page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
			return nil, testErr
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	resp := listNode(args)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
}

func testListNodeErrEvalNodeGroup() {
	convey.Convey("error get relations by id", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getRelationsByNodeID",
			func(n *NodeServiceImpl, id uint64) (*[]NodeRelation, error) {
				return nil, testErr
			})
		defer p1.Reset()
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		resp := listNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
	})
}

func TestBatchDeleteNode(t *testing.T) {
	convey.Convey("batchDeleteNode should be success", t, testBatchDeleteNode)
	convey.Convey("batchDeleteNode should be failed", t, testBatchDeleteNodeErr)
}

func testBatchDeleteNode() {
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

func testBatchDeleteNodeErr() {
	convey.Convey("empty request", func() {
		args := ``
		resp := batchDeleteNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("param error", func() {
		args := `{"nodeIDs": []}`
		resp := batchDeleteNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("delete node id error", func() {
		args := `{"nodeIDs": [20]}`
		resp := batchDeleteNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteNode)
	})
}

func TestDeleteUnManagedNode(t *testing.T) {
	convey.Convey("deleteUnManagedNode should be success", t, testDeleteUnManagedNode)
	convey.Convey("deleteUnManagedNode should be failed", t, testDeleteUnManagedNodeErr)
}

func testDeleteUnManagedNode() {
	node := &NodeInfo{
		Description:  "test-delete-unmanaged-node-1-description",
		NodeName:     "test-delete-unmanaged-node-1-name",
		UniqueName:   "test-delete-unmanaged-node-1-unique-name",
		SerialNumber: "test-delete-unmanaged-node-1-serial-number",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`{"nodeIDs": [%d]}`, node.ID)
		resp := deleteUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node)
		convey.So(verifyRes, convey.ShouldNotEqual, "record not found")
	})
}

func testDeleteUnManagedNodeErr() {
	convey.Convey("empty request", func() {
		args := ``
		resp := deleteUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("param error", func() {
		args := `{"nodeIDs": []}`
		resp := deleteUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("delete node id error", func() {
		args := `{"nodeIDs": [20]}`
		resp := deleteUnManagedNode(args)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteNode)
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

func TestInnerGetNodeSoftwareInfo(t *testing.T) {
	convey.Convey("InnerGetNodeSoftwareInfo functional test", t, func() {
		convey.Convey("InnerGetNodeSoftwareInfo failed", func() {
			node := &NodeInfo{
				NodeName:     "test-inner-node-software",
				UniqueName:   "test-inner-node-unique-software",
				SerialNumber: "test-inner-node-serial-number-software",
				IP:           "0.0.0.0",
				IsManaged:    true,
				CreatedAt:    time.Now().Format(TimeFormat),
				UpdatedAt:    time.Now().Format(TimeFormat),
				SoftwareInfo: "",
			}
			resNode := env.createNode(node)
			convey.So(resNode, convey.ShouldBeNil)
			input := types.InnerGetSfwInfoBySNReq{SerialNumber: "test-inner-node-serial-number-software"}
			res := innerGetNodeSoftwareInfo(input)
			convey.So(res.Status, convey.ShouldEqual, "")
		})
	})
}

func TestInnerGetNodesByNodeGroupID(t *testing.T) {
	convey.Convey("InnerGetNodesByNodeGroupID functional test", t, func() {
		convey.Convey("InnerGetNodesByNodeGroupID success", func() {
			input := types.InnerGetNodesReq{NodeGroupID: 1}
			res := innerGetNodesByNodeGroupID(input)
			convey.So(res.Status, convey.ShouldEqual, common.Success)
		})
	})
}

func TestInnerAllNodeInfos(t *testing.T) {
	convey.Convey("innerAllNodeInfos functional test", t, func() {
		convey.Convey("innerAllNodeInfos success", func() {
			res := innerAllNodeInfos("")
			convey.So(res.Status, convey.ShouldEqual, common.Success)
		})
	})
}

func TestInnerCheckNodeGroupResReq(t *testing.T) {
	convey.Convey("innerAllNodeInfos functional test", t, func() {
		convey.Convey("innerAllNodeInfos success", func() {
			input := types.InnerCheckNodeResReq{NodeGroupID: 1}
			res := innerCheckNodeGroupResReq(input)
			convey.So(res.Status, convey.ShouldEqual, common.Success)
		})
	})
}

func TestGetAppInstanceCountByGroupId(t *testing.T) {
	convey.Convey("getAppInstanceCountByGroupId functional test", t, func() {
		convey.Convey("getAppInstanceCountByGroupId success", func() {
			counts := map[uint64]int64{1: 1}
			gomonkey.ApplyFuncReturn(common.SendSyncMessageByRestful,
				common.RespMsg{Status: common.Success, Data: counts})
			_, err := getAppInstanceCountByGroupId(1)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
