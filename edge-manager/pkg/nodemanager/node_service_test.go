// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"edge-manager/pkg/database"
	"edge-manager/pkg/util"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindxedge/base/common"
)

var (
	gormInstance *gorm.DB
	dbPath       = "./test.db"
)

func setup() {
	var err error
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err = common.InitHwlogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
	}
	if err = os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		hwlog.RunLog.Errorf("cleanup db failed, error: %v", err)
	}
	gormInstance, err = gorm.Open(sqlite.Open(dbPath))
	if err != nil {
		hwlog.RunLog.Errorf("failed to init test db, %v\n", err)
	}
	if err = gormInstance.AutoMigrate(&NodeInfo{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v\n", err)
	}
	if err = gormInstance.AutoMigrate(&NodeRelation{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v\n", err)
	}
	if err = gormInstance.AutoMigrate(&NodeGroup{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v\n", err)
	}
}

func teardown() {

}

func mockGetDb() *gorm.DB {
	return gormInstance
}

func TestMain(m *testing.M) {
	patches := gomonkey.ApplyFunc(database.GetDb, mockGetDb)
	defer patches.Reset()
	setup()
	code := m.Run()
	teardown()
	hwlog.RunLog.Infof("exit_code=%d\n", code)
}

func TestAll(t *testing.T) {
	createGroupAndRelation(1)
	convey.Convey("node manager function test", t, func() {

		convey.Convey("create node should success", testCreateNode)
		convey.Convey("get node detail should success", testGetNodeDetail)
		convey.Convey("modify node should success", testModifyNode)
		convey.Convey("get nod statistics should success", testGetNodeStatistics)
		convey.Convey("list node group should success", testListNodeGroup)
		convey.Convey("get node group detail should success", testGetNodeGroupDetail)
		convey.Convey("create node group should success", testCreateNodeGroup)
		convey.Convey("list node should success", testListNode)
	})
}

func testCreateNode() {
	req := util.CreateEdgeNodeReq{
		Description: "my-desc",
		NodeName:    "node-name",
		UniqueName:  "unique-name",
		NodeGroup:   "node-group",
	}
	reqBytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	resp := createNode(string(reqBytes))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testCreateNodeGroup() {
	req := util.CreateNodeGroupReq{
		Description:   "my-desc",
		NodeGroupName: "node-group-name",
	}
	reqBytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	resp := createGroup(string(reqBytes))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testGetNodeDetail() {
	testGetNodeDetailInternal("my-desc", "node-name", "unique-name", "my-group", 1)
}

func testGetNodeDetailInternal(description, nodeName, uniqueName, nodeGroup string, nodeId int64) {
	req := map[string][]string{"id": {strconv.Itoa(int(nodeId))}}
	reqBytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	resp := getNodeDetail(string(reqBytes))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	convey.So(resp.Data, convey.ShouldHaveSameTypeAs, util.GetNodeDetailResp{})
	node, _ := resp.Data.(util.GetNodeDetailResp)
	convey.So(node.Description, convey.ShouldEqual, description)
	convey.So(node.NodeName, convey.ShouldEqual, nodeName)
	convey.So(node.UniqueName, convey.ShouldEqual, uniqueName)
	convey.So(node.NodeGroup, convey.ShouldEqual, nodeGroup)
}

func testModifyNode() {
	req := util.ModifyNodeGroupReq{
		NodeId:      1,
		Description: "my-desc-new",
		NodeName:    "node-name-new",
	}
	reqBytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	resp := modifyNode(string(reqBytes))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	testGetNodeDetailInternal("my-desc-new", "node-name-new", "unique-name", "my-group", 1)
}

func testGetNodeStatistics() {
	resp := getNodeStatistics("")
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	convey.So(resp.Data, convey.ShouldHaveSameTypeAs, util.GetNodeStatisticsResp{})
	counts, _ := resp.Data.(util.GetNodeStatisticsResp)
	convey.So(counts, convey.ShouldContainKey, statusReady)
	convey.So(counts[statusReady], convey.ShouldEqual, 0)
	convey.So(counts, convey.ShouldContainKey, statusNotReady)
	convey.So(counts[statusNotReady], convey.ShouldEqual, 0)
	convey.So(counts, convey.ShouldContainKey, statusOffline)
	convey.So(counts[statusOffline], convey.ShouldEqual, 1)
	convey.So(counts, convey.ShouldContainKey, statusUnknown)
	convey.So(counts[statusUnknown], convey.ShouldEqual, 0)
}

func testListNodeGroup() {
	resp := listEdgeNodeGroup("")
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	convey.So(resp.Data, convey.ShouldHaveSameTypeAs, util.ListNodeGroupResp{})
	respData, _ := resp.Data.(util.ListNodeGroupResp)
	convey.So(len(respData), convey.ShouldEqual, 1)
	group := respData[0]
	convey.So(group.NodeCount, convey.ShouldEqual, 1)
}

func testListNode() {
	req := util.ListReq{
		PageSize: 1,
		PageNum:  1,
	}
	reqBytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	resp := listNode(string(reqBytes))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testGetNodeGroupDetail() {
	req := map[string][]string{"id": {"1"}}
	reqBytes, err := json.Marshal(req)
	convey.So(err, convey.ShouldBeNil)
	resp := getEdgeNodeGroupDetail(string(reqBytes))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
	convey.So(resp.Data, convey.ShouldHaveSameTypeAs, util.GetNodeGroupDetailResp{})
	respData, _ := resp.Data.(util.GetNodeGroupDetailResp)
	convey.So(len(respData.Nodes), convey.ShouldEqual, 1)
}

func createGroupAndRelation(nodeId int64) {
	if gormInstance == nil {
		hwlog.RunLog.Error("null pointer error")
		return
	}
	nodeGroup := NodeGroup{
		Description: "my-description",
		GroupName:   "my-group",
		Label:       "my-label",
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdateAt:    time.Now().Format(TimeFormat),
	}
	if err := gormInstance.Create(&nodeGroup).Error; err != nil {
		hwlog.RunLog.Errorf("create group failed, %v\n", err)
	}
	nodeRelation := NodeRelation{
		GroupID:   nodeGroup.ID,
		NodeID:    nodeId,
		CreatedAt: time.Now().Format(TimeFormat),
	}
	if err := gormInstance.Create(&nodeRelation).Error; err != nil {
		hwlog.RunLog.Errorf("create relation failed, %v\n", err)
	}
}
