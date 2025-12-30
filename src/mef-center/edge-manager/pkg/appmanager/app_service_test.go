// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appmanager to init app manager service
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/apps/v1"

	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-manager/pkg/config"
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

func TestQueryApp(t *testing.T) {
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

// TestUpdateNodeGroupDaemonSet tests success and fails cases of the updateNodeGroupDaemonSet function
func TestUpdateNodeGroupDaemonSet(t *testing.T) {
	convey.Convey("Given updateNodeGroupDaemonSet function", t, func() {
		appInfo := &AppInfo{}
		nodeGroups := []types.NodeGroupInfo{
			{NodeGroupID: 1},
			{NodeGroupID: 2},
		}

		convey.Convey("When initDaemonSet fails", func() {
			patches := gomonkey.ApplyFunc(initDaemonSet, func(appInfo *AppInfo, nodeGroupID uint64) (*v1.DaemonSet, error) {
				return nil, test.ErrTest
			})
			defer patches.Reset()
			err := updateNodeGroupDaemonSet(appInfo, nodeGroups)

			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("init daemon set failded: %s", test.ErrTest.Error()))
		})

		convey.Convey("When UpdateDaemonSet fails", func() {
			patches := gomonkey.ApplyFunc(initDaemonSet, func(appInfo *AppInfo, nodeGroupID uint64) (*v1.DaemonSet, error) {
				return nil, nil
			}).ApplyMethod(reflect.TypeOf(&kubeclient.Client{}), "UpdateDaemonSet",
				func(k *kubeclient.Client, daemonSet *v1.DaemonSet) (*v1.DaemonSet, error) {
					return nil, test.ErrTest
				})
			defer patches.Reset()
			err := updateNodeGroupDaemonSet(appInfo, nodeGroups)

			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("update daemon set failded: %s", test.ErrTest.Error()))
		})

		convey.Convey("When updateNodeGroupDaemonSet success", func() {
			patches := gomonkey.ApplyFuncReturn(initDaemonSet, nil, nil).
				ApplyMethodReturn(&kubeclient.Client{}, "UpdateDaemonSet", nil, nil)
			defer patches.Reset()
			err := updateNodeGroupDaemonSet(appInfo, nodeGroups)

			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestPreCheckForDeployApp test preCheckForDeployApp success and failed cases
func TestPreCheckForDeployApp(t *testing.T) {
	convey.Convey("Given preCheckForDeployApp function success", t, testPreCheckForDeployAppSuccess)
	convey.Convey("Given preCheckForDeployApp function failed", t, testPreCheckForDeployAppFailed)
}

func testPreCheckForDeployAppSuccess() {
	convey.Convey("When all checks pass", func() {
		var maxDsNumber int64 = 1
		patches := gomonkey.ApplyFunc(AppRepositoryInstance().getAppDaemonSet, func(uint64, uint64) (*AppDaemonSet, error) {
			return nil, errors.New("not found")
		})
		patches.ApplyFunc(getNodeGroupInfos, func([]uint64) ([]types.NodeGroupInfo, error) {
			return nil, nil
		})
		patches.ApplyFunc(AppRepositoryInstance().countDeployedAppByGroupID, func(uint64) (int64, error) {
			return 0, nil
		})
		patches.ApplyGlobalVar(&config.PodConfig.MaxDsNumberPerNodeGroup, maxDsNumber)

		defer patches.Reset()
		err := preCheckForDeployApp(1, 1)

		convey.So(err, convey.ShouldBeNil)
	})
}

func testPreCheckForDeployAppFailed() {
	convey.Convey("When group id does not exist", func() {
		patches := gomonkey.ApplyFuncReturn(AppRepositoryInstance().getAppDaemonSet, nil, errors.New("not found")).
			ApplyFuncReturn(getNodeGroupInfos, nil, errors.New("group id no exist"))

		defer patches.Reset()
		err := preCheckForDeployApp(1, 1)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "group id no exist")
	})

	convey.Convey("When node group is out of max app limit", func() {
		patches := gomonkey.ApplyFuncReturn(AppRepositoryInstance().getAppDaemonSet, nil, errors.New("not found")).
			ApplyFuncReturn(getNodeGroupInfos, nil, nil).
			ApplyFuncReturn(AppRepositoryInstance().countDeployedAppByGroupID, config.PodConfig.MaxDsNumberPerNodeGroup, nil)

		defer patches.Reset()
		err := preCheckForDeployApp(1, 1)

		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldEqual, "node group out of max app limit")
	})
}
