// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nodemanager test about node
package nodemanager

import (
	"encoding/json"
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
	"gorm.io/gorm"
	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-manager/pkg/config"
	"edge-manager/pkg/constants"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
)

const (
	defaultPageSize     = 20
	shuffledNumberCount = 10000
	errPageSize         = 200
)

var env environment
var errTest = errors.New("error test")

type environment struct {
	shuffledNumbers     []int
	shuffledNumbersLock sync.Mutex
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

func TestGetNodeDetail(t *testing.T) {
	convey.Convey("getNodeDetail should be success", t, testGetNodeDetail)
	convey.Convey("getNodeDetail should be failed, input error", t, testGetNodeDetailErrInput)
	convey.Convey("getNodeDetail should be failed, param error", t, testGetNodeDetailErrParam)
	convey.Convey("getNodeDetail should be failed, id error", t, testGetNodeDetailErrGetNodeByID)
	convey.Convey("getNodeDetail should be failed, eval node group error", t, testGetNodeDetailErrEvalNodeGroup)
}

func testGetNodeDetail() {
	node := &NodeInfo{
		Description:  "test-node-description-1",
		NodeName:     "test-node-name-1",
		UniqueName:   "test-node-unique-name-1",
		SerialNumber: "test-node-serial-number-1",
		IP:           "0.0.0.0",
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		msg := model.Message{}
		err := msg.FillContent(map[string]interface{}{
			constants.KeySymbol:   constants.IdKey,
			constants.ValueSymbol: node.ID,
		})
		convey.So(err, convey.ShouldBeNil)
		resp := getNodeDetail(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		nodeInfoDetail, ok := resp.Data.(NodeInfoDetail)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(nodeInfoDetail.NodeInfoEx.NodeInfo, convey.ShouldResemble, *node)
	})
}

func testGetNodeDetailErrInput() {
	msg := model.Message{}
	err := msg.FillContent("1")
	convey.So(err, convey.ShouldBeNil)
	resp := getNodeDetail(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
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
	msg := model.Message{}
	err := msg.FillContent(map[string]interface{}{
		constants.KeySymbol:   constants.IdKey,
		constants.ValueSymbol: uint64(0),
	})
	convey.So(err, convey.ShouldBeNil)
	resp := getNodeDetail(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testGetNodeDetailErrGetNodeByID() {
	const testNode = 20
	msg := model.Message{}
	err := msg.FillContent(map[string]interface{}{
		constants.KeySymbol:   constants.IdKey,
		constants.ValueSymbol: uint64(testNode),
	})
	convey.So(err, convey.ShouldBeNil)
	resp := getNodeDetail(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNode)
}

func testGetNodeDetailErrEvalNodeGroup() {
	convey.Convey("error get relations by id", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getRelationsByNodeID",
			func(n *NodeServiceImpl, id uint64) (*[]NodeRelation, error) {
				return nil, test.ErrTest
			})
		defer p1.Reset()
		msg := model.Message{}
		err := msg.FillContent(map[string]interface{}{
			constants.KeySymbol:   constants.IdKey,
			constants.ValueSymbol: uint64(1),
		})
		convey.So(err, convey.ShouldBeNil)
		resp := getNodeDetail(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNode)
	})
}

func TestModifyNode(t *testing.T) {
	convey.Convey("modifyNode should be success", t, testModifyNode)
	convey.Convey("modifyNode should be success, test description", t, testModifyNodeDescription)
	convey.Convey("modifyNode should be failed", t, testModifyNodeErr)
}

func testModifyNode() {
	node := &NodeInfo{
		Description:  "test-node-description-2",
		NodeName:     "test-node-name-2",
		UniqueName:   "test-node-unique-name-2",
		SerialNumber: "test-node-serial-number-2",
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
		msg := model.Message{}
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := modifyNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node, "UpdatedAt")
		convey.So(verifyRes, convey.ShouldBeNil)
	})
}

