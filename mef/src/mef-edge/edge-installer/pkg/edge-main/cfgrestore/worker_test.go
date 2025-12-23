// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package cfgrestore
package cfgrestore

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	commondatabase "huawei.com/mindx/common/database"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/database"
)

var okMessage = &model.Message{Content: model.RawMessage(constants.OK)}

func TestMain(m *testing.M) {
	tcBaseWithDb := &test.TcBaseWithDb{
		DbPath: ":memory:?cache=shared",
	}
	patches := gomonkey.ApplyFunc(commondatabase.GetDb, test.MockGetDb)
	test.RunWithPatches(tcBaseWithDb, m, patches)
}

func initDb() error {
	mockDb := test.MockGetDb()
	if mockDb == nil {
		return errors.New("nil pointer")
	}
	if err := mockDb.Exec("drop table if exists metas").Error; err != nil {
		return err
	}
	if err := mockDb.AutoMigrate(&database.Meta{}); err != nil {
		return err
	}
	metas := []database.Meta{
		{Key: "websocket/pod/a", Type: constants.ResourceTypePod, Value: "{}"},
		{Key: "websocket/secret/fusion-director-docker-registry-secret", Type: constants.ResourceTypeSecret, Value: "{}"},
		{Key: "websocket/secret/b", Type: constants.ResourceTypeSecret, Value: "{}"},
		{Key: "websocket/configmap/c", Type: constants.ResourceTypeConfigMap, Value: "{}"},
	}
	for _, meta := range metas {
		if err := database.GetMetaRepository().CreateOrUpdate(meta); err != nil {
			return err
		}
	}
	return nil
}

func TestDeletePodsData(t *testing.T) {
	convey.Convey("test delete pods data successful", t, testDeletePodsDataSuccessful)
	convey.Convey("test delete pods data timeout", t, testDeletePodsDataTimeout)
}

func testDeletePodsDataSuccessful() {
	convey.Convey("delete pods data", func() {
		err := initDb()
		convey.So(err, convey.ShouldBeNil)

		patches := gomonkey.
			ApplyFuncReturn(modulemgr.SendSyncMessage, okMessage, nil).
			ApplyFunc(modulemgr.SendAsyncMessage, expectConfigResult(
				constants.ResultProcessing, constants.ResultProcessing, constants.ResultProcessing))
		defer patches.Reset()

		deletePodsData()

		pods, err := database.GetMetaRepository().GetByType(constants.ResourceTypePod)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(pods), convey.ShouldEqual, 0)
		secrets, err := database.GetMetaRepository().GetByType(constants.ResourceTypeSecret)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(secrets), convey.ShouldEqual, 0)
		configMaps, err := database.GetMetaRepository().GetByType(constants.ResourceTypeConfigMap)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(configMaps), convey.ShouldEqual, 0)
	})
}

func testDeletePodsDataTimeout() {
	convey.Convey("delete pods data", func() {
		const (
			oneTime  = 1
			twoTimes = 2
		)
		err := initDb()
		convey.So(err, convey.ShouldBeNil)

		patches := gomonkey.
			ApplyFuncReturn(modulemgr.SendSyncMessage, okMessage, errors.New("timeout")).
			ApplyFunc(modulemgr.SendAsyncMessage, expectConfigResult(constants.ResultFailed))
		defer patches.Reset()

		deletePodsData()

		pods, err := database.GetMetaRepository().GetByType(constants.ResourceTypePod)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(pods), convey.ShouldEqual, oneTime)
		secrets, err := database.GetMetaRepository().GetByType(constants.ResourceTypeSecret)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(secrets), convey.ShouldEqual, twoTimes)
		configMaps, err := database.GetMetaRepository().GetByType(constants.ResourceTypeConfigMap)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(configMaps), convey.ShouldEqual, oneTime)
	})
}

func expectConfigResult(result ...string) func(message *model.Message) error {
	var invokeCount int
	return func(message *model.Message) error {
		invokeCount++
		convey.So(len(result), convey.ShouldBeGreaterThanOrEqualTo, invokeCount)
		var contentStr string
		err := message.ParseContent(&contentStr)
		convey.So(err, convey.ShouldBeNil)
		var contentData config.ProgressTip
		if err := json.Unmarshal([]byte(contentStr), &contentData); err != nil {
			convey.So(err, convey.ShouldBeNil)
			return err
		}
		convey.So(contentData.Result, convey.ShouldEqual, result[invokeCount-1])
		return nil
	}
}
