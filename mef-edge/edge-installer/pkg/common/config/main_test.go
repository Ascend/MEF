// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config for package main test
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/mindx/common/database"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

const testPath = "/tmp"

var (
	testErr = errors.New("test error")
	dbMgr   = DbMgr{}
)

func TestMain(m *testing.M) {
	tables := make([]interface{}, 0)
	tcConfig := &TcConfig{
		tcBaseWithDb: &test.TcBaseWithDb{
			DbPath: "./test.db",
			Tables: append(tables, &Configuration{}),
		},
	}
	patches := gomonkey.ApplyFunc(database.GetDb, test.MockGetDb)
	test.RunWithPatches(tcConfig, m, patches)
}

// TcConfig struct for test case base
type TcConfig struct {
	tcBaseWithDb *test.TcBaseWithDb
}

// Setup pre-processing
func (tc *TcConfig) Setup() error {
	if err := tc.tcBaseWithDb.Setup(); err != nil {
		return err
	}

	dbPath, err := filepath.Abs(tc.tcBaseWithDb.DbPath)
	if err != nil {
		fmt.Printf("get db abs path failed: %v\n", err)
		return errors.New("get db abs path failed")
	}
	dbMgr.dbName = filepath.Base(dbPath)
	dbMgr.dbDir = filepath.Dir(dbPath)
	return nil
}

// Teardown post-processing
func (tc *TcConfig) Teardown() {
	xmlFile := filepath.Join("./", constants.VersionXml)
	jsonFile := filepath.Join("./", "edgecore.json")
	containerCfgFile := filepath.Join("/tmp", constants.Config, constants.EdgeOm, constants.ContainerCfgFile)
	podCfgFile := filepath.Join("/tmp", constants.Config, constants.EdgeOm, constants.PodCfgFile)
	cfgFile := filepath.Join("/tmp", constants.MEFEdgeName, constants.Config,
		constants.EdgeCore, constants.EdgeCoreJsonName)
	omCfgFile := filepath.Join("/tmp", constants.MEFEdgeName, constants.Config,
		constants.EdgeOm, constants.ContainerCfgFile)

	needCleaned := []string{tc.tcBaseWithDb.DbPath, xmlFile, jsonFile, containerCfgFile, podCfgFile, cfgFile, omCfgFile}
	for _, file := range needCleaned {
		if err := os.Remove(file); err != nil && errors.Is(err, os.ErrExist) {
			fmt.Printf("cleanup [%s] failed, error: %v", file, err)
		}
	}
}
