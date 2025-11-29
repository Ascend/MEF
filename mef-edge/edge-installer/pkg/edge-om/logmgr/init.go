// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logmgr contains implements log management functions.
package logmgr

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/logmgmt/logrotate"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/path"
)

type logManager struct {
	name    string
	ctx     context.Context
	rotator *logrotate.LogRotator
	syncer  *logSyncer
}

// NewLogMgr creates a log manager
func NewLogMgr(name string, ctx context.Context) model.Module {
	return &logManager{
		name: name,
		ctx:  ctx,
	}
}

// Name returns module's name
func (l *logManager) Name() string {
	return l.name
}

// Enable enable log manager
func (l *logManager) Enable() bool {
	logPathMgr, err := path.GetLogPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("unable to enable module %s, %s", l.Name(), err.Error())
		return false
	}
	l.rotator = newLogRotator(logPathMgr)
	l.syncer = newLogSyncer()
	return true
}

// Start starts module
func (l *logManager) Start() {
	go l.syncer.start(l.ctx)
	l.rotator.Start(l.ctx)
}
