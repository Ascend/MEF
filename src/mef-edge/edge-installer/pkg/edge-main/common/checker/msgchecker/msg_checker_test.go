// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package mefmsgchecker
package msgchecker

import (
	"testing"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/edge-main/common/configpara"

	"github.com/agiledragon/gomonkey/v2"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcMsgCheck := &TcMsgCheck{
		tcBaseWithDb: &test.TcBaseWithDb{
			Tables: append(tables, &Meta{}),
		},
	}

	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb).
		ApplyFunc(configpara.GetPodConfig, MockPodConfig)

	test.RunWithPatches(tcMsgCheck, m, patches)
}

type Meta struct {
	Key   string `gorm:"column:key; size:256; primaryKey"`
	Type  string `gorm:"column:type; size:32"`
	Value string `gorm:"column:value; type:text"`
}

// TcBase struct for test case base, init log only
type TcMsgCheck struct {
	tcBaseWithDb *test.TcBaseWithDb
}

// Setup pre-processing
func (tc *TcMsgCheck) Setup() error {
	if err := tc.tcBaseWithDb.Setup(); err != nil {
		return err
	}
	nodeData := `
{
    "metadata":{
        "name":"2102314nmv10p7100006"
    },
    "status":{
        "allocatable":{
            "cpu":"3",
            "huawei.com/Ascend310":"2",
            "memory":"10240Mi"
        },
        "capacity":{
            "cpu":"4",
            "huawei.com/Ascend310":"2",
            "memory":"10640Mi"
        }
    }
}`
	meta := Meta{Type: "node", Key: "default/node/2102314nmv10p7100006", Value: nodeData}
	return database.GetDb().Create(&meta).Error
}

// Teardown post-processing
func (tc *TcMsgCheck) Teardown() {
	tc.tcBaseWithDb.Teardown()
}

func MockPodConfig() config.PodConfig {
	cfg := config.PodConfig{}
	cfg.MaxContainerNumber = 20
	cfg.HostPath = []string{
		"/usr/lib64/libstackcore.so",
		"/usr/local/Ascend/driver/lib64",
		"/var/lib/docker/modelfile"}
	return cfg
}
