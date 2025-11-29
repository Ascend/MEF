// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package mefmsgchecker
package msgchecker

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/msglistchecker"
)

func TestConfigMap(t *testing.T) {
	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb).
		ApplyFunc(configpara.GetPodConfig, MockPodConfig).
		ApplyFuncReturn(configpara.GetNetType, constants.FDWithOM, nil).
		ApplyPrivateMethod(&MsgValidator{}, "checkSystemResources", func() error { return nil })

	defer patches.Reset()

	convey.Convey("test fd config map para", t, func() {
		convey.Convey("test config map failed", testConfigMapFailed)
	})

}

var configMapData = `
{
    "data":{
        "test":"1234"
    },
    "kind":"Configmap",
    "metadata":{
        "creationTimestamp":"2023-12-25T03:10:56Z",
        "name":"cfg-test",
        "namespace":"websocket",
        "resourceVersion":"9410238",
        "uid":"02255f79-546f-42ca-9974-1a350e7b8bf0"
    }
}`

func getBaseConfigMapInfo() types.ConfigMap {
	var cm types.ConfigMap
	err := json.Unmarshal([]byte(configMapData), &cm)
	if err != nil {
		hwlog.RunLog.Infof("unmarshal config map data failed:%v", err)
	}

	return cm
}

func setConfigMapMsg(msg *model.Message, cm types.ConfigMap) {
	data, err := json.Marshal(cm)
	if err != nil {
		fmt.Printf("marshal cm failed:%v", err)
		return
	}
	msg.KubeEdgeRouter = model.MessageRoute{
		Source:    "controller",
		Group:     "resource",
		Operation: "update",
		Resource:  "websocket/configmap/cfg-test",
	}

	msg.Header.ID = "90fca461-8d3f-43d7-9f44-0090b8d3389d"
	msg.Header.Timestamp = 1678505303009
	msg.Header.ResourceVersion = "3558793"
	msg.Header.Sync = true

	msg.FillContent(data)
}

func testConfigMapFailed() {
	var msg model.Message

	var cm types.ConfigMap
	cm = getBaseConfigMapInfo()

	cm.Name = "test-"
	setConfigMapMsg(&msg, cm)

	msgValidator := NewMsgValidator(msglistchecker.NewFdMsgHeaderValidator())
	var err error
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err.Error(), convey.ShouldContainSubstring, "ConfigMap.ObjectMeta.Name")

	cm.Name = "cfg-test"
	data := make([]byte, 2049)
	for i := 0; i < 2049; i++ {
		data[i] = byte(i % math.MaxUint8)
	}
	cm.Data["test"] = string(data)

	setConfigMapMsg(&msg, cm)

	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}
	convey.So(err.Error(), convey.ShouldContainSubstring, "failed on the 'max' tag")
	delete(cm.Data, "test")

	cm.Data["123"] = "1234"
	setConfigMapMsg(&msg, cm)
	if err = msgValidator.Check(&msg); err != nil {
		hwlog.RunLog.Errorf("check msg failed: %v", err)
	}

	convey.So(err.Error(), convey.ShouldContainSubstring, "configmap data key check failed")
}
