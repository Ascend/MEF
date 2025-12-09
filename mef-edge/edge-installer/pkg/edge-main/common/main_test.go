// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