func testModifyNodeDescription() {
	node := &NodeInfo{
		Description:  "test-node-description-19",
		NodeName:     "test-node-name-19",
		UniqueName:   "test-node-unique-name-19",
		SerialNumber: "test-node-serial-number-19",
		IsManaged:    true,
		IP:           "0.0.0.0",
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	msg := model.Message{}
	convey.Convey("test description will not be modified when description is not set", func() {
		node.NodeName = "test-modify-node-1-name-modified-20"
		args := fmt.Sprintf(`
			{
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.NodeName, node.ID)
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)

		resp := modifyNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node, "UpdatedAt")
		convey.So(verifyRes, convey.ShouldBeNil)
	})

	convey.Convey("test description will be modified when description is set to empty string", func() {
		node.Description = ""
		args := fmt.Sprintf(`
			{
                "description": "%s",
    			"nodeName": "%s",
    			"nodeID": %d
			}`, node.Description, node.NodeName, node.ID)
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)

		resp := modifyNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node, "UpdatedAt")
		convey.So(verifyRes, convey.ShouldBeNil)
	})
}

func testModifyNodeErr() {
	node := &NodeInfo{
		Description:  "test-node-description-3-#{random}",
		NodeName:     "test-node-name-3-#{random}",
		UniqueName:   "test-node-unique-name-3-#{random}",
		SerialNumber: "test-node-serial-number-3-#{random}",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	env.randomize(node)
	msg := model.Message{}

	convey.Convey("input error", func() {
		err := msg.FillContent(``)
		convey.So(err, convey.ShouldBeNil)
		resp := modifyNode(&msg)
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
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := modifyNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("empty description", func() {
		nodeRes := env.createNode(node)
		convey.So(nodeRes, convey.ShouldBeNil)
		args := fmt.Sprintf(`{"nodeName": "%s", "nodeID": %d}`, node.NodeName, node.ID)
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := modifyNode(&msg)
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
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := modifyNode(&msg)
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
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := modifyNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorModifyNode)
	})
}

func TestGetNodeStatistics(t *testing.T) {
	convey.Convey("getNodeStatistics should be success", t, testGetNodeStatistics)
	convey.Convey("getNodeStatistics should be failed, list node error", t, testGetNodeStatisticsErr)
}

func testGetNodeStatistics() {
	convey.Convey("normal input", func() {
		msg := model.Message{}
		err := msg.FillContent(``)
		convey.So(err, convey.ShouldBeNil)
		resp := getNodeStatistics(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func testGetNodeStatisticsErr() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listNodes",
		func(n *NodeServiceImpl) (*[]NodeInfo, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()
	msg := model.Message{}
	err := msg.FillContent(``)
	convey.So(err, convey.ShouldBeNil)
	resp := getNodeStatistics(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCountNodeByStatus)
}

func TestAddUnManagedNode(t *testing.T) {
	convey.Convey("addUnmanagedNode should be success", t, testAddUnManagedNode)
	convey.Convey("addUnmanagedNode should be failed", t, testAddUnManagedNodeErr)
}

func testAddUnManagedNode() {
	node := &NodeInfo{
		Description:  "test-node-description-4",
		NodeName:     "test-node-name-4",
		UniqueName:   "test-node-unique-name-4",
		SerialNumber: "test-node-serial-number-4",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	resNode := env.createNode(node)
	convey.So(resNode, convey.ShouldBeNil)

	group := &NodeGroup{
		Description: "test-group-description-17",
		GroupName:   "test_group_name_17",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdatedAt:   time.Now().Format(TimeFormat),
	}
	resGroup := env.createGroup(group)
	convey.So(resGroup, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
            "groupIDs": [%d],
            "nodeID": %d
			}`, node.NodeName, node.Description, group.ID, node.ID)
		p := gomonkey.ApplyFuncReturn(checkNodeBeforeAddToGroup, nil).
			ApplyFuncReturn(getRequestItemsOfAddGroup, nil, int64(0), nil)
		defer p.Reset()
		msg := model.Message{}
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := addUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		node.IsManaged = true
		convey.So(env.verifyDbNodeInfo(node, "CreatedAt", "UpdatedAt"), convey.ShouldBeNil)
		relation := &NodeRelation{NodeID: node.ID, GroupID: group.ID}
		convey.So(env.verifyDbNodeRelation(relation, "CreatedAt"), convey.ShouldBeNil)
	})
}

func getTestNodeInfo() *NodeInfo {
	return &NodeInfo{
		Description:  "test-node-description-5-#{random}",
		NodeName:     "test-node-name-5-#{random}",
		UniqueName:   "test-node-unique-name-5-#{random}",
		SerialNumber: "test-node-serial-number-5-#{random}",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
}

func testAddUnManagedNodeErr() {
	node := getTestNodeInfo()
	env.randomize(node)
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)
	msg := model.Message{}

	testAddUnManagedNodeErrInvalidParam(node)

	convey.Convey("groupIDs is not exist", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
            "nodeID": %d
			}`, node.NodeName, node.Description, uint64(20))
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := addUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorAddUnManagedNode)
	})

	convey.Convey("groupIDs is empty", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
			"groupIDs": [],
            "nodeID": %d
			}`, node.NodeName, node.Description, node.ID)
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := addUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func testAddUnManagedNodeErrInvalidParam(node *NodeInfo) {
	msg := model.Message{}
	convey.Convey("input error", func() {
		err := msg.FillContent("")
		convey.So(err, convey.ShouldBeNil)
		resp := addUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("invalid param: node name is not exist", func() {
		args := fmt.Sprintf(`{
			"name": "",
            "description": "%s",
            "groupIDs": [],
            "nodeID": %d
			}`, node.Description, node.ID)
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := addUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("invalid param: the size of groupIDs is greater than MaxGroupPerNode", func() {
		args := fmt.Sprintf(`{
			"name": "%s",
            "description": "%s",
            "groupIDs": [1,2,3,4,5,6,7,8,9,10,11],
            "nodeID": %d
			}`, node.NodeName, node.Description, node.ID)
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := addUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
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
		msg := model.Message{}
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := listManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		resp = listUnmanagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func testListManagedAndUnmanagedErrInput() {
	msg := model.Message{}
	err := msg.FillContent("")
	convey.So(err, convey.ShouldBeNil)
	resp := listManagedNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	resp = listUnmanagedNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testListManagedAndUnmanagedErrParam() {
	args := types.ListReq{PageNum: 1, PageSize: errPageSize}
	msg := model.Message{}
	err := msg.FillContent(args)
	convey.So(err, convey.ShouldBeNil)
	resp := listManagedNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	resp = listUnmanagedNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testListManagedAndUnmanagedErrCount() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "countNodesByName",
		func(n *NodeServiceImpl, name string, isManaged int) (int64, error) {
			return 0, test.ErrTest
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	msg := model.Message{}
	err := msg.FillContent(args)
	convey.So(err, convey.ShouldBeNil)
	resp := listManagedNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
	resp = listUnmanagedNode(&msg)
	convey.So(resp.Msg, convey.ShouldEqual, "count node failed")
}

func testListManagedAndUnmanagedErrList() {
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	msg := model.Message{}
	err := msg.FillContent(args)
	convey.So(err, convey.ShouldBeNil)

	var c1 *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c1), "listManagedNodesByName",
		func(n *NodeServiceImpl, page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()

	resp := listManagedNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)

	var c2 *NodeServiceImpl
	var p2 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c2), "listUnManagedNodesByName",
		func(n *NodeServiceImpl, page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
			return nil, test.ErrTest
		})
	defer p2.Reset()
	resp = listUnmanagedNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListUnManagedNode)
}

func testListManagedNodeErrEvalNodeGroup() {
	convey.Convey("error get relations by id", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getRelationsByNodeID",
			func(n *NodeServiceImpl, id uint64) (*[]NodeRelation, error) {
				return nil, test.ErrTest
			})
		defer p1.Reset()
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		msg := model.Message{}
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := listManagedNode(&msg)
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
		msg := model.Message{}
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := listNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

func testListNodeErrInput() {
	msg := model.Message{}
	err := msg.FillContent("")
	convey.So(err, convey.ShouldBeNil)
	resp := listNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testListNodeErrParam() {
	args := types.ListReq{PageNum: 1, PageSize: errPageSize}
	msg := model.Message{}
	err := msg.FillContent(args)
	convey.So(err, convey.ShouldBeNil)
	resp := listNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testListNodeErrCount() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "countAllNodesByName",
		func(n *NodeServiceImpl, name string) (int64, error) {
			return 0, test.ErrTest
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	msg := model.Message{}
	err := msg.FillContent(args)
	convey.So(err, convey.ShouldBeNil)
	resp := listNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
}

func testListNodeErrList() {
	var c *NodeServiceImpl
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "listAllNodesByName",
		func(n *NodeServiceImpl, page, pageSize uint64, nodeName string) (*[]NodeInfo, error) {
			return nil, test.ErrTest
		})
	defer p1.Reset()
	args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
	msg := model.Message{}
	err := msg.FillContent(args)
	convey.So(err, convey.ShouldBeNil)
	resp := listNode(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
}

func testListNodeErrEvalNodeGroup() {
	convey.Convey("error get relations by id", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getRelationsByNodeID",
			func(n *NodeServiceImpl, id uint64) (*[]NodeRelation, error) {
				return nil, test.ErrTest
			})
		defer p1.Reset()
		args := types.ListReq{PageNum: 1, PageSize: defaultPageSize}
		msg := model.Message{}
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := listNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorListNode)
	})
}

func TestBatchDeleteNode(t *testing.T) {
	var p1 = gomonkey.ApplyFunc(sendDeleteNodeMessageToNode, func(s string) error {
		return nil
	})
	defer p1.Reset()
	convey.Convey("batchDeleteNode should be success", t, testBatchDeleteNode)
	convey.Convey("batchDeleteNode should be failed", t, testBatchDeleteNodeErr)
}

func testBatchDeleteNode() {
	node := &NodeInfo{
		Description:  "test-node-description-6",
		NodeName:     "test-node-name-6",
		UniqueName:   "test-node-unique-name-6",
		SerialNumber: "test-node-serial-number-6",
		IP:           "0.0.0.0",
		IsManaged:    true,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`{"nodeIDs": [%d]}`, node.ID)
		msg := model.Message{}
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := batchDeleteNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node)
		convey.So(verifyRes, convey.ShouldNotEqual, "record not found")
	})
}

func testBatchDeleteNodeErr() {
	msg := model.Message{}

	convey.Convey("empty request", func() {
		err := msg.FillContent("")
		convey.So(err, convey.ShouldBeNil)
		resp := batchDeleteNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("param error", func() {
		args := `{"nodeIDs": []}`
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := batchDeleteNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("delete node id error", func() {
		args := `{"nodeIDs": [20]}`
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := batchDeleteNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteNode)
	})
}

func TestDeleteUnManagedNode(t *testing.T) {
	var p1 = gomonkey.ApplyFunc(sendDeleteNodeMessageToNode, func(s string) error {
		return nil
	})
	defer p1.Reset()
	convey.Convey("deleteUnManagedNode should be success", t, testDeleteUnManagedNode)
	convey.Convey("deleteUnManagedNode should be failed", t, testDeleteUnManagedNodeErr)
}

func testDeleteUnManagedNode() {
	node := &NodeInfo{
		Description:  "test-node-description-7",
		NodeName:     "test-node-name-7",
		UniqueName:   "test-node-unique-name-7",
		SerialNumber: "test-node-serial-number-7",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	convey.Convey("normal input", func() {
		args := fmt.Sprintf(`{"nodeIDs": [%d]}`, node.ID)
		msg := model.Message{}
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := deleteUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		verifyRes := env.verifyDbNodeInfo(node)
		convey.So(verifyRes, convey.ShouldNotEqual, "record not found")
	})
}

func testDeleteUnManagedNodeErr() {
	msg := model.Message{}
	convey.Convey("empty request", func() {
		err := msg.FillContent("")
		convey.So(err, convey.ShouldBeNil)
		resp := deleteUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("param error", func() {
		args := `{"nodeIDs": []}`
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := deleteUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
	})

	convey.Convey("delete node id error", func() {
		args := `{"nodeIDs": [20]}`
		err := msg.FillContent(args)
		convey.So(err, convey.ShouldBeNil)
		resp := deleteUnManagedNode(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteNode)
	})
}

func TestUpdateNodeSoftwareInfo(t *testing.T) {
	convey.Convey("updateNodeSoftwareInfo should be success", t, testUpdateNodeSoftwareInfo)
	convey.Convey("updateNodeSoftwareInfo should be failed", t, testUpdateNodeSoftwareInfoErr)
}

func testUpdateNodeSoftwareInfo() {
	node := &NodeInfo{
		Description:  "test-node-description-8",
		NodeName:     "test-node-name-8",
		UniqueName:   "test-node-unique-name-8",
		SerialNumber: "test-node-serial-number-8",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}

	sfwInfo := types.SoftwareInfo{
		InactiveVersion: "v1.12",
		Name:            "edgecore",
		Version:         "v1.12",
	}

	req := types.EdgeReportSoftwareInfoReq{
		SerialNumber: "test-update-node-software-info-2-serial-number",
		SoftwareInfo: []types.SoftwareInfo{sfwInfo},
	}
	reqByte, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("marshal req failed, error: %v", err)
	}
	msg := model.Message{}
	err = msg.FillContent(reqByte)
	convey.So(err, convey.ShouldBeNil)
	msg.SetPeerInfo(model.MsgPeerInfo{Sn: "test-update-node-software-info-2-serial-number"})
	resp := updateNodeSoftwareInfo(&msg)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)

	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	req = types.EdgeReportSoftwareInfoReq{
		SerialNumber: node.SerialNumber,
		SoftwareInfo: []types.SoftwareInfo{sfwInfo},
	}
	reqByte, err = json.Marshal(req)
	if err != nil {
		fmt.Printf("marshal req failed, error: %v", err)
	}
	convey.So(err, convey.ShouldBeNil)
	err = msg.FillContent(reqByte)
	convey.So(err, convey.ShouldBeNil)
	msg.SetPeerInfo(model.MsgPeerInfo{Sn: node.SerialNumber})
	resp = updateNodeSoftwareInfo(&msg)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func getUpdateNodeSftErrReqMsg() *model.Message {
	node := &NodeInfo{
		Description:  "test-node-description-9",
		NodeName:     "test-node-name-9",
		UniqueName:   "test-node-unique-name-9",
		SerialNumber: "test-node-serial-number-9",
		IP:           "0.0.0.0",
		IsManaged:    false,
		CreatedAt:    time.Now().Format(TimeFormat),
		UpdatedAt:    time.Now().Format(TimeFormat),
	}
	res := env.createNode(node)
	convey.So(res, convey.ShouldBeNil)

	sfwInfo := types.SoftwareInfo{
		InactiveVersion: "v1.12",
		Name:            "edgecore",
		Version:         "v1.12"}

	req := types.EdgeReportSoftwareInfoReq{
		SerialNumber: node.SerialNumber,
		SoftwareInfo: []types.SoftwareInfo{sfwInfo},
	}
	reqByte, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("marshal req failed, error: %v", err)
	}
	msg := model.Message{}
	msg.SetPeerInfo(model.MsgPeerInfo{
		Sn: node.SerialNumber,
	})
	err = msg.FillContent(reqByte)
	convey.So(err, convey.ShouldBeNil)
	return &msg
}

func testUpdateNodeSoftwareInfoErr() {
	convey.Convey("empty request", func() {
		msg := model.Message{}
		err := msg.FillContent("")
		convey.So(err, convey.ShouldBeNil)
		resp := updateNodeSoftwareInfo(&msg)
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
	})

	convey.Convey("marshal error", func() {
		var p1 = gomonkey.ApplyFunc(json.Marshal,
			func(v interface{}) ([]byte, error) {
				return nil, test.ErrTest
			})
		defer p1.Reset()
		resp := updateNodeSoftwareInfo(getUpdateNodeSftErrReqMsg())
		convey.So(resp.Msg, convey.ShouldEqual, "marshal version info failed")
	})

	convey.Convey("getNodeInfoBySerialNumber error is not [record not found]", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getNodeInfoBySerialNumber",
			func(n *NodeServiceImpl, name string) (*NodeInfo, error) {
				return nil, test.ErrTest
			})
		defer p1.Reset()

		resp := updateNodeSoftwareInfo(getUpdateNodeSftErrReqMsg())
		convey.So(resp.Msg, convey.ShouldEqual, "get node info failed")
	})

	convey.Convey("update error", func() {
		var c *NodeServiceImpl
		var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "updateNodeInfoBySerialNumber",
			func(n *NodeServiceImpl, sn string, nodeInfo *NodeInfo) error {
				return test.ErrTest
			})
		defer p1.Reset()

		resp := updateNodeSoftwareInfo(getUpdateNodeSftErrReqMsg())
		convey.So(resp.Msg, convey.ShouldEqual, "update node software info failed")
	})
}

// TestSendDeleteNodeMessageToNode tests the functionality of sending a delete node message
func TestSendDeleteNodeMessageToNode(t *testing.T) {
	convey.Convey("send delete message should be success", t, testSendDeleteNodeMessageToNode)
	convey.Convey("send delete message should be failed", t, testSendDeleteNodeMessageToNodeErr)
}

func testSendDeleteNodeMessageToNode() {
	newMsg := model.Message{}
	convey.Convey("send delete message", func() {
		patches := gomonkey.ApplyFuncReturn(model.NewMessage, &newMsg, nil).
			ApplyMethodReturn(&model.Message{}, "FillContent", nil).
			ApplyFuncReturn(modulemgr.SendMessage, nil)
		defer patches.Reset()
		serialNumber := "123"
		resp := sendDeleteNodeMessageToNode(serialNumber)
		convey.So(resp, convey.ShouldBeNil)
	})
}

func testSendDeleteNodeMessageToNodeErr() {
	newMsg := model.Message{}
	convey.Convey("send delete message create new message failed", func() {
		patch := gomonkey.ApplyFuncReturn(model.NewMessage, &newMsg, errTest)
		defer patch.Reset()
		convey.So(sendDeleteNodeMessageToNode("123"), convey.ShouldResemble,
			fmt.Errorf("create new message failed, error: %v", errTest))
	})

	convey.Convey("send delete message fill content failed", func() {
		patches := gomonkey.ApplyFuncReturn(model.NewMessage, &newMsg, nil).
			ApplyMethodReturn(&model.Message{}, "FillContent", errTest)
		defer patches.Reset()
		convey.So(sendDeleteNodeMessageToNode("123"), convey.ShouldResemble,
			fmt.Errorf("fill content failed: %v", errTest))
	})

	convey.Convey("send delete message failed", func() {
		patches := gomonkey.ApplyFuncReturn(model.NewMessage, &newMsg, nil).
			ApplyMethodReturn(&model.Message{}, "FillContent", nil).
			ApplyFuncReturn(modulemgr.SendMessage, errTest)
		defer patches.Reset()
		convey.So(sendDeleteNodeMessageToNode("123"), convey.ShouldResemble,
			fmt.Errorf("%s sends message to %s failed, error: %v",
				common.NodeManagerName, common.CloudHubName, errTest))
	})
}

// TestGetRequestItemsOfAddGroup tests the functionality of getting request items for adding a group
func TestGetRequestItemsOfAddGroup(t *testing.T) {
	convey.Convey("get request items of add group should be success", t, testGetRequestItemsOfAddGroup)
	convey.Convey("get request items of add group should be failed", t, testGetRequestItemsOfAddGroupErr)
}

func testGetRequestItemsOfAddGroup() {
	var allocatedRes v1.ResourceList
	var count int64
	var nodeGroup NodeGroup
	var err error

	convey.Convey("getRequestItemsOfAddGroup should be success", func() {
		patch := gomonkey.ApplyFuncReturn(getNodeGroupResReq, allocatedRes, nil).
			ApplyFuncReturn(getAppInstanceCountByGroupId, count, nil)
		defer patch.Reset()
		allocatedRes, count, err = getRequestItemsOfAddGroup(&nodeGroup)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testGetRequestItemsOfAddGroupErr() {
	var allocatedRes v1.ResourceList
	var nodeGroup NodeGroup
	var err error
	var count int64

	convey.Convey("getRequestItemsOfAddGroup getNodeGroupResReq error", func() {
		patch := gomonkey.ApplyFuncReturn(getNodeGroupResReq, allocatedRes, errTest)
		defer patch.Reset()
		allocatedRes, count, err = getRequestItemsOfAddGroup(&nodeGroup)
		convey.So(err, convey.ShouldResemble, errTest)
	})

	convey.Convey("getRequestItemsOfAddGroup count request error", func() {
		patch := gomonkey.ApplyFuncReturn(getNodeGroupResReq, allocatedRes, nil).
			ApplyFuncReturn(getAppInstanceCountByGroupId, count, errTest)
		defer patch.Reset()
		allocatedRes, count, err = getRequestItemsOfAddGroup(&nodeGroup)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("get node group id [%d] deployed app count request error", nodeGroup.ID))
	})
}

// TestGetNodeDetailBySn get node details for success and fail cases
func TestGetNodeDetailBySn(t *testing.T) {
	convey.Convey("Given a getNodeDetailBySn function success", t, testGetNodeDetailBySnSuccess)
	convey.Convey("Given a getNodeDetailBySn function err input", t, testGetNodeDetailBySnErrInput)
	convey.Convey("Given a getNodeDetailBySn function err query db", t, testGetNodeDetailBySnErrQueryDb)
	convey.Convey("Given a getNodeDetailBySn function err ext", t, testGetNodeDetailBySnErrExt)
}

func testGetNodeDetailBySnErrInput() {
	convey.Convey("When the input is not a string", func() {
		resp := getNodeDetailBySn(123)

		convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
		convey.So(resp.Msg, convey.ShouldEqual, "query node detail convert param failed")
	})
}

func testGetNodeDetailBySnErrQueryDb() {
	convey.Convey("When getNodeBySn returns an error", func() {
		patches := gomonkey.ApplyFunc(checker.GetSnChecker, func(sn string, flag bool) *checker.RegChecker {
			return &checker.RegChecker{}
		})
		defer patches.Reset()

		patches.ApplyFunc(setNodeExtInfos, func(resp NodeInfoDetail) (NodeInfoDetail, error) {
			return NodeInfoDetail{}, nil
		})
		patches.ApplyPrivateMethod(NodeServiceInstance(), "getNodeBySn",
			func(sn string) (*NodeInfo, error) {
				return nil, fmt.Errorf("db query error")
			},
		)

		resp := getNodeDetailBySn("valid-sn")

		convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNode)
		convey.So(resp.Msg, convey.ShouldEqual, "query node in db error")
	})
}

func testGetNodeDetailBySnErrExt() {
	convey.Convey("When setNodeExtInfos returns an error", func() {
		patches := gomonkey.ApplyFunc(checker.GetSnChecker, func(sn string, flag bool) *checker.RegChecker {
			return &checker.RegChecker{}
		})
		defer patches.Reset()

		patches.ApplyPrivateMethod(NodeServiceInstance(), "getNodeBySn",
			func(sn string) (*NodeInfo, error) {
				return &NodeInfo{}, nil
			},
		)
		patches.ApplyFunc(setNodeExtInfos, func(resp NodeInfoDetail) (NodeInfoDetail, error) {
			return NodeInfoDetail{}, fmt.Errorf("ext info error")
		})

		resp := getNodeDetailBySn("valid-sn")

		convey.So(resp.Status, convey.ShouldEqual, common.ErrorGetNode)
		convey.So(resp.Msg, convey.ShouldEqual, "ext info error")
	})
}

func testGetNodeDetailBySnSuccess() {
	convey.Convey("When all operations are successful", func() {
		patches := gomonkey.ApplyFuncReturn(checker.GetSnChecker, &checker.RegChecker{})
		patches.ApplyPrivateMethod(NodeServiceInstance(), "getNodeBySn",
			func(sn string) (*NodeInfo, error) {
				return &NodeInfo{}, nil
			},
		)
		patches.ApplyFuncReturn(setNodeExtInfos, NodeInfoDetail{}, nil)

		defer patches.Reset()
		resp := getNodeDetailBySn("valid-sn")

		convey.So(resp.Status, convey.ShouldEqual, common.Success)
	})
}

// TestCheckNodePodLimit test check node pod limit for success and fail cases
func TestCheckNodePodLimit(t *testing.T) {
	convey.Convey("Given a checkNodePodLimit function success", t, testCheckNodePodLimitSuccess)
	convey.Convey("Given a checkNodePodLimit function for limit node error", t, testCheckNodePodLimitNodeErr)
	convey.Convey("Given a checkNodePodLimit function for app count error", t, testCheckNodePodLimitAppCountErr)
	convey.Convey("Given a checkNodePodLimit function for out of max error", t, testCheckNodePodLimitOutOfMaxErr)
}

func testCheckNodePodLimitSuccess() {
	convey.Convey("When the total pod count is within the limit", func() {
		var maxPodNumber int64 = 100
		patches := gomonkey.ApplyPrivateMethod(NodeServiceInstance(), "getGroupsByNodeID",
			func(nodeId uint64) (*[]NodeGroup, error) {
				return &[]NodeGroup{
					{ID: 1},
					{ID: 2},
				}, nil
			},
		)
		patches.ApplyFunc(getAppInstanceCountByGroupId, func(groupId uint64) (int64, error) {
			return 5, nil
		})
		patches.ApplyGlobalVar(&config.PodConfig.MaxPodNumberPerNode, maxPodNumber)

		defer patches.Reset()
		err := checkNodePodLimit(5, 1)

		convey.So(err, convey.ShouldBeNil)
	})
}

func testCheckNodePodLimitNodeErr() {
	convey.Convey("When getGroupsByNodeID returns an error", func() {
		patches := gomonkey.ApplyPrivateMethod(NodeServiceInstance(), "getGroupsByNodeID",
			func(nodeId uint64) (*[]NodeGroup, error) {
				return nil, fmt.Errorf("db query error")
			},
		)
		patches.ApplyFunc(getAppInstanceCountByGroupId, func(groupId uint64) (int64, error) {
			return 5, nil
		})

		defer patches.Reset()
		err := checkNodePodLimit(10, 1)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "get node groups by node id [1] error")
	})
}

func testCheckNodePodLimitAppCountErr() {
	convey.Convey("When getAppInstanceCountByGroupId returns an error", func() {
		patches := gomonkey.ApplyPrivateMethod(NodeServiceInstance(), "getGroupsByNodeID",
			func(nodeId uint64) (*[]NodeGroup, error) {
				return &[]NodeGroup{
					{ID: 1},
				}, nil
			},
		)
		patches.ApplyFunc(getAppInstanceCountByGroupId, func(groupId uint64) (int64, error) {
			return 0, fmt.Errorf("count error")
		})

		defer patches.Reset()
		err := checkNodePodLimit(10, 1)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "get deployed app count by node group id [1] error")
	})
}

func testCheckNodePodLimitOutOfMaxErr() {
	convey.Convey("When the total pod count exceeds the limit", func() {
		var maxPodNumber int64 = 20
		patches := gomonkey.ApplyPrivateMethod(NodeServiceInstance(), "getGroupsByNodeID",
			func(nodeId uint64) (*[]NodeGroup, error) {
				return &[]NodeGroup{
					{ID: 3},
				}, nil
			},
		)
		patches.ApplyFunc(getAppInstanceCountByGroupId, func(groupId uint64) (int64, error) {
			return 5, nil
		})
		patches.ApplyGlobalVar(&config.PodConfig.MaxPodNumberPerNode, maxPodNumber)

		defer patches.Reset()
		err := checkNodePodLimit(100, 3)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "pod addedNumber is out of node [3] max allowed addedNumber")
	})
}

// TestCheckNodeResource test check node resource for failed errors
func TestCheckNodeResource(t *testing.T) {
	convey.Convey("test check node resource get node error", t, testCheckNodeResourceGetNodeErr)
	convey.Convey("test check node resource do not have enough cpu resources", t, testCheckNodeResourceCpuShort)
	convey.Convey("test check node resource do not have enough memory resources", t, testCheckNodeResourceMemShort)
	convey.Convey("test check node resource do not have enough npu resources", t, testCheckNodeResourceNpuShort)
}

func testCheckNodeResourceGetNodeErr() {
	convey.Convey("When getNodeByID returns an error", func() {
		patches := gomonkey.ApplyPrivateMethod(NodeServiceInstance(), "getNodeByID",
			func(nodeId uint64) (*NodeInfo, error) {
				return nil, fmt.Errorf("db query error")
			},
		)
		defer patches.Reset()

		req := v1.ResourceList{
			v1.ResourceCPU:    *resource.NewQuantity(5, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(50, resource.DecimalSI),
		}

		err := checkNodeResource(req, 1)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "get node info by node id [1] error")
	})
}

func testCheckNodeResourceCpuShort() {
	convey.Convey("When the node does not have enough CPU resources", func() {
		patches := gomonkey.ApplyPrivateMethod(NodeServiceInstance(), "getNodeByID",
			func(nodeId uint64) (*NodeInfo, error) {
				return &NodeInfo{
					ID:         1,
					UniqueName: "test-node",
				}, nil
			},
		)
		defer patches.Reset()

		req := v1.ResourceList{
			v1.ResourceCPU:    *resource.NewQuantity(15, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(50, resource.DecimalSI),
		}

		patches.ApplyMethodReturn(&nodeSyncImpl{}, "GetAvailableResource", &NodeResource{
			Cpu: *resource.NewQuantity(10, resource.DecimalSI),
		}, nil)

		err := checkNodeResource(req, 1)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "node [1] do not have enough cpu resources")
	})
}

func testCheckNodeResourceMemShort() {
	convey.Convey("When the node does not have enough memory resources", func() {
		patches := gomonkey.ApplyPrivateMethod(NodeServiceInstance(), "getNodeByID",
			func(nodeId uint64) (*NodeInfo, error) {
				return &NodeInfo{
					ID:         2,
					UniqueName: "test-node",
				}, nil
			},
		)
		defer patches.Reset()

		req := v1.ResourceList{
			v1.ResourceCPU:    *resource.NewQuantity(5, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(150, resource.DecimalSI),
		}

		patches.ApplyMethodReturn(&nodeSyncImpl{}, "GetAvailableResource", &NodeResource{
			Cpu:    *resource.NewQuantity(10, resource.DecimalSI),
			Memory: *resource.NewQuantity(100, resource.DecimalSI),
		}, nil)

		err := checkNodeResource(req, 2)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "node [2] do not have enough memory resources")
	})
}

func testCheckNodeResourceNpuShort() {
	convey.Convey("When the node does not have enough NPU resources", func() {
		patches := gomonkey.ApplyPrivateMethod(NodeServiceInstance(), "getNodeByID",
			func(nodeId uint64) (*NodeInfo, error) {
				return &NodeInfo{
					ID:         3,
					UniqueName: "test-node",
				}, nil
			},
		)
		defer patches.Reset()
		req := v1.ResourceList{
			v1.ResourceCPU:    *resource.NewQuantity(5, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(50, resource.DecimalSI),
			common.DeviceType: *resource.NewQuantity(10, resource.DecimalSI),
		}

		patches.ApplyMethodReturn(&nodeSyncImpl{}, "GetAvailableResource", &NodeResource{
			Cpu:    *resource.NewQuantity(10, resource.DecimalSI),
			Memory: *resource.NewQuantity(100, resource.DecimalSI),
			Npu:    *resource.NewQuantity(8, resource.DecimalSI),
		}, nil)

		err := checkNodeResource(req, 3)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "node [3] do not have enough npu resources")
	})
}
