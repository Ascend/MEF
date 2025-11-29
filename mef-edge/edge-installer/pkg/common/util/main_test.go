// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
