// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package commands this file for edge control command update crl
package commands

import (
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/veripkgutils"
	"edge-installer/pkg/installer/edgectl/common"
)

type updateCrlCmd struct {
	crlPath string
}

// UpdateCrlCmd edge control command update crl
func UpdateCrlCmd() common.Command {
	return &updateCrlCmd{}
}

// Name command name
func (cmd *updateCrlCmd) Name() string {
	return common.UpdateCrl
}

// Description command description
func (cmd *updateCrlCmd) Description() string {
	return common.UpdateCrlDesc
}

// BindFlag command flag binding
func (cmd *updateCrlCmd) BindFlag() bool {
	flag.StringVar(&(cmd.crlPath), "crl_path", "", "new crl path")
	utils.MarkFlagRequired("crl_path")
	return true
}

// LockFlag command lock flag
func (cmd *updateCrlCmd) LockFlag() bool {
	return true
}

// UpdateCrlFlow update crl flow
type UpdateCrlFlow struct {
	param Param
}

// Param the parameters for updating crl
type Param struct {
	CrlPath string
}

// NewUpdateCrlFlow create update crl flow instance
func NewUpdateCrlFlow(param Param) *UpdateCrlFlow {
	return &UpdateCrlFlow{param: param}
}

// Execute execute command
func (cmd *updateCrlCmd) Execute(ctx *common.Context) error {
	hwlog.RunLog.Info("start to update crl")
	fmt.Println("start to update crl...")

	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}
	if err := cmd.checkParam(); err != nil {
		return fmt.Errorf("check param failed, error: %v", err)
	}

	if err := common.InitEdgeOmResource(); err != nil {
		return fmt.Errorf("init resource failed, error: %v", err)
	}

	param := Param{
		CrlPath: cmd.crlPath,
	}
	flow := NewUpdateCrlFlow(param)
	if err := flow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("update crl failed, error: %v", err)
		return fmt.Errorf("update crl failed, error: %v", err)
	}

	hwlog.RunLog.Info("update crl success")
	return nil
}

// RunTasks run update crl task
func (cmf UpdateCrlFlow) RunTasks() error {
	crlPath := cmf.param.CrlPath

	needUpdateCrl, verifyCrl, err := veripkgutils.PrepareVerifyCrl(crlPath)
	if err != nil {
		hwlog.RunLog.Errorf("prepare crl for verifying package failed, error: %v", err)
		return err
	}

	if needUpdateCrl {
		if err = veripkgutils.UpdateLocalCrl(verifyCrl); err != nil {
			hwlog.RunLog.Errorf("update crl file failed, error: %v", err)
			return errors.New("update crl file failed")
		}
	}
	return nil
}

// PrintOpLogOk print operation success log
func (cmd *updateCrlCmd) PrintOpLogOk(user, ip string) {
	hwlog.OpLog.Infof("[%s@%s] update crl success", user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *updateCrlCmd) PrintOpLogFail(user, ip string) {
	hwlog.OpLog.Errorf("[%s@%s] update crl failed", user, ip)
}

func (cmd *updateCrlCmd) checkParam() error {
	if _, err := fileutils.RealFileCheck(cmd.crlPath, true, false, constants.MaxCertSize); err != nil {
		hwlog.RunLog.Errorf("check param crl_path failed, error: %v", err)
		return errors.New("param crl_path is invalid")
	}

	return nil
}
