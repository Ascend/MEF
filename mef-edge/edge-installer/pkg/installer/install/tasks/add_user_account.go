// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
