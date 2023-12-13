// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appmanager for package main test
package appmanager

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcBaseWithDb := &test.TcBaseWithDb{
		Tables: append(tables, &AppInfo{}, &AppInstance{}, &AppDaemonSet{}),
	}
	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb)
	test.RunWithPatches(tcBaseWithDb, m, patches)
}
