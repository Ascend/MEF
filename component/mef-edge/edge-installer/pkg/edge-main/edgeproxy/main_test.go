// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgeproxy for package test main
package edgeproxy

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/edge-main/job"
)

const WaitingDuration = 3 * time.Second

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcEdgeProxy := &TcEdgeProxy{
		tcBaseWithDb: &test.TcBaseWithDb{
			Tables: append(tables, &Meta{}),
		},
	}
	patches := gomonkey.ApplyFuncReturn(job.SyncNodeStatus, nil).
		ApplyFunc(database.GetDb, test.MockGetDb)
	test.RunWithPatches(tcEdgeProxy, m, patches)
}

// TcBase struct for test case base, init log only
type TcEdgeProxy struct {
	tcBaseWithDb *test.TcBaseWithDb
}

// Setup pre-processing
func (tc *TcEdgeProxy) Setup() error {
	moduleInit()
	RegistryMsgRouters()

	if err := tc.tcBaseWithDb.Setup(); err != nil {
		return err
	}
	meta := Meta{Type: "test", Key: "key", Value: "value"}
	return database.GetDb().Create(&meta).Error
}

// Teardown post-processing
func (tc *TcEdgeProxy) Teardown() {
	tc.tcBaseWithDb.Teardown()
}

func moduleInit() {
	modulemgr.ModuleInit()
	modules := []model.Module{
		NewDeviceOmProxy(true),
		NewEdgeOmProxy(true),
		NewEdgeCoreProxy(true),
	}

	for _, mod := range modules {
		if err := modulemgr.Registry(mod); err != nil {
			panic(err)
		}
	}
}

type Meta struct {
	Key   string `gorm:"column:key; size:256; primaryKey"`
	Type  string `gorm:"column:type; size:32"`
	Value string `gorm:"column:value; type:text"`
}
