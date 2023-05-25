// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logmanager enables collecting logs
package logmanager

import (
	"context"
	"os"
	"strconv"
	"sync"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/handlerbase"

	"edge-manager/pkg/logmanager/handlers"
	"edge-manager/pkg/logmanager/modules"
)

// NewLogManager creates a new log manager
func NewLogManager(ctx context.Context, enable bool) model.Module {
	return &logManager{
		enable:    enable,
		ctx:       ctx,
		taskMgr:   modules.NewTaskMgr(ctx),
		uploadMgr: modules.NewUploadMgr(ctx),
	}
}

type logManager struct {
	enable         bool
	nodeIp         string
	nodePort       int
	ctx            context.Context
	handlerMgr     handlerbase.HandlerMgr
	handlerMgrOnce sync.Once
	uploadMgr      modules.UploadMgr
	taskMgr        modules.TaskMgr
}

// Name returns name of module
func (l *logManager) Name() string {
	return common.LogManagerName
}

// Enable does preparation for module
func (l *logManager) Enable() bool {
	if l.enable {
		portStr := os.Getenv("LOG_MGR_NODE_PORT")
		port, err := strconv.Atoi(portStr)
		if err != nil {
			hwlog.RunLog.Error("failed to parse nodePort value")
			return !l.enable
		}
		l.nodePort = port
		l.nodeIp = os.Getenv("NODE_IP")
	}
	return l.enable
}

// Start starts module
func (l *logManager) Start() {

	l.taskMgr.Start()
	l.uploadMgr.Start()

	for {
		select {
		case _, ok := <-l.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		req, err := modulemgr.ReceiveMessage(l.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message from channel failed, error: %v", l.Name(), err)
			continue
		}
		go l.dispatchMsg(req)
	}
}

func (l *logManager) dispatchMsg(msg *model.Message) {
	if err := l.getHandlerMgr().Process(msg); err != nil {
		hwlog.RunLog.Errorf("failed to process msg, %v", err)
	}
}

func (l *logManager) getHandlerMgr() *handlerbase.HandlerMgr {
	l.handlerMgrOnce.Do(l.registerHandlers)
	return &l.handlerMgr
}

func (l *logManager) registerHandlers() {
	l.handlerMgr.Register(handlerbase.RegisterInfo{
		MsgOpt:  common.OptReport,
		MsgRes:  common.ResLogTaskProgressEdge,
		Handler: handlers.GetReportEdgeProgressHandler(l.taskMgr),
	})
	l.handlerMgr.Register(handlerbase.RegisterInfo{
		MsgOpt:  common.OptPost,
		MsgRes:  common.LogCollectPathPrefix + common.ResRelLogTask,
		Handler: handlers.GetCreateTaskHandler(l.taskMgr, l.nodeIp, l.nodePort),
	})
	l.handlerMgr.Register(handlerbase.RegisterInfo{
		MsgOpt:  common.OptGet,
		MsgRes:  common.LogCollectPathPrefix + common.ResRelLogTaskProgress,
		Handler: handlers.GetQueryTaskProgressHandler(l.taskMgr),
	})
	l.handlerMgr.Register(handlerbase.RegisterInfo{
		MsgOpt:  common.OptGet,
		MsgRes:  common.LogCollectPathPrefix + common.ResRelLogTaskPath,
		Handler: handlers.GetQueryTaskPathHandler(l.taskMgr),
	})
}
