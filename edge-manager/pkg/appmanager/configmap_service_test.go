// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appmanager for configmap service
package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/api/core/v1"

	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

var testErr = errors.New("test error")

func setupGoMonkeyPatches() *gomonkey.Patches {
	c := &kubeclient.Client{}
	return gomonkey.ApplyFuncReturn(common.GetItemCount, 1, nil).
		ApplyMethodReturn(c, "CreateConfigMap", &v1.ConfigMap{}, nil).
		ApplyMethodReturn(c, "UpdateConfigMap", &v1.ConfigMap{}, nil).
		ApplyMethodReturn(c, "DeleteConfigMap", nil)
}

func TestConfigmap(t *testing.T) {
	p := setupGoMonkeyPatches()
	defer p.Reset()

	convey.Convey("test creat configmap", t, func() {
		convey.Convey("create configmap should success", testCreateConfigmap)
		convey.Convey("create configmap should failed", testCreateConfigmapDuplicateName)
		convey.Convey("create configmap should failed, check item count in db error", testCreateConfigmapItemCountError)
		convey.Convey("create configmap should failed, check param error", testCreateConfigmapParamError)
		convey.Convey("create configmap should failed, param convert error", testCreateConfigmapParamConvertError)
		convey.Convey("create configmap should failed, create by k8s error", testCreateConfigmapK8SError)
	})
	convey.Convey("test update configmap", t, func() {
		convey.Convey("update configmap should success", testUpdateConfigmap)
		convey.Convey("update configmap should failed, name does not exist", testUpdateConfigmapNotExist)
		convey.Convey("update configmap should failed, check param error", testUpdateConfigmapParamError)
		convey.Convey("update configmap should failed, param convert error", testUpdateConfigmapParamConvertError)
		convey.Convey("update configmap should failed, update by k8s error", testUpdateConfigmapK8SError)
	})
	convey.Convey("test query configmap", t, func() {
		convey.Convey("query configmap should success", testQueryConfigmap)
		convey.Convey("query configmap should failed, id does not exist", testQueryConfigmapNotExist)
		convey.Convey("query configmap should failed, param convert error", testQueryConfigmapParamConvertError)
		convey.Convey("query configmap should failed, content unmarshal error", testQueryConfigmapContentUnmarshalError)
	})
	convey.Convey("test list configmap", t, func() {
		convey.Convey("list configmap should success", testListConfigmap)
		convey.Convey("list configmap should failed, param convert error", testListConfigmapParamConvertError)
		convey.Convey("list configmap should failed, name does not exist", testListConfigmapNotExist)
	})
	convey.Convey("test delete configmap", t, func() {
		convey.Convey("delete configmap should success", testDeleteConfigmap)
		convey.Convey("delete configmap should failed， id does not exist", testDeleteConfigmapNotExist)
		convey.Convey("delete configmap should failed, param convert error", testDeleteConfigmapParamConvertError)
		convey.Convey("delete configmap should failed, delete by k8s error", testDeleteConfigmapK8SError)
	})
}

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
	resp := createConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
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
	resp := createConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorAppMrgDuplicate)
}

func testCreateConfigmapItemCountError() {
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
	const itemCount = 65
	var p2 = gomonkey.ApplyFunc(common.GetItemCount, func(configmapInfo interface{}) (int, error) {
		return itemCount, nil
	})
	defer p2.Reset()

	resp := createConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testCreateConfigmapParamError() {
	configmapReqs := constructConfigmapReqs()
	for index := range configmapReqs {
		configmapReq := configmapReqs[index]

		configmapData, err := json.Marshal(configmapReq)
		if err != nil {
			hwlog.RunLog.Errorf("marshal configmap request failed, error: %v", err)
			return
		}
		resp := createConfigmap(string(configmapData))
		convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
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

	configmapInput := [1]int{1}
	resp := createConfigmap(configmapInput)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testCreateConfigmapK8SError() {
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
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "CreateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, testErr
		})
	defer p1.Reset()

	resp := createConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorCreateCm)
}

