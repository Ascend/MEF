// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_A500

// Package innercommands for this file for inner control command
package innercommands

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	com "edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/edgectl/common"
)

type copyResetScriptCmd struct {
}

// CopyResetScriptCmd is to init a copyResetScriptCmd struct which is to copy reset script to filesystem P7
func CopyResetScriptCmd() common.Command {
	return &copyResetScriptCmd{}
}

// Name command name
func (cmd *copyResetScriptCmd) Name() string {
	return common.CopyResetScriptCmd
}

// Description command description
func (cmd *copyResetScriptCmd) Description() string {
	return common.CopyResetScriptDesc
}

// BindFlag command flag binding
func (cmd *copyResetScriptCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *copyResetScriptCmd) LockFlag() bool {
	return false
}

// Execute execute command
func (cmd *copyResetScriptCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	if err := com.CopyResetScriptToP7(); err != nil {
		hwlog.RunLog.Error(err)
		return err
	}

	hwlog.RunLog.Info("copy reset script successful")
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *copyResetScriptCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *copyResetScriptCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

type restoreCfgCmd struct {
}

// RestoreCfgCmd is to init a restoreCfgCmd struct which is to restore default config
func RestoreCfgCmd() common.Command {
	return &restoreCfgCmd{}
}

// Name command name
func (cmd *restoreCfgCmd) Name() string {
	return common.RestoreCfgCmd
}

// Description command description
func (cmd *restoreCfgCmd) Description() string {
	return common.RestoreCfgDesc
}

// BindFlag command flag binding
func (cmd *restoreCfgCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *restoreCfgCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *restoreCfgCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	restoreCfg := NewRestoreCfg(ctx.ConfigPathMgr)
	if err := restoreCfg.Run(); err != nil {
		return err
	}

	hwlog.RunLog.Info("restore default config successful")
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *restoreCfgCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *restoreCfgCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}
