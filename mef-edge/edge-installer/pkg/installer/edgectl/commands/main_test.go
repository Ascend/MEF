// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package commands for package main test
package commands

import (
	"os"
	"path/filepath"
	"testing"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/edgectl/common"
)

const (
	userRoot    = "root"
	ipLocalhost = "localhost"
	testRootDir = "/tmp"
)

var (
	cmd common.Command
	ctx = &common.Context{
		WorkPathMgr:   pathmgr.NewWorkPathMgr("./"),
		ConfigPathMgr: pathmgr.NewConfigPathMgr("./"),
	}
)

func TestMain(m *testing.M) {
	dbDir := filepath.Join(testRootDir, constants.Config, constants.EdgeMain)
	if err := os.MkdirAll(dbDir, constants.Mode600); err != nil {
		panic(err)
	}
	tables := make([]interface{}, 0)
	tcBaseWithDb := &test.TcBaseWithDb{
		Tables: append(tables, &config.AlarmConfig{}),
	}
	test.RunWithPatches(tcBaseWithDb, m, nil)
}
