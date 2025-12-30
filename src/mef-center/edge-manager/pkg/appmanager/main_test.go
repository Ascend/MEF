// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appmanager for package main test
package appmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-manager/pkg/util"
)

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcBaseWithDb := &test.TcBaseWithDb{
		Tables: append(tables, &AppInfo{}, &AppInstance{}, &AppDaemonSet{}),
	}
	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb).
		ApplyFuncReturn(util.InWhiteList, true)
	test.RunWithPatches(tcBaseWithDb, m, patches)
}

func newMsgWithContentForUT(v interface{}) *model.Message {
	msg, err := model.NewMessage()
	if err != nil {
		panic(err)
	}
	err = msg.FillContent(v)
	if err != nil {
		panic(err)
	}
	return msg
}
