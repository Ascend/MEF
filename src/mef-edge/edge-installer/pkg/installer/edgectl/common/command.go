// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
