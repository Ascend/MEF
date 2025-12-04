// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package innercommands for package main test
package innercommands

import (
	"testing"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/edgectl/common"
)

var ctx = &common.Context{
	WorkPathMgr:   pathmgr.NewWorkPathMgr("./"),
	ConfigPathMgr: pathmgr.NewConfigPathMgr("./"),
}

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}
