// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package commands for
package commands

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
)

const (
	defaultErrorCode = 255
	fdWithOmCode     = 0
	fdWithOutOmCode  = 1
	otherNetCode     = 2
)

type getNetCfgCmd struct {
}

// GetNetCfgCmd edge control command for get net config
func GetNetCfgCmd() common.Command {
	return &getNetCfgCmd{}
}

// Name command name
func (cmd *getNetCfgCmd) Name() string {
	return common.GetNetCfg
}

// Description command description
func (cmd *getNetCfgCmd) Description() string {
	return common.GetNetCfgDesc
}

// BindFlag command flag binding
func (cmd *getNetCfgCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *getNetCfgCmd) LockFlag() bool {
	return false
}

// Execute execute command
func (cmd *getNetCfgCmd) Execute(ctx *common.Context) error {
	return nil
}

// ExecuteWithCode Execute command with code
func (cmd *getNetCfgCmd) ExecuteWithCode(ctx *common.Context) (int, error) {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return defaultErrorCode, errors.New("ctx is nil")
	}
	return cmd.getNetConfig()
}

// PrintOpLogOk print operation success log
func (cmd *getNetCfgCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *getNetCfgCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}

func (cmd *getNetCfgCmd) getNetConfig() (int, error) {
	dbMgr, err := config.GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		hwlog.RunLog.Errorf("get component db manager failed, error: %v", err)
		return defaultErrorCode, errors.New("get component db manager failed")
	}
	netManager, err := config.GetNetManager(dbMgr)
	if err != nil {
		return defaultErrorCode, err
	}
	hwlog.RunLog.Info("get net config success")
	if netManager.NetType == constants.FD && netManager.WithOm {
		hwlog.RunLog.Info("current net type is [FD] and with om is [True]")
		fmt.Println("current net type is [FD] and with om is [True]")
		return fdWithOmCode, nil
	}
	if netManager.NetType == constants.FD && !netManager.WithOm {
		hwlog.RunLog.Info("current net type is [FD] and with om is [False]")
		fmt.Println("current net type is [FD] and with om is [False]")
		return fdWithOutOmCode, nil
	}
	hwlog.RunLog.Infof("current net type is [%s]", netManager.NetType)
	fmt.Printf("current net type is [%s]\n", netManager.NetType)
	return otherNetCode, nil
}
