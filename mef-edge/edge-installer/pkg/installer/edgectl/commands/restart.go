// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package commands this file for edge control command restart
package commands

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	com "edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/edgectl/common"
)

type restartCmd struct {
}

// RestartCmd edge control command restart
func RestartCmd() common.Command {
	return &restartCmd{}
}

// Name command name
func (cmd *restartCmd) Name() string {
	return common.Restart
}

// Description command description
func (cmd *restartCmd) Description() string {
	return common.RestartDesc
}

// BindFlag command flag binding
func (cmd *restartCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *restartCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *restartCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}
	if err := com.NewComponentMgr(ctx.WorkPathMgr.GetInstallRootDir()).RestartAll(); err != nil {
		return err
	}
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *restartCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *restartCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}
