// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
