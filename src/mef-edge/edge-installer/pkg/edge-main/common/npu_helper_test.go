// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common for test npu helper
package common

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/database"
)

var mockDbContent = `{"status":{
        "capacity":{
            "cpu":"4",
            "huawei.com/Ascend310":100,
            "memory":"11856388Ki",
            "pods":"110"
        }
    }}`

func TestLoadNpuFromDb(t *testing.T) {
	metas := []database.Meta{{Key: "websocket/node/test", Type: constants.ResourceTypeNode, Value: mockDbContent}}
	p := gomonkey.ApplyFuncReturn(database.GetMetaRepository, &mockMetaRepository{}).
		ApplyMethodReturn(&mockMetaRepository{}, "GetByType", metas, nil)
	defer p.Reset()
	convey.Convey("load npu form db should be success", t, loadNpuFromDbSuccess)
	convey.Convey("load npu form db should be failed, get node from db failed", t, getNodeFromDbFailed)
	convey.Convey("load npu form db should be failed, meta count not correct", t, metaCountNotCorrect)
	convey.Convey("load npu form db should be failed, get content map failed", t, getContentMapFailed)
}

func loadNpuFromDbSuccess() {
	npuName, ok := LoadNpuFromDb()
	convey.So(npuName, convey.ShouldEqual, "huawei.com/Ascend310")
	convey.So(ok, convey.ShouldBeTrue)
}

func getNodeFromDbFailed() {
	p := gomonkey.ApplyMethodReturn(&mockMetaRepository{}, "GetByType", []database.Meta{}, test.ErrTest)
	defer p.Reset()
	npuName, ok := LoadNpuFromDb()
	convey.So(npuName, convey.ShouldBeBlank)
	convey.So(ok, convey.ShouldBeFalse)
}

func metaCountNotCorrect() {
	metas := []database.Meta{
		{Key: "websocket/node/test1", Type: constants.ResourceTypeNode, Value: mockDbContent},
		{Key: "websocket/node/test2", Type: constants.ResourceTypeNode, Value: mockDbContent},
	}
	p := gomonkey.ApplyMethodReturn(&mockMetaRepository{}, "GetByType", metas, nil)
	defer p.Reset()
	npuName, ok := LoadNpuFromDb()
	convey.So(npuName, convey.ShouldBeBlank)
	convey.So(ok, convey.ShouldBeFalse)
}

func getContentMapFailed() {
	p := gomonkey.ApplyFuncReturn(util.GetContentMap, nil, test.ErrTest)
	defer p.Reset()
	npuName, ok := LoadNpuFromDb()
	convey.So(npuName, convey.ShouldBeBlank)
	convey.So(ok, convey.ShouldBeFalse)
}

type mockMetaRepository struct{}

func (m *mockMetaRepository) GetByKey(_ string) (database.Meta, error) { return database.Meta{}, nil }
func (m *mockMetaRepository) GetByType(_ string) ([]database.Meta, error) {
	return []database.Meta{}, nil
}
func (m *mockMetaRepository) GetKeyByType(_ string) ([]string, error) { return []string{}, nil }
func (m *mockMetaRepository) DeleteByKey(_ string) error              { return nil }
func (m *mockMetaRepository) CreateOrUpdate(_ database.Meta) error    { return nil }
func (m *mockMetaRepository) CountByType(_ string) (int64, error)     { return 0, nil }
