// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for adding user account
package tasks

import (
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
)

// AddUserAccountTask the task for add user account
type AddUserAccountTask struct{}

// Run add user account task
func (aua *AddUserAccountTask) Run() error {
	var addFunc = []func() error{
		aua.addEdgeUserAccount,
	}
	for _, function := range addFunc {
		if err := function(); err != nil {
			hwlog.RunLog.Error(err)
			return err
		}
	}
	return nil
}

func (aua *AddUserAccountTask) addEdgeUserAccount() error {
	userMgr := envutils.NewUserMgr(constants.EdgeUserName, constants.EdgeUserGroup,
		constants.EdgeUserUid, constants.EdgeUserGid)
	if err := userMgr.AddUserAccount(); err != nil {
		hwlog.RunLog.Errorf("add user account [%s] failed", constants.EdgeUserName)
		return err
	}

	hwlog.RunLog.Info("add user account success")
	return nil
}
