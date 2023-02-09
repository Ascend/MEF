// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appmanager for
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	. "github.com/agiledragon/gomonkey/v2"
	. "github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"

	"edge-manager/pkg/database"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

var testErr = errors.New("test error")

func testCreateConfigmap() {
	input := `{
    "configmapName":"test01",
    "description":"",
    "configmapContent":[
        {
            "name":"name01",
            "value":"value01"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "CreateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	resp := createConfigmap(input)
	So(resp.Status, ShouldEqual, common.Success)
}

func testCreateConfigmapDuplicateName() {
	input := `{
    "configmapName":"test01",
    "description":"",
    "configmapContent":[
        {
            "name":"name001",
            "value":"value001"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "CreateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	resp := createConfigmap(input)
	So(resp.Status, ShouldEqual, "")
}

func testCreateConfigmapItemCountError() {
	input := `{
    "configmapName":"test03",
    "description":"",
    "configmapContent":[
        {
            "name":"name03",
            "value":"value03"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "CreateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	var itemCount = 65
	var p2 = ApplyFunc(database.GetItemCount, func(configmapInfo interface{}) (int, error) {
		return itemCount, nil
	})
	defer p2.Reset()

	resp := createConfigmap(input)
	So(resp.Status, ShouldEqual, "")
}

func testCreateConfigmapParamError() {
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "CreateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	configmapReqs := constructConfigmapReqs()
	for index := range configmapReqs {
		configmapReq := configmapReqs[index]

		configmapData, err := json.Marshal(configmapReq)
		if err != nil {
			hwlog.RunLog.Errorf("marshal configmap request failed, error: %v", err)
			return
		}
		resp := createConfigmap(string(configmapData))
		So(resp.Status, ShouldEqual, "")
	}
}

func constructConfigmapReqs() []ConfigmapReq {
	var content01 = make([]ConfigmapContent, 2)
	content01[0] = ConfigmapContent{
		Name:  "name01",
		Value: "value01",
	}
	content01[1] = ConfigmapContent{
		Name:  "name01",
		Value: "value02",
	}

	var content02 = make([]ConfigmapContent, 2)
	content02[0] = ConfigmapContent{
		Name:  " /name01",
		Value: "value01",
	}

	configmapReqs := []ConfigmapReq{
		{"./test01", "", nil},
		{
			"test01",
			"this is a description of test01,this is a description of test01,this is a description of test01," +
				"this is a description of test01,this is a description of test01,this is a description of test01," +
				"this is a description of test01,this is a description of test01...",
			nil,
		},
		{"test01", "", content01},
		{"test01", "", content02},
	}
	return configmapReqs
}

func testCreateConfigmapParamConvertError() {
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "CreateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	configmapInput := [1]int{1}
	resp := createConfigmap(configmapInput)
	So(resp.Status, ShouldEqual, "")
}

func testCreateConfigmapK8SError() {
	input := `{
    "configmapName":"test04",
    "description":"",
    "configmapContent":[
        {
            "name":"name04",
            "value":"value04"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "CreateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, testErr
		})
	defer p1.Reset()

	resp := createConfigmap(input)
	So(resp.Status, ShouldEqual, "")
}

func testUpdateConfigmap() {
	input := `{
    "configmapName":"test01",
    "description":"",
    "configmapContent":[
        {
            "name":"name001",
            "value":"value001"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "UpdateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	resp := updateConfigmap(input)
	So(resp.Status, ShouldEqual, common.Success)
}

func testUpdateConfigmapNotExist() {
	input := `{
    "configmapName":"test02",
    "description":"",
    "configmapContent":[
        {
            "name":"name02",
            "value":"value02"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "UpdateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	resp := updateConfigmap(input)
	So(resp.Status, ShouldEqual, "")
}

func testUpdateConfigmapParamError() {
	input := `{
    "configmapName":"./test01",
    "description":"",
    "configmapContent":[
        {
            "name":"name01",
            "value":"value01"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "UpdateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	resp := updateConfigmap(input)
	So(resp.Status, ShouldEqual, "")
}

func testUpdateConfigmapParamConvertError() {
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "UpdateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	updateInput := [1]int{1}
	resp := updateConfigmap(updateInput)
	So(resp.Status, ShouldEqual, "")
}

func testUpdateConfigmapK8SError() {
	input := `{
    "configmapName":"test01",
    "description":"",
    "configmapContent":[
        {
            "name":"name005",
            "value":"value005"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "UpdateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, testErr
		})
	defer p1.Reset()

	resp := updateConfigmap(input)
	So(resp.Status, ShouldEqual, "")
}

func testQueryConfigmap() {
	var reqData = int64(1)
	resp := queryConfigmap(reqData)
	So(resp.Status, ShouldEqual, common.Success)
}

func testQueryConfigmapNotExist() {
	var reqData = int64(100)
	resp := queryConfigmap(reqData)
	So(resp.Status, ShouldEqual, "")
}

func testQueryConfigmapParamConvertError() {
	reqData := [1]int{1}
	resp := queryConfigmap(reqData)
	So(resp.Status, ShouldEqual, "")
}

func testQueryConfigmapContentUnmarshalError() {
	content := 1
	contentData, err := json.Marshal(content)
	if err != nil {
		hwlog.RunLog.Errorf("marshal content failed, error: %v", err)
		return
	}

	var configmapInfo = &ConfigmapInfo{
		ConfigmapContent: string(contentData),
	}

	ConfigmapRepositoryInstance()
	var p1 = ApplyPrivateMethod(configmapRepository, "queryConfigmapByID",
		func(cri ConfigmapRepositoryImpl, configmapID int64) (*ConfigmapInfo, error) {
			return configmapInfo, nil
		})
	defer p1.Reset()

	var reqData = int64(1)
	resp := queryConfigmap(reqData)
	So(resp.Status, ShouldEqual, "")
}

func testListConfigmap() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 5,
		Name:     "test01",
	}
	resp := listConfigmap(reqData)
	So(resp.Status, ShouldEqual, common.Success)
}

func testListConfigmapParamConvertError() {
	reqData := [1]int{1}
	resp := listConfigmap(reqData)
	So(resp.Status, ShouldEqual, "")
}

func testListConfigmapNotExist() {
	var p1 = ApplyFunc(getListConfigmapReturnInfo, func(listReq types.ListReq) (*ListConfigmapReturnInfo, error) {
		return nil, fmt.Errorf("unmarshal configmap [%d] content failed", 1)
	})
	defer p1.Reset()

	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 5,
		Name:     "test01",
	}
	resp := listConfigmap(reqData)
	So(resp.Status, ShouldEqual, "")
}

func testDeleteConfigmap() {
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "DeleteConfigMap",
		func(*kubeclient.Client, string) error {
			return nil
		})
	defer p1.Reset()

	reqData := `{
				"configmapIDs": [1]
				}`
	resp := deleteConfigmap(reqData)
	So(resp.Status, ShouldEqual, common.Success)
}

func testDeleteConfigmapNotExist() {
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "DeleteConfigMap",
		func(*kubeclient.Client, string) error {
			return nil
		})
	defer p1.Reset()

	reqData := `{
				"configmapIDs": [100]
				}`
	resp := deleteConfigmap(reqData)
	So(resp.Status, ShouldEqual, "")
}

func testDeleteConfigmapParamConvertError() {
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "DeleteConfigMap",
		func(*kubeclient.Client, string) error {
			return nil
		})
	defer p1.Reset()

	reqData := [1]int{1}
	resp := deleteConfigmap(reqData)
	So(resp.Status, ShouldEqual, "")
}

func testDeleteConfigmapK8SError() {
	input := `{
    "configmapName":"test05",
    "description":"",
    "configmapContent":[
        {
            "name":"name05",
            "value":"value05"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = ApplyMethod(reflect.TypeOf(c), "CreateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, nil
		})
	defer p1.Reset()

	createResp := createConfigmap(input)
	if createResp.Status != common.Success {
		hwlog.RunLog.Error("create configmap failed")
		return
	}

	var p2 = ApplyMethod(reflect.TypeOf(c), "DeleteConfigMap",
		func(*kubeclient.Client, string) error {
			return testErr
		})
	defer p2.Reset()
	reqData := `{
				"configmapIDs": [1]
				}`
	resp := deleteConfigmap(reqData)
	So(resp.Status, ShouldEqual, "")
}
