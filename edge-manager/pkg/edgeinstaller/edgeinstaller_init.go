// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller module edgeinstaller init
package edgeinstaller

import (
	"context"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/database"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// Installer edge-installer struct
type Installer struct {
	ctx    context.Context
	enable bool
}

// Name returns the name of edge installer module
func (i *Installer) Name() string {
	return common.EdgeInstallerName
}

// Enable indicates whether this module is enabled
func (i *Installer) Enable() bool {
	if !i.enable {
		return !i.enable
	}
	if err := initSoftwareMgrInfoTable(); err != nil {
		hwlog.RunLog.Errorf("module [%s] init database table failed, error: %v, cannot enable",
			common.EdgeInstaller, err)
		return !i.enable
	}
	return i.enable
}

// NewInstaller new Installer
func NewInstaller(enable bool) *Installer {
	i := &Installer{
		enable: enable,
		ctx:    context.Background(),
	}
	return i
}

// Start receives and sends message
func (i *Installer) Start() {
	for {
		select {
		case _, ok := <-i.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		req, err := modulemanager.ReceiveMessage(i.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message from channel failed, error: %v", i.Name(), err)
			continue
		}
		go i.dispatchMsg(req)
	}
}

func (i *Installer) dispatchMsg(msg *model.Message) {
	handlerMgr := GetHandlerMgr()
	if err := handlerMgr.Process(msg); err != nil {
		return
	}
}

func initSoftwareMgrInfoTable() error {
	if err := database.CreateTableIfNotExists(EdgeAccountInfo{}); err != nil {
		hwlog.RunLog.Error("table edge_account_infos create failed")
		return err
	}

	if err := database.CreateTableIfNotExists(SoftwareMgrInfo{}); err != nil {
		hwlog.RunLog.Error("table software_mgr_infos create failed")
		return err
	}

	if err := CreateTableSfwInfo(); err != nil {
		hwlog.RunLog.Error("create item in table software manager info failed")
		return err
	}

	return nil
}
