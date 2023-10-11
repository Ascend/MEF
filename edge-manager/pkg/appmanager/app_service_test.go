// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
)

var (
	gormInstance *gorm.DB
	dbPath       = "./test.db"
)

const (
	notExitID      = 100
	exceedPageSize = 101
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
		hwlog.RunLog.Errorf("failed to init test db, %v", err)
	}
	if err = gormInstance.AutoMigrate(&AppInfo{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v", err)
	}
	if err = gormInstance.AutoMigrate(&AppInstance{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v", err)
	}
	if err = gormInstance.AutoMigrate(&AppTemplateDb{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v", err)
	}
	if err = gormInstance.AutoMigrate(&AppDaemonSet{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v", err)
	}
	if err = gormInstance.AutoMigrate(&ConfigmapInfo{}); err != nil {
		hwlog.RunLog.Errorf("setup table error, %v", err)
	}

	if _, err = kubeclient.NewClientK8s(); err != nil {
		hwlog.RunLog.Error("init k8s failed")
	}
}

func teardown() {
	if err := os.Remove(dbPath); err != nil && errors.Is(err, os.ErrExist) {
		fmt.Printf("cleanup [%s] failed, error: %v", dbPath, err)
	}
}

func mockGetDb() *gorm.DB {
	return gormInstance
}

func TestMain(m *testing.M) {
	patches := gomonkey.
		ApplyFunc(database.GetDb, mockGetDb)
	defer patches.Reset()
	setup()
	code := m.Run()
	teardown()
	hwlog.RunLog.Infof("exit_code=%d\n", code)
}

func TestCreateApp(t *testing.T) {
	convey.Convey("create app should success", t, testCreateApp)
	convey.Convey("create app should success", t, testCreateAppError)
}

func TestQuryApp(t *testing.T) {
	convey.Convey("query not exist app should failed", t, testQueryAppNotExist)
	convey.Convey("query app should success", t, testQueryApp)
	convey.Convey("query app error input", t, testQueryAppError)
}

func TestListApp(t *testing.T) {
	convey.Convey("list app info", t, testListAppInfo)
	convey.Convey("list app info error input", t, testListAppInfoError)
	convey.Convey("list app info invalid input", t, testListAppInfoInvalid)
}

func TestDeployApp(t *testing.T) {
	convey.Convey("deploy app info should success", t, testDeployApInfo)
	convey.Convey("deploy app info not exit", t, testDeployApInfoError)
	convey.Convey("deploy app info error input", t, testDeployInvalid)
}

func TestUnDeployApp(t *testing.T) {
	convey.Convey("undeploy app info should success", t, testUndeployApInfo)
	convey.Convey("undeploy app info not exit", t, testUndeployNotExit)
}

func TestUpdateApp(t *testing.T) {
	convey.Convey("update app info should success", t, testUpdateApp)
	convey.Convey("update app info should success", t, testUpdateAppDuplicate)
	convey.Convey("update app not exit", t, testUpdateNotExistApp)
	convey.Convey("update app not exit", t, testUpdateAppErrorInput)
}

func TestDeleteApp(t *testing.T) {
	convey.Convey("delete app info should success", t, testDeleteApp)
	convey.Convey("delete not exist app should failed", t, testDeleteNotExistApp)
	convey.Convey("delete app info should success", t, testDeleteAppError)
}

func TestListAppInstances(t *testing.T) {
	convey.Convey("list app instance should success", t, testListAppInstance)
	convey.Convey("list app instance error input", t, testListAppInstanceError)
	convey.Convey("list app instance invalid input", t, testListAppInstanceInvalid)
}

func TestParseDaemonsetToDB(t *testing.T) {
	convey.Convey("list app instance should success", t, testGetInstanceFromAppInstances)
}

func TestListAppInstancesByNode(t *testing.T) {
	convey.Convey("test ListAppInstancesByNode", t, testListAppInstancesByNode)
	convey.Convey("test ListAppInstancesByNode error input", t, testListAppInstancesByNodeError)
}

func TestListAppInstancesById(t *testing.T) {
	convey.Convey("test ListAppInstancesById", t, testListAppInstancesById)
	convey.Convey("test ListAppInstancesById error input", t, testListAppInstancesByIdError)
}

func testCreateApp() {
	reqData := `{
    "appName":"face-check",
    "description":"",
    "containers":[{
            "args":[],
            "command":[],
            "containerPort":[],
			"memRequest": 1024,
            "cpuRequest": 1,
            "env":[],
			"containerPort": [
				{
					"name": "test-port",
                    "proto": "TCP",
                    "containerPort": 1234,
                    "hostIP": "12.23.45.78",
                    "hostPort": 6666
				}
			],
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "memRequest": 1024,
            "name":"afafda",
            "userId":1024
	}]
}`
	resp := createApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testCreateAppError() {
	reqData := `{
    "appName":"face-check",
    "description":"",
    "containers":[{
			"memRequest": 1024,
            "cpuRequest": 100000,
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "memRequest": 1024,
            "name":"afafda",
            "userId":1024
	}]
}`
	resp := createApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testDeleteNotExistApp() {
	reqData := `{
				"appIDs": [100]
				}`
	resp := deleteApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteApp)
}

func testDeleteApp() {
	reqData := `{
				"appIDs": [1]
				}`
	resp := deleteApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDeleteAppError() {
	reqData := ""
	resp := deleteApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testUpdateNotExistApp() {
	reqData := `{
	"appID": 1000,
    "appName":"face-check",
    "description":"",
    "containers":[{
			"memRequest": 1024,
            "cpuRequest": 1,
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "memRequest":1024,
            "name":"afafda",
            "userId":1024
}]}`
	resp := updateApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorAppMrgRecodeNoFound)
}

func testUpdateApp() {
	reqData := `{
	"appID": 1,
    "appName":"face-check",
    "description":"",
    "containers":[{
            "args":[],
            "command":[],
            "containerPort":[],
			"memRequest": 1024,
            "cpuRequest": 1,
            "env":[],
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "memRequest":1024,
            "name":"afafda",
            "userId":1024
}]}`
	var p1 = gomonkey.ApplyPrivateMethod(AppRepositoryInstance(), "updateApp",
		func(*AppInfo) error {
			return nil
		})
	defer p1.Reset()
	var p2 = gomonkey.ApplyFunc(updateNodeGroupDaemonSet, func(appInfo *AppInfo, nodeGroups []types.NodeGroupInfo) error {
		return nil
	})
	defer p2.Reset()
	resp := updateApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUpdateAppDuplicate() {
	reqData := `{
	"appID": 1,
    "appName":"face-check",
    "description":"",
    "containers":[{
            "args":[],
            "command":[],
            "containerPort":[],
			"memRequest": 1024,
            "cpuRequest": 1,
            "env":[],
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "memRequest":1024,
            "name":"afafda",
            "userId":1024
}]}`
	var p1 = gomonkey.ApplyPrivateMethod(AppRepositoryInstance(), "updateApp",
		func(*AppInfo) error {
			return nil
		})
	defer p1.Reset()
	var p2 = gomonkey.ApplyFunc(updateNodeGroupDaemonSet, func(appInfo *AppInfo, nodeGroups []types.NodeGroupInfo) error {
		return nil
	})
	defer p2.Reset()
	resp := updateApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUpdateAppErrorInput() {
	reqData := `{
	"appID": 1,
    "appName":"face-check",
    "containers":[{
			"memRequest": 1024,
            "cpuRequest": 100000,
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "memRequest":1024,
            "name":"afafda",
            "userId":1024
}]}`
	resp := updateApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testQueryAppNotExist() {
	var reqData = uint64(notExitID)
	resp := queryApp(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testQueryApp() {
	var reqData = uint64(1)
	resp := queryApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testQueryAppError() {
	var reqData = ""
	resp := queryApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testListAppInfo() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
		Name:     "face-check",
	}
	resp := listAppInfo(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testListAppInfoError() {
	reqData := ""
	resp := listAppInfo(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testListAppInfoInvalid() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: exceedPageSize,
		Name:     "face-check",
	}
	resp := listAppInfo(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testDeployApInfo() {
	reqData := `{
    "appId": 1,
    "nodeGroupIds": [1,2]}`
	var c *kubeclient.Client
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "CreateDaemonSet",
		func(*kubeclient.Client, *v1.DaemonSet) (*v1.DaemonSet, error) {
			return &v1.DaemonSet{}, nil
		})
	var p2 = gomonkey.ApplyFunc(getNodeGroupInfos,
		func(nodeGroupIds []uint64) ([]types.NodeGroupInfo, error) {
			return []types.NodeGroupInfo{{NodeGroupID: 1, NodeGroupName: "group1"},
				{NodeGroupID: 2, NodeGroupName: "group2"}}, nil
		})
	var p3 = gomonkey.ApplyFunc(checkNodeGroupRes,
		func(nodeGroupId uint64, daemonSet *v1.DaemonSet, deployedNode map[uint64]int) error {
			return nil
		})
	defer p1.Reset()
	defer p2.Reset()
	defer p3.Reset()
	resp := deployApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDeployApInfoError() {
	reqData := `{
    "appId": 10000,
    "nodeGroupIds": [1,2]}`
	resp := deployApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorAppMrgRecodeNoFound)
}

func testDeployInvalid() {
	reqData := DeleteAppReq{}
	resp := deployApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testUndeployApInfo() {
	reqData := `{
    "appId": 1,
    "nodeGroupIds": [1,2]}`
	var c *kubeclient.Client
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "DeleteDaemonSet",
		func(*kubeclient.Client, string) error {
			return nil
		})
	defer p1.Reset()
	resp := unDeployApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUndeployNotExit() {
	reqData := `{
    "appId": 100,
    "nodeGroupIds": [1,2]}`
	var c *kubeclient.Client
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "DeleteDaemonSet",
		func(*kubeclient.Client, string) error {
			return nil
		})
	defer p1.Reset()
	resp := unDeployApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorUnDeployApp)
}

func testListAppInstance() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
	}
	resp := listAppInstances(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testListAppInstanceError() {
	reqData := ""
	resp := listAppInstances(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testListAppInstanceInvalid() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: exceedPageSize,
	}
	resp := listAppInstances(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testGetInstanceFromAppInstances() {
	patchFunc := gomonkey.ApplyFunc(getAppIdFromDaemonSet, func(_ *v1.DaemonSet) (uint64, error) {
		return 1, nil
	})
	patchFunc2 := gomonkey.ApplyFunc(common.SendSyncMessageByRestful, func(interface{}, *common.Router,
		time.Duration) common.RespMsg {
		data := types.NodeGroupInfo{NodeGroupID: 1, NodeGroupName: "name"}
		return common.RespMsg{Status: common.Success, Msg: "", Data: types.InnerGetNodeGroupInfosResp{
			NodeGroupInfos: []types.NodeGroupInfo{data}}}
	})
	defer patchFunc.Reset()
	defer patchFunc2.Reset()
	selector := map[string]string{fmt.Sprintf("%s%s", common.NodeGroupLabelPrefix, "1024"): ""}
	eventSet := v1.DaemonSet{
		Spec: v1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					NodeSelector: selector,
				}},
		},
	}
	_, res := parseDaemonSetToDB(&eventSet)
	convey.So(res, convey.ShouldBeNil)
}

func testListAppInstancesByNode() {
	input := uint64(1)
	res := listAppInstancesByNode(input)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testListAppInstancesByNodeError() {
	input := ""
	res := listAppInstancesByNode(input)
	convey.So(res.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testListAppInstancesById() {
	input := uint64(1)
	res := listAppInstancesById(input)
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testListAppInstancesByIdError() {
	input := ""
	res := listAppInstancesById(input)
	convey.So(res.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}
