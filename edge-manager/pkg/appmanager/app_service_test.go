// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	"k8s.io/api/apps/v1"

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

const (
	notExitID      = 100
	exceedPageSize = 101
	testId         = 100
)

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

func TestListAppInstancesByNode(t *testing.T) {
	convey.Convey("test ListAppInstancesByNode", t, testListAppInstancesByNode)
	convey.Convey("test ListAppInstancesByNode error input", t, testListAppInstancesByNodeError)
}

func TestListAppInstancesById(t *testing.T) {
	convey.Convey("test ListAppInstancesById", t, testListAppInstancesById)
	convey.Convey("test ListAppInstancesById error input", t, testListAppInstancesByIdError)
}

func TestGetAppInstanceRespFromAppInstances(t *testing.T) {
	convey.Convey("test getAppInstanceRespFromAppInstances", t, testGetAppInstanceRespFromAppInstances)
	convey.Convey("test getAppInstanceRespFromAppInstances error input", t, testGetAppInstanceRespFromAppInstancesError)
}

func getTestCreateAppReq(containers ...Container) CreateAppReq {
	req := CreateAppReq{
		AppName:    "face-check",
		Containers: containers,
	}
	return req
}

func getTestContainer() Container {
	uid, gid := int64(1024), int64(1024)
	memoryLimit, cpuLimit, npu := int64(1024), float64(1), int64(1)
	container := Container{
		Name:         "container1",
		UserID:       &uid,
		GroupID:      &gid,
		MemRequest:   1024,
		MemLimit:     &memoryLimit,
		CpuRequest:   1,
		CpuLimit:     &cpuLimit,
		Npu:          &npu,
		Image:        "euler_image",
		ImageVersion: "1.0",
		Ports: []ContainerPort{{
			Name:          "test-port",
			Proto:         "TCP",
			ContainerPort: 1234,
			HostIP:        "127.0.0.1",
			HostPort:      6666,
		}},
		HostPathVolumes: []HostPathVolume{{
			Name:      "v1",
			HostPath:  "/usr/local/sbin/npu-smi",
			MountPath: "/usr/local/sbin/npu-smi",
		}},
	}
	return container
}

func getTestJsonString(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func testCreateApp() {
	resp := createApp(&model.Message{Content: getTestJsonString(getTestCreateAppReq(getTestContainer()))})
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testCreateAppError() {
	resp := createApp(&model.Message{Content: []byte("error content")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)

	container := getTestContainer()
	reqData := getTestJsonString(getTestCreateAppReq(container, container))
	resp = createApp(&model.Message{Content: reqData})
	convey.So(resp.Msg, convey.ShouldEqual, "para check failed: check containers par failed: duplicated name")

	container.CpuRequest = 10000
	reqData = getTestJsonString(getTestCreateAppReq(container))
	resp = createApp(&model.Message{Content: reqData})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)

	patches := gomonkey.ApplyFuncSeq(GetTableCount, []gomonkey.OutputCell{
		{Values: gomonkey.Params{0, test.ErrTest}, Times: 1},
		{Values: gomonkey.Params{MaxApp, nil}, Times: 1},
	})
	defer patches.Reset()
	reqData = getTestJsonString(getTestCreateAppReq(getTestContainer()))
	resp = createApp(&model.Message{Content: reqData})
	convey.So(resp.Msg, convey.ShouldEqual, "get app table num failed")
	resp = createApp(&model.Message{Content: reqData})
	convey.So(resp.Msg, convey.ShouldEqual, "app number is enough, can not be created")
}

func testDeleteNotExistApp() {
	reqData := `{
				"appIDs": [100]
				}`
	resp := deleteApp(&model.Message{Content: []byte(reqData)})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteApp)
}

func testDeleteApp() {
	reqData := `{
				"appIDs": [1]
				}`
	resp := deleteApp(&model.Message{Content: []byte(reqData)})
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDeleteAppError() {
	resp := deleteApp(&model.Message{Content: []byte("")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)

	resp = deleteApp(&model.Message{Content: []byte(`{"appIDs": [0]}`)})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)

	patches := gomonkey.ApplyPrivateMethod(&AppRepositoryImpl{}, "getAppInfoById",
		func(a *AppRepositoryImpl, appId uint64) (*AppInfo, error) { return &AppInfo{}, nil })
	defer patches.Reset()
	resp = deleteApp(&model.Message{Content: []byte(`{"appIDs": [1]}`)})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteApp)

	patches.ApplyPrivateMethod(&AppRepositoryImpl{}, "deleteAppById",
		func(a *AppRepositoryImpl, appId uint64) (int64, error) {
			return 0, errors.New("app is referenced, can not be deleted")
		})
	resp = deleteApp(&model.Message{Content: []byte(`{"appIDs": [1]}`)})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteApp)
}

