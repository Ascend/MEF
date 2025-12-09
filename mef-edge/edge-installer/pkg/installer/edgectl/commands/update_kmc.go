// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package commands this file for edge control command
package commands

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/kmcupdate"
)

type updateKmc struct {
}

// UpdateKmcCmd is used to update all kmc RK/MK
func UpdateKmcCmd() common.Command {
	return &updateKmc{}
}

// Name command name
func (cmd *updateKmc) Name() string {
	return common.UpdateKmc
}

// Description command description
func (cmd *updateKmc) Description() string {
	return common.UpdateKmcDesc
}

// BindFlag command flag binding
func (cmd *updateKmc) BindFlag() bool {
	return false
}

// Execute execute command
func (cmd *updateKmc) Execute(ctx *common.Context) error {
	if ctx == nil {
		hwlog.RunLog.Error("ctx is nil")
		return errors.New("ctx is nil")
	}

	updateFlow := kmcupdate.NewUpdateKmcFlow(ctx.ConfigPathMgr)
	if err := updateFlow.RunFlow(); err != nil {
		return err
	}

	return nil
}

// LockFlag command lock flag
func (cmd *updateKmc) LockFlag() bool {
	return false
}

// PrintOpLogOk print operation success log
func (cmd *updateKmc) PrintOpLogOk(user, ip string) {
	hwlog.OpLog.Infof("[%s@%s] update %s kmc keys success", user, ip, constants.MEFEdgeName)
}

// PrintOpLogFail print operation fail log
func (cmd *updateKmc) PrintOpLogFail(user, ip string) {
	hwlog.OpLog.Errorf("[%s@%s] update %s kmc keys failed", user, ip, constants.MEFEdgeName)
}
