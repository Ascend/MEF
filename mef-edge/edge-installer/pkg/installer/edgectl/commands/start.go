// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package commands this file for edge control command start
package commands

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/util"
	com "edge-installer/pkg/installer/common"
	"edge-installer/pkg/installer/edgectl/common"
)

type startCmd struct {
}

// StartCmd edge control command start
func StartCmd() common.Command {
	return &startCmd{}
}

// Name command name
func (cmd *startCmd) Name() string {
	return common.Start
}

// Description command description
func (cmd *startCmd) Description() string {
	return common.StartDesc
}

// BindFlag command flag binding
func (cmd *startCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *startCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *startCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}
	mgr := com.NewComponentMgr(ctx.WorkPathMgr.GetInstallRootDir())
	for _, component := range mgr.GetComponents() {
		if component.IsExist() && util.IsServiceActive(component.Service.Name) {
			fmt.Printf("warning: component [%s] is already started!\n", component.Name)
		}
	}
	if err := mgr.StartAll(); err != nil {
		return err
	}
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *startCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *startCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}
