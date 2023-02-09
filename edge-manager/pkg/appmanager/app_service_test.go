// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appmanager to init app manager service
package appmanager

import (
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/apps/v1"

	"edge-manager/pkg/database"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
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

	if _, err := kubeclient.NewClientK8s(""); err != nil {
		hwlog.RunLog.Error("init k8s failed")
	}
}

func teardown() {

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

func TestAll(t *testing.T) {
	convey.Convey("app manager function test", t, func() {
		convey.Convey("test app operate", func() {
			convey.Convey("test creat app", func() {
				convey.Convey("create app should success", testCreateApp)
			})

			convey.Convey("test query app", func() {
				convey.Convey("query not exist app should failed", testQueryAppNotExist)
				convey.Convey("query app should success", testQueryApp)
			})

			convey.Convey("list app info ", func() {
				convey.Convey("list app info should success", testListAppInfo)
			})

			convey.Convey("deploy app info ", func() {
				convey.Convey("deploy app info should success", testDeployApInfo)
			})

			convey.Convey("undeploy app info ", func() {
				convey.Convey("undeploy app info should success", testUndeployApInfo)
			})

			convey.Convey("test update app", func() {
				convey.Convey("update app should success", testUpdateApp)
			})

			convey.Convey("test delete app", func() {
				convey.Convey("delete not exist app should failed", testDeleteNotExistApp)
				convey.Convey("delete app should success", testDeleteApp)
			})
		})

		convey.Convey("test template operate", func() {
			convey.Convey("test creat app template", func() {
				convey.Convey("create app template should success", testCreateTemplate)
			})

			convey.Convey("test update app template", func() {
				convey.Convey("update app template should success", testUpdateTemplate)
			})

			convey.Convey("test get app template", func() {
				convey.Convey("get app template should success", testGetTemplate)
			})

			convey.Convey("test get app templates", func() {
				convey.Convey("get app templates should success", testGetTemplates)
			})

			convey.Convey("test delete app templates", func() {
				convey.Convey("delete app templates should success", testDeleteTemplate)
			})
		})

	})
}

func TestConfigmap(t *testing.T) {
	convey.Convey("test configmap operate", t, func() {
		convey.Convey("test creat configmap", func() {
			convey.Convey("create configmap should success", testCreateConfigmap)
			convey.Convey("create configmap should failed", testCreateConfigmapDuplicateName)
			convey.Convey("create configmap should failed, check item count in db error", testCreateConfigmapItemCountError)
			convey.Convey("create configmap should failed, check param error", testCreateConfigmapParamError)
			convey.Convey("create configmap should failed, param convert error", testCreateConfigmapParamConvertError)
			convey.Convey("create configmap should failed, create by k8s error", testCreateConfigmapK8SError)
		})

		convey.Convey("test update configmap", func() {
			convey.Convey("update configmap should success", testUpdateConfigmap)
			convey.Convey("update configmap should failed, name is not exist", testUpdateConfigmapNotExist)
			convey.Convey("update configmap should failed, check param error", testUpdateConfigmapParamError)
			convey.Convey("update configmap should failed, param convert error", testUpdateConfigmapParamConvertError)
			convey.Convey("update configmap should failed, update by k8s error", testUpdateConfigmapK8SError)
		})

		convey.Convey("test query configmap", func() {
			convey.Convey("query configmap should success", testQueryConfigmap)
			convey.Convey("query configmap should failed, id is not exist", testQueryConfigmapNotExist)
			convey.Convey("query configmap should failed, param convert error", testQueryConfigmapParamConvertError)
			convey.Convey("query configmap should failed, content unmarshal error", testQueryConfigmapContentUnmarshalError)
		})

		convey.Convey("test list configmap", func() {
			convey.Convey("list configmap should success", testListConfigmap)
			convey.Convey("list configmap should failed, param convert error", testListConfigmapParamConvertError)
			convey.Convey("list configmap should failed, name is not exist", testListConfigmapNotExist)
		})

		convey.Convey("test delete configmap", func() {
			convey.Convey("delete configmap should success", testDeleteConfigmap)
			convey.Convey("delete configmap should failed， id is not exist", testDeleteConfigmapNotExist)
			convey.Convey("delete configmap should failed, param convert error", testDeleteConfigmapParamConvertError)
			convey.Convey("delete configmap should failed, delete by k8s error", testDeleteConfigmapK8SError)
		})
	})
}

func testCreateApp() {
	reqData := `{
    "appName":"face-check",
    "description":"",
    "containers":[
        {
            "args":[
            ],
            "command":[
            ],
            "containerPort":[
            ],
			"memRequest": 1024,
            "cpuRequest": 1,
            "env":[
            ],
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "memRequest": 1024,
            "name":"afafda",
            "userId":1024
        }
    ]
}`
	resp := createApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
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

func testUpdateNotExistApp() {
	reqData := `{
				"appIDs": [100]
				}`
	resp := updateApp(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testUpdateApp() {
	reqData := `{
	"appID": 1,
    "appName":"face-check",
    "description":"",
    "containers":[
        {
            "args":[
            ],
            "command":[
            ],
            "containerPort":[
            ],
			"memRequest": 1024,
            "cpuRequest": 1,
            "env":[
            ],
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "memRequest":1024,
            "name":"afafda",
            "userId":1024
        }
    ]
}`
	var p1 = gomonkey.ApplyPrivateMethod(AppRepositoryInstance(), "queryNodeGroup",
		func(uint64) ([]types.NodeGroupInfo, error) {
			return []types.NodeGroupInfo{}, nil
		})
	defer p1.Reset()
	var p2 = gomonkey.ApplyFunc(updateNodeGroupDaemonSet, func(appInfo *AppInfo, nodeGroups []types.NodeGroupInfo) error {
		return nil
	})
	defer p2.Reset()
	resp := updateApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testQueryAppNotExist() {
	var reqData = uint64(100)
	resp := queryApp(reqData)
	convey.So(resp.Status, convey.ShouldNotEqual, common.Success)
}

func testQueryApp() {
	var reqData = uint64(1)
	resp := queryApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
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

func testDeployApInfo() {
	reqData := `{
    "appId": 1,
    "nodeGroupIds": [1,2]
}`
	var c *kubeclient.Client

	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "CreateDaemonSet",
		func(*kubeclient.Client, *v1.DaemonSet) (*v1.DaemonSet, error) {
			return &v1.DaemonSet{}, nil
		})
	var p2 = gomonkey.ApplyFunc(getNodeGroupInfos,
		func(nodeGroupIds []uint64) ([]types.NodeGroupInfo, error) {
			return []types.NodeGroupInfo{{1, "group1"},
				{2, "group2"}}, nil
		})
	defer p1.Reset()
	defer p2.Reset()
	resp := deployApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUndeployApInfo() {
	reqData := `{
    "appId": 1,
    "nodeGroupIds": [1,2]
}`
	var c *kubeclient.Client

	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "DeleteDaemonSet",
		func(*kubeclient.Client, string) error {
			return nil
		})
	defer p1.Reset()
	resp := unDeployApp(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testCreateTemplate() {
	reqData := `{
    "name":"template1",
    "description":"",
    "containers":[
        {
            "args":[
            ],
            "command":[
            ],
            "containerPort":[
            ],
			"memRequest": 1024,
            "cpuRequest": 1,
            "env":[
            ],
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "name":"afafda",
            "userId":1024
        }
    ]
}`
	resp := createTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUpdateTemplate() {
	reqData := `{
	"id":1,
    "name":"template1",
    "description":"",
    "containers":[
        {
            "args":[
            ],
            "command":[
            ],
            "containerPort":[
            ],
  			"memRequest": 1024,
            "cpuRequest": 1,
            "env":[
            ],
            "groupId":1024,
            "image":"euler_image",
            "imageVersion":"2.0",
            "name":"afafda",
            "userId":1024
        }
    ]
}`
	resp := updateTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testGetTemplate() {
	reqData := uint64(1)
	resp := getTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testGetTemplates() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
		Name:     "template1",
	}

	resp := getTemplates(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDeleteTemplate() {
	var reqData = `{
		"ids": [1]
 	}`

	resp := deleteTemplate(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}
