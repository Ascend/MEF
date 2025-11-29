// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package upgrade this file for upgrade module
package upgrade

import (
	"context"
	"sync/atomic"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/edge-om/upgrade/handlers"
	"edge-installer/pkg/edge-om/upgrade/reporter"
)

var goroutineCount int32

const (
	maxGoroutineNumber         = 100
	softwareVersionReportCount = 3
)

type upgradeMgr struct {
	enable bool
	ctx    context.Context
}

// NewUpgradeMgr new upgrade manager
func NewUpgradeMgr(ctx context.Context, enable bool) model.Module {
	um := &upgradeMgr{
		enable: false,
		ctx:    ctx,
	}
	edgeOmCfg, err := path.GetCompConfigDir()
	if err != nil {
		hwlog.RunLog.Errorf("get config dir failed: %v", err)
		return um
	}
	dbMgr := config.NewDbMgr(edgeOmCfg, constants.DbEdgeOmPath)
	manager, err := config.GetNetManager(dbMgr)
	if err != nil {
		hwlog.RunLog.Errorf("check net manager failed: %s", err.Error())
		return um
	}
	if manager.NetType != constants.MEF {
		hwlog.RunLog.Info("net manager type is not MEF, upgrade manager will not enabled")
		return um
	}
	um.enable = enable
	return um
}

// Name returns the name of upgrade module
func (u *upgradeMgr) Name() string {
	return constants.UpgradeManagerName
}

// Enable indicates whether this module is enabled
func (u *upgradeMgr) Enable() bool {
	return u.enable
}

// Start receives and sends message
func (u *upgradeMgr) Start() {
	const startWaitTime = 15 * time.Second

	time.Sleep(startWaitTime)
	go reporter.ReportSoftwareVersion(softwareVersionReportCount)
	for {
		select {
		case <-u.ctx.Done():
			hwlog.RunLog.Info("----------------upgrade manager exit-------------------")
			return
		default:
		}

		req, err := modulemgr.ReceiveMessage(u.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receives request failed", u.Name())
			continue
		}

		hwlog.RunLog.Infof("upgrade receive msg option:[%s] resource:[%s]", req.GetOption(), req.GetResource())

		if goroutineCount >= maxGoroutineNumber {
			hwlog.RunLog.Warnf("message handler is out of routine limit, discard message, router: %v", req.Router)
			continue
		}
		atomic.AddInt32(&goroutineCount, 1)
		go u.dispatchMsg(req)
	}
}

func (u *upgradeMgr) dispatchMsg(msg *model.Message) {
	defer func() {
		atomic.AddInt32(&goroutineCount, -1)
	}()
	handlerMgr := handlers.GetHandlerMgr()
	err := handlerMgr.Process(msg)
	if err != nil {
		hwlog.RunLog.Errorf("process msg failed: %v", err)
		return
	}
}
