// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package commands this file for edge control command get alarm configuration
package commands

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
)

type getAlarmCfgCmd struct {
}

// GetAlarmCfgCmd edge control command for get alarm config
func GetAlarmCfgCmd() common.Command {
	return &getAlarmCfgCmd{}
}

// Name command name
func (cmd *getAlarmCfgCmd) Name() string {
	return common.GetAlarmCfg
}

// Description command description
func (cmd *getAlarmCfgCmd) Description() string {
	return common.GetAlarmCfgDesc
}

// BindFlag command flag binding
func (cmd *getAlarmCfgCmd) BindFlag() bool {
	return false
}

// LockFlag command lock flag
func (cmd *getAlarmCfgCmd) LockFlag() bool {
	return true
}

// Execute execute command
func (cmd *getAlarmCfgCmd) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	dbMgr, err := config.GetComponentDbMgr(constants.EdgeOm)
	if err != nil {
		return err
	}

	alarmCfgs := []struct {
		cfgInDb string
		cfgCmd  string
		unit    string
	}{
		{constants.CertCheckPeriodDB, common.CertCheckPeriodCmd, common.UnitDay},
		{constants.CertOverdueThresholdDB, common.CertOverdueThresholdCmd, common.UnitDay},
	}
	for _, alarmCfg := range alarmCfgs {
		cfg, err := dbMgr.GetAlarmConfig(alarmCfg.cfgInDb)
		if err != nil {
			return err
		}
		fmt.Printf("%s is [%v], the unit is %s\n", alarmCfg.cfgCmd, cfg, alarmCfg.unit)
	}

	return nil
}

// PrintOpLogOk print operation success log
func (cmd *getAlarmCfgCmd) PrintOpLogOk(user, ip string) {
	common.DefaultPrintOpLogOk(cmd, user, ip)
}

// PrintOpLogFail print operation fail log
func (cmd *getAlarmCfgCmd) PrintOpLogFail(user, ip string) {
	common.DefaultPrintOpLogFail(cmd, user, ip)
}
