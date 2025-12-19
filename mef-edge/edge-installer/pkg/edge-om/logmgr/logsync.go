// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
