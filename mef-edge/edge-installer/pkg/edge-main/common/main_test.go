// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package common for main test
package common

import (
	"testing"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

var (
	source    = "testSource"
	group     = "testGroup"
	operation = constants.OptUpdate
	resource  = "/test/resource"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

func getTestMsg() *model.Message {
	testMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("new test msg failed: %v", err)
		return nil
	}
	testMsg.Header.ID = testMsg.Header.Id
	testMsg.SetKubeEdgeRouter(source, group, operation, resource)
	return testMsg
}