func testUpdateConfigmap() {
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
	resp := createConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)

	input2 := `{
    "configmapName":"test04",
    "description":"",
    "configmapContent":[
        {
            "name":"name004",
            "value":"value004"
        }
    ]
}`
	resp = updateConfigmap(input2)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testUpdateConfigmapNotExist() {
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
	resp := updateConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorAppMrgRecodeNoFound)
}

func testUpdateConfigmapParamError() {
	input := `{
    "configmapName":"./test05",
    "description":"",
    "configmapContent":[
        {
            "name":"name05",
            "value":"value05"
        }
    ]
}`
	resp := updateConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamInvalid)
}

func testUpdateConfigmapParamConvertError() {
	updateInput := [1]int{1}
	resp := updateConfigmap(updateInput)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testUpdateConfigmapK8SError() {
	input := `{
    "configmapName":"test06",
    "description":"",
    "configmapContent":[
        {
            "name":"name006",
            "value":"value006"
        }
    ]
}`
	var c *kubeclient.Client
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "UpdateConfigMap",
		func(*kubeclient.Client, *v1.ConfigMap) (*v1.ConfigMap, error) {
			return &v1.ConfigMap{}, testErr
		})
	defer p1.Reset()

	resp := updateConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorAppMrgRecodeNoFound)
}

func testQueryConfigmap() {
	input := `{
    "configmapName":"test07",
    "description":"",
    "configmapContent":[
        {
            "name":"name07",
            "value":"value07"
        }
    ]
}`
	resp := createConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)

	var reqData = uint64(1)
	resp = queryConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testQueryConfigmapNotExist() {
	var reqData = uint64(notExitID)
	resp := queryConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorAppMrgRecodeNoFound)
}

func testQueryConfigmapParamConvertError() {
	reqData := [1]int{1}
	resp := queryConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
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

	CmRepositoryInstance()
	var p1 = gomonkey.ApplyPrivateMethod(configmapRepository, "queryCmByID",
		func(cri CmRepositoryImpl, configmapID uint64) (*ConfigmapInfo, error) {
			return configmapInfo, nil
		})
	defer p1.Reset()

	var reqData = uint64(1)
	resp := queryConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorUnmarshalCm)
}

func testListConfigmap() {
	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
		Name:     "test01",
	}
	resp := listConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testListConfigmapParamConvertError() {
	reqData := [1]int{1}
	resp := listConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorTypeAssert)
}

func testListConfigmapNotExist() {
	var p1 = gomonkey.ApplyFunc(getListConfigmapResp,
		func(listReq types.ListReq) (*ListConfigmapResp, error) {
			return nil, fmt.Errorf("unmarshal configmap [%d] content failed", 1)
		})
	defer p1.Reset()

	var reqData = types.ListReq{
		PageNum:  1,
		PageSize: 1,
		Name:     "test01",
	}
	resp := listConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorListCm)
}

func testDeleteConfigmap() {
	input := `{
    "configmapName":"test08",
    "description":"",
    "configmapContent":[
        {
            "name":"name08",
            "value":"value08"
        }
    ]
}`
	resp := createConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)

	reqData := `{
				"configmapIDs": [1]
				}`
	resp = deleteConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)
}

func testDeleteConfigmapNotExist() {
	reqData := `{
				"configmapIDs": [100]
				}`
	resp := deleteConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteCm)
}

func testDeleteConfigmapParamConvertError() {
	reqData := [1]int{1}
	resp := deleteConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorParamConvert)
}

func testDeleteConfigmapK8SError() {
	input := `{
    "configmapName":"test09",
    "description":"",
    "configmapContent":[
        {
            "name":"name09",
            "value":"value09"
        }
    ]
}`
	resp := createConfigmap(input)
	convey.So(resp.Status, convey.ShouldEqual, common.Success)

	reqData := `{
				"configmapIDs": [1]
				}`
	resp = deleteConfigmap(reqData)
	convey.So(resp.Status, convey.ShouldEqual, common.ErrorDeleteCm)
}
