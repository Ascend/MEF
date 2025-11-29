// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package commands this file for edge control command uninstall
package commands

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/uninstall"
)

type uninstallCmd struct {
	netType     string
	ip          string
	port        int
	user        string
	cert        string
	testConnect bool
}

// UninstallCmd edge control command uninstall
func UninstallCmd() common.Command {
	return &uninstallCmd{}
}

// Name command name
func (cmd *uninstallCmd) Name() string {
	return common.Uninstall
}

// Description command description
func (cmd *uninstallCmd) Description() string {
	return common.UninstallDesc
}

// BindFlag command flag binding
func (cmd *uninstallCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *uninstallCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *uninstallCmd) Execute(ctx *common.Context) error {
	hwlog.RunLog.Info("start to uninstall")
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}
	uninstallFlow := uninstall.NewFlowUninstall(ctx.WorkPathMgr, ctx.ConfigPathMgr)
	if err := uninstallFlow.RunTasks(); err != nil {
		return fmt.Errorf("uninstall %s failed", constants.MEFEdgeName)
	}
	hwlog.RunLog.Infof("uninstall %s success", constants.MEFEdgeName)
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *uninstallCmd) PrintOpLogOk(user, ip string) {
	hwlog.OpLog.Infof("[%s@%s] uninstall %s success", user, ip, constants.MEFEdgeName)
}

// PrintOpLogFail print operation fail log
func (cmd *uninstallCmd) PrintOpLogFail(user, ip string) {
	hwlog.OpLog.Errorf("[%s@%s] uninstall %s failed", user, ip, constants.MEFEdgeName)
}