func testUpdateNotExistApp() {
	reqData := getTestJsonString(UpdateAppReq{
		AppID: 1000,
		CreateAppReq: CreateAppReq{
			AppName:    "face-check",
			Containers: []Container{getTestContainer()},
		},
	})
	resp := updateApp(&model.Message{Content: reqData})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorAppMrgRecodeNoFound)
}

func testUpdateApp() {
	container := getTestContainer()
	container.ImageVersion = "2.0"
	reqData := getTestJsonString(UpdateAppReq{
		AppID: 1,
		CreateAppReq: CreateAppReq{
			AppName:    "face-check",
			Containers: []Container{container},
		},
	})
	var p1 = gomonkey.ApplyPrivateMethod(AppRepositoryInstance(), "updateApp",
		func(*AppInfo) error {
			return nil
		})
	defer p1.Reset()
	var p2 = gomonkey.ApplyFunc(updateNodeGroupDaemonSet, func(appInfo *AppInfo, nodeGroups []types.NodeGroupInfo) error {
		return nil
	})
	defer p2.Reset()
	resp := updateApp(&model.Message{Content: reqData})
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUpdateAppDuplicate() {
	container := getTestContainer()
	container.ImageVersion = "2.0"
	reqData := getTestJsonString(UpdateAppReq{
		AppID: 1,
		CreateAppReq: CreateAppReq{
			AppName:    "face-check",
			Containers: []Container{container},
		},
	})
	var p1 = gomonkey.ApplyPrivateMethod(AppRepositoryInstance(), "updateApp",
		func(*AppInfo) error {
			return nil
		})
	defer p1.Reset()
	var p2 = gomonkey.ApplyFunc(updateNodeGroupDaemonSet, func(appInfo *AppInfo, nodeGroups []types.NodeGroupInfo) error {
		return nil
	})
	defer p2.Reset()
	resp := updateApp(&model.Message{Content: reqData})
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUpdateAppErrorInput() {
	container := getTestContainer()
	container.CpuRequest = 10000
	container.ImageVersion = "2.0"
	reqData := getTestJsonString(UpdateAppReq{
		AppID: 1,
		CreateAppReq: CreateAppReq{
			AppName:    "face-check",
			Containers: []Container{container},
		},
	})
	resp := updateApp(&model.Message{Content: reqData})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testQueryAppNotExist() {
	resp := queryApp(newMsgWithContentForUT(uint64(notExitID)))
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testQueryApp() {
	var p = gomonkey.ApplyFuncReturn(getNodeGroupInfos,
		[]types.NodeGroupInfo{{NodeGroupID: 1, NodeGroupName: "group1"}}, nil)
	defer p.Reset()
	resp := queryApp(newMsgWithContentForUT(uint64(1)))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testQueryAppError() {
	resp := queryApp(&model.Message{Content: []byte("")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testListAppInfo() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
		Name:     "face-check",
	}
	var p = gomonkey.ApplyFuncReturn(getNodeGroupInfos,
		[]types.NodeGroupInfo{{NodeGroupID: 1, NodeGroupName: "group1"}}, nil)
	defer p.Reset()
	resp := listAppInfo(newMsgWithContentForUT(reqData))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testListAppInfoError() {
	resp := listAppInfo(&model.Message{Content: []byte("")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)

	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
		Name:     "face-check",
	}
	patches := gomonkey.ApplyPrivateMethod(AppRepositoryInstance(), "listAppsInfo",
		func(a *AppRepositoryImpl, page, pageSize uint64, name string) ([]AppInfo, error) {
			info := AppInfo{
				ID:         1,
				AppName:    "fake-check",
				Containers: "error-container",
			}
			return []AppInfo{info}, nil
		})
	defer patches.Reset()
	resp = listAppInfo(newMsgWithContentForUT(reqData))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListApp)
}

func testListAppInfoInvalid() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: exceedPageSize,
		Name:     "face-check",
	}
	resp := listAppInfo(newMsgWithContentForUT(reqData))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testDeployApInfo() {
	reqData := `{
    "appId": 1,
    "nodeGroupIds": [1,2]}`
	var p = gomonkey.ApplyPrivateMethod(&AppRepositoryImpl{}, "addDaemonSet",
		func(ds *v1.DaemonSet, nodeGroupId, appId uint64) error { return nil }).
		ApplyFuncReturn(getNodeGroupInfos, []types.NodeGroupInfo{{NodeGroupID: 1, NodeGroupName: "group1"}}, nil).
		ApplyFuncReturn(checkNodeGroupResource, nil).
		ApplyFuncReturn(updateAllocatedNodeRes, nil).
		ApplyFuncReturn(preCheckForDeployApp, nil)
	defer p.Reset()

	resp := deployApp(&model.Message{Content: []byte(reqData)})
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDeployApInfoError() {
	reqData := `{
    "appId": 10000,
    "nodeGroupIds": [1,2]}`
	resp := deployApp(&model.Message{Content: []byte(reqData)})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorAppMrgRecodeNoFound)

	resp = deployApp(&model.Message{Content: []byte("error data")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)

	var patches = gomonkey.ApplyPrivateMethod(&AppRepositoryImpl{}, "getAppInfoById",
		func(a *AppRepositoryImpl, appId uint64) (*AppInfo, error) { return nil, test.ErrTest })
	defer patches.Reset()
	resp = deployApp(&model.Message{Content: []byte(`{"appId": 1,"nodeGroupIds": [1,2]}`)})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeployApp)
}

func testDeployInvalid() {
	resp := deployApp(newMsgWithContentForUT(DeleteAppReq{}))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testUndeployApInfo() {
	reqData := `{
    "appId": 1,
    "nodeGroupIds": [1,2]}`
	var p1 = gomonkey.ApplyPrivateMethod(kubeclient.GetKubeClient(), "DeleteDaemonSet",
		func(string) error { return nil }).
		ApplyPrivateMethod(kubeclient.GetKubeClient(), "GetDaemonSet",
			func(string) (*v1.DaemonSet, error) { return &v1.DaemonSet{}, nil }).
		ApplyFuncReturn(updateAllocatedNodeRes, nil).
		ApplyFuncReturn(getNodeGroupInfos, []types.NodeGroupInfo{
			{NodeGroupName: "test"},
		}, nil)
	defer p1.Reset()
	resp := unDeployApp(&model.Message{Content: []byte(reqData)})
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUndeployNotExit() {
	reqData := `{
    "appId": 100,
    "nodeGroupIds": [1,2]}`
	var p1 = gomonkey.ApplyPrivateMethod(kubeclient.GetKubeClient(), "DeleteDaemonSet",
		func(string) error { return nil }).
		ApplyPrivateMethod(kubeclient.GetKubeClient(), "GetDaemonSet",
			func(string) (*v1.DaemonSet, error) { return &v1.DaemonSet{}, nil }).
		ApplyFuncReturn(updateAllocatedNodeRes, nil)
	defer p1.Reset()
	resp := unDeployApp(&model.Message{Content: []byte(reqData)})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorUnDeployApp)
}

func testListAppInstance() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
	}
	resp := listAppInstances(newMsgWithContentForUT(reqData))
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testListAppInstanceError() {
	resp := listAppInstances(&model.Message{Content: []byte("")})
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testListAppInstanceInvalid() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: exceedPageSize,
	}
	resp := listAppInstances(newMsgWithContentForUT(reqData))
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testListAppInstancesByNode() {
	res := listAppInstancesByNode(newMsgWithContentForUT(uint64(1)))
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testListAppInstancesByNodeError() {
	res := listAppInstancesByNode(&model.Message{Content: []byte("")})
	convey.So(res.Status, convey.ShouldEqual, common.ErrorParamConvert)

	res = listAppInstancesByNode(newMsgWithContentForUT(uint64(math.MaxUint32 + 1)))
	convey.So(res.Status, convey.ShouldEqual, common.ErrorParamInvalid)

	patches := gomonkey.ApplyFuncReturn(getAppInstanceRespFromAppInstances, nil, test.ErrTest)
	defer patches.Reset()
	res = listAppInstancesByNode(newMsgWithContentForUT(uint64(1)))
	convey.So(res.Status, convey.ShouldEqual, common.ErrorListAppInstancesByNode)
}

func testListAppInstancesById() {
	res := listAppInstancesById(newMsgWithContentForUT(uint64(1)))
	convey.So(res.Status, convey.ShouldEqual, common.Success)
}

func testListAppInstancesByIdError() {
	res := listAppInstancesById(&model.Message{Content: []byte("")})
	convey.So(res.Status, convey.ShouldEqual, common.ErrorParamConvert)

	res = listAppInstancesById(newMsgWithContentForUT(uint64(math.MaxUint32 + 1)))
	convey.So(res.Status, convey.ShouldEqual, common.ErrorParamInvalid)

	patches := gomonkey.ApplyFuncReturn(getAppInstanceRespFromAppInstances, nil, test.ErrTest)
	defer patches.Reset()
	res = listAppInstancesById(newMsgWithContentForUT(uint64(1)))
	convey.So(res.Status, convey.ShouldEqual, common.ErrorListAppInstancesByID)
}

func testGetAppInstanceRespFromAppInstances() {
	instance := AppInstance{
		ContainerInfo: `[{"name":"testContainer","image":"testImage:v1","status":"running","restartCount":0}]`,
	}
	patches := gomonkey.ApplyFuncReturn(getNodeStatus, "", nil).
		ApplyFuncReturn(getNodeGroupInfos, []types.NodeGroupInfo{{}}, nil)
	defer patches.Reset()

	appInstanceResp, err := getAppInstanceRespFromAppInstances([]AppInstance{instance})
	convey.So(len(appInstanceResp), convey.ShouldEqual, 1)
	convey.So(err, convey.ShouldBeNil)
}

func testGetAppInstanceRespFromAppInstancesError() {
	patches := gomonkey.ApplyFuncSeq(getNodeStatus, []gomonkey.OutputCell{
		{Values: gomonkey.Params{"", test.ErrTest}, Times: 1},
		{Values: gomonkey.Params{nodeStatusReady, nil}, Times: 2}}).
		ApplyFuncSeq(getNodeGroupInfos, []gomonkey.OutputCell{
			{Values: gomonkey.Params{nil, test.ErrTest}, Times: 1}})
	defer patches.Reset()

	appInstanceResp, err := getAppInstanceRespFromAppInstances([]AppInstance{{ContainerInfo: "error data"}})
	convey.So(len(appInstanceResp), convey.ShouldEqual, 0)
	convey.So(err, convey.ShouldBeNil)
	appInstanceResp, err = getAppInstanceRespFromAppInstances([]AppInstance{{}})
	convey.So(len(appInstanceResp), convey.ShouldEqual, 0)
	convey.So(err, convey.ShouldBeNil)
	appInstanceResp, err = getAppInstanceRespFromAppInstances([]AppInstance{{}})
	convey.So(len(appInstanceResp), convey.ShouldEqual, 0)
	convey.So(err, convey.ShouldBeNil)
}

func TestGetAppInstanceCountByNodeGroup(t *testing.T) {
	convey.Convey("Test Get AppInstance CountByNodeGroup", t, func() {
		ad := &AppDaemonSet{
			ID:            testId,
			AppID:         testId,
			NodeGroupID:   testId,
			NodeGroupName: "NodeGroupName",
		}
		err := test.MockGetDb().Model(&AppDaemonSet{}).Create(ad).Error
		convey.So(err, convey.ShouldBeNil)
		resp := getAppInstanceCountByNodeGroup(newMsgWithContentForUT([]uint64{testId}))
		convey.So(resp.Status, convey.ShouldEqual, common.Success)
		data, ok := resp.Data.(map[uint64]int64)
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(data[testId], convey.ShouldEqual, 1)
		err = test.MockGetDb().Model(&AppDaemonSet{}).Where(&AppDaemonSet{
			ID: testId,
		}).Delete(&AppDaemonSet{}).Error
		convey.So(err, convey.ShouldBeNil)
	})
}
