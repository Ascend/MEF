// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logmgr
package logmgr

import (
	"context"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/util"
)

const (
	logSyncInterval = 10 * time.Minute
)

type logSyncer struct {
}

func newLogSyncer() *logSyncer {
	return &logSyncer{}
}

func (l logSyncer) start(ctx context.Context) {
	l.doSyncLog()
	tick := time.NewTicker(logSyncInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Warn("sync log stop")
			return
		case <-tick.C:
			l.doSyncLog()
		}
	}
}

func (l logSyncer) doSyncLog() {
	logSyncMgr := util.NewLogSyncMgr()
	if err := logSyncMgr.BackupLogs(); err != nil {
		hwlog.RunLog.Errorf("backup tmpfs log failed, %v", err)
		return
	}
	hwlog.RunLog.Infof("backup tmpfs log success")
}
