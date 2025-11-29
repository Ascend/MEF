// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
