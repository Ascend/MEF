// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util for package main test
package util

import (
	"os"
	"syscall"
	"testing"
	"time"

	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

type mockFileInfo struct {
}

func (fi *mockFileInfo) Name() string {
	return ""
}

func (fi *mockFileInfo) Size() int64 {
	return 0
}

func (fi *mockFileInfo) Mode() os.FileMode {
	const fileMode = 0600
	return fileMode
}

func (fi *mockFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (fi *mockFileInfo) IsDir() bool {
	return false
}

func (fi *mockFileInfo) Sys() interface{} {
	return &syscall.Stat_t{}
}
