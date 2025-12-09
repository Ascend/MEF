// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common
package common

import (
	"testing"
)

type testCmd struct {
}

// Name command name
func (cmd *testCmd) Name() string {
	return "test"
}

// Description command description
func (cmd *testCmd) Description() string {
	return "testDesc"
}

// BindFlag command flag binding
func (cmd *testCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *testCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *testCmd) Execute(ctx *Context) error {
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *testCmd) PrintOpLogOk(user, ip string) {
	DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *testCmd) PrintOpLogFail(user, ip string) {
	DefaultPrintOpLogFail(cmd, user, ip)
}

func TestDefaultPrintOpLogOk(t *testing.T) {
	DefaultPrintOpLogOk(&testCmd{}, "root", "127.0.0.1")
}

func TestDefaultPrintOpLogFail(t *testing.T) {
	DefaultPrintOpLogFail(&testCmd{}, "root", "127.0.0.1")
}
