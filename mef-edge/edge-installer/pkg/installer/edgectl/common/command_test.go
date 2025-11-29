// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
