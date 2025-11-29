// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package monitors for log monitor
package monitors

import (
	"errors"
	"syscall"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/path"
)

const (
	logMonitorInterval    = 1 * time.Minute
	maxOccupiedSpaceRatio = 0.8
	logMonitorName        = "log"
)

var logTask = &cronTask{
	alarmId:         almutils.EdgeLogAbnormal,
	name:            logMonitorName,
	interval:        logMonitorInterval,
	checkStatusFunc: isDiskSpaceEnough,
}

func isDiskSpaceEnough() error {
	edgeLogDir, edgeLogBackupDir, err := path.GetEdgeLogDirs()
	if err != nil {
		hwlog.RunLog.Errorf("get edge log dirs failed, err:%s", err.Error())
		return err
	}
	checkPaths := []string{edgeLogDir, edgeLogBackupDir}
	for _, checkPath := range checkPaths {
		var fs syscall.Statfs_t
		if err := syscall.Statfs(checkPath, &fs); err != nil {
			hwlog.RunLog.Errorf("check whether disk space is enough failed, %v", err)
			return err
		}
		used := uint64(fs.Bsize) * (fs.Blocks - fs.Bfree)
		avail := uint64(fs.Bsize) * fs.Bavail
		if avail == 0 {
			hwlog.RunLog.Error("available space is zero")
			return errors.New("available space is zero")
		}
		occupiedSpaceRatio := float64(used) / float64(used+avail)
		if occupiedSpaceRatio > maxOccupiedSpaceRatio {
			return errors.New("disk space is not enough")
		}
	}
	return nil
}
