// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package path for package main test
package path

import (
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	patches := gomonkey.ApplyFuncReturn(filepath.EvalSymlinks, "./", nil)
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, patches)
}
