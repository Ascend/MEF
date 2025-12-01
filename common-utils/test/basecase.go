// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package test for test case interface and 2 implementations
package test

import (
	"errors"
	"fmt"
	"os"
)

// TcModule test case interface
type TcModule interface {
	Setup() error
	Teardown()
}

// TcBase struct for test case base, init log only
type TcBase struct {
}

// Setup pre-processing
func (tb *TcBase) Setup() error {
	return InitLog()
}

// Teardown post-processing
func (tb *TcBase) Teardown() {
}

// TcBaseWithDb struct for test case base, includes init log and init db
type TcBaseWithDb struct {
	// the db handle needs to be released in teardown step
	dbHandle *os.File
	// If this parameter is not specified, the db file is generated in the tmp dir by default
	DbPath string
	// tables need to be created
	Tables []interface{}
}

// Setup pre-processing
func (tbd *TcBaseWithDb) Setup() error {
	if err := InitLog(); err != nil {
		return err
	}
	if tbd.DbPath != "" {
		const dbMode = 0600
		dbHandle, err := os.OpenFile(tbd.DbPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, dbMode)
		if err != nil {
			fmt.Printf("create db in specified path failed, %v\n", err)
			return errors.New("create db in specified path failed")
		}
		tbd.dbHandle = dbHandle
	} else {
		dbHandle, err := os.CreateTemp("", "test-*.db")
		if err != nil {
			fmt.Printf("create db in temp dir failed, %v\n", err)
			return errors.New("create db in temp dir failed")
		}
		tbd.dbHandle = dbHandle
		tbd.DbPath = dbHandle.Name()
	}

	return InitDb(tbd.DbPath, tbd.Tables...)
}

// Teardown post-processing
func (tbd *TcBaseWithDb) Teardown() {
	// Error encountered, continue
	if err := CloseDb(); err != nil {
		fmt.Printf("close db failed, %v\n", err)
	}

	if tbd.dbHandle != nil {
		if err := tbd.dbHandle.Close(); err != nil {
			fmt.Printf("close db handle failed, %v\n", err)
		}
	}

	if err := RemoveDb(tbd.DbPath); err != nil {
		fmt.Printf("remove db [%s] failed, %v\n", tbd.DbPath, err)
	}
}
