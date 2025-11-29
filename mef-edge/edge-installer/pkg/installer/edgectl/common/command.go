// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common this file for define edge control command interface
package common

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/path/pathmgr"
)

// Command edge control command interface
type Command interface {
	Name() string
	Description() string
	BindFlag() bool
	LockFlag() bool
	Execute(ctx *Context) error
	PrintOpLogOk(user, ip string)
	PrintOpLogFail(user, ip string)
}

// CommandWithRetCode [interface] execute cmd with return code
type CommandWithRetCode interface {
	Name() string
	Description() string
	BindFlag() bool
	LockFlag() bool
	Execute(ctx *Context) error
	ExecuteWithCode(ctx *Context) (int, error)
	PrintOpLogOk(user, ip string)
	PrintOpLogFail(user, ip string)
}

// Context edge control context
type Context struct {
	WorkPathMgr   *pathmgr.WorkPathMgr
	ConfigPathMgr *pathmgr.ConfigPathMgr
	Args          []string
}

// DefaultPrintOpLogOk default print operation success log
func DefaultPrintOpLogOk(cmd Command, user, ip string) {
	if hwlog.OpLog != nil {
		hwlog.OpLog.Infof("[%s@%s] command: %s, result: success", user, ip, cmd.Name())
	}
}

// DefaultPrintOpLogFail default print operation fail log
func DefaultPrintOpLogFail(cmd Command, user, ip string) {
	if hwlog.OpLog != nil {
		hwlog.OpLog.Errorf("[%s@%s] command: %s, result: failed", user, ip, cmd.Name())
	}
}
