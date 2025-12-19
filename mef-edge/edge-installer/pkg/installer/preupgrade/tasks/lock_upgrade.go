// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for lock upgrade task
package tasks

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/common"
)

type lockUpgrade struct{}

// LockUpgrade lock upgrade installer
func LockUpgrade() common.Task {
	return &lockUpgrade{}
}

// Run task
func (l *lockUpgrade) Run() error {
	upgradeFlag := util.FlagLockInstance(constants.FlagPath, constants.ProcessFlag, constants.Upgrade)
	if err := upgradeFlag.Lock(); err != nil {
		return errors.New("lock upgrade failed,there may be another process processing the upgrade")
	}
	hwlog.RunLog.Info("lock upgrade success")
	return nil
}
