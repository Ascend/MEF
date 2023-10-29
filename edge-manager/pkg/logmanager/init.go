// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logmanager enables collecting logs
package logmanager

import (
	"context"
	"sync"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/handler"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/handlers"
	"edge-manager/pkg/logmanager/tasks"
	"edge-manager/pkg/logmanager/utils"
)

// NewLogManager creates a new log manager
func NewLogManager(ctx context.Context, enable bool) model.Module {
	return &logMgr{
		enable: enable,
		ctx:    ctx,
	}
}

type logMgr struct {
	enable         bool
	ctx            context.Context
	msgHandler     handler.MsgHandler
	msgHandlerOnce sync.Once
}

// Name returns name of module
func (l *logMgr) Name() string {
	return constants.LogManagerName
}

// Enable does preparation for module
func (l *logMgr) Enable() bool {
	if l.enable {
		if err := tasks.InitTasks(); err != nil {
			hwlog.RunLog.Errorf("failed to init task scheduler, %v", err)
			return !l.enable
		}
		_, err := utils.CleanTempFiles()
		if err != nil {
			hwlog.RunLog.Warnf("clean temp files failed, error: %v", err)
		}
	}
	return l.enable
}

// Start starts module
func (l *logMgr) Start() {
	hwlog.RunLog.Infof("module [%s] started", l.Name())
	for {
		select {
		case <-l.ctx.Done():
			hwlog.RunLog.Infof("module [%s] exited", l.Name())
			return
		default:
		}

		req, err := modulemgr.ReceiveMessage(l.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message from channel failed, error: %v", l.Name(), err)
			continue
		}
		if err := l.getMsgHandler().Process(req); err != nil {
			hwlog.RunLog.Errorf("failed to process message, %v", err)
		}
	}
}

func (l *logMgr) getMsgHandler() *handler.MsgHandler {
	l.msgHandlerOnce.Do(l.registerHandlers)
	return &l.msgHandler
}

func (l *logMgr) registerHandlers() {
	l.msgHandler.Register(handler.RegisterInfo{
		MsgOpt:  common.OptReport,
		MsgRes:  constants.ResLogDumpError,
		Handler: handlers.NewReportErrorHandler(),
	})
	l.msgHandler.Register(handler.RegisterInfo{
		MsgOpt:  common.OptPost,
		MsgRes:  constants.LogDumpUrlPrefix + constants.ResTask,
		Handler: handlers.NewCreateTaskHandler(),
	})
	l.msgHandler.Register(handler.RegisterInfo{
		MsgOpt:  common.OptGet,
		MsgRes:  constants.LogDumpUrlPrefix + constants.ResTask,
		Handler: handlers.NewQueryProgressHandler(),
	})
}
