// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
)

// CleanTempFiles clean temp files
func CleanTempFiles() (bool, error) {
	dirs := []string{constants.LogDumpTempDir, constants.LogDumpPublicDir}
	for _, dir := range dirs {
		if !fileutils.IsLexist(dir) {
			return false, nil
		}
		fileHandle, entries, err := fileutils.ReadDir(dir,
			fileutils.NewFileModeChecker(false, common.Umask077, false, false),
			fileutils.NewFileOwnerChecker(false, true, 0, 0))
		if err != nil {
			return false, fmt.Errorf("failed to read dir %s, %v", dir, err)
		}
		defer func() {
			if fileHandle == nil {
				return
			}

			if err = fileHandle.Close(); err != nil {
				hwlog.RunLog.Errorf("closer dir %s's handle failed: %v", dir, err)
			}
		}()
		for _, entry := range entries {
			if err = fileutils.DeleteFile(filepath.Join(dir, entry.Name())); err != nil {
				return false, fmt.Errorf("failed to delete file %s, %v", entry.Name(), err)
			}
		}
	}
	return true, nil
}

// FeedbackTaskError feedback task error
func FeedbackTaskError(ctx taskschedule.TaskContext, err error) {
	errStatus := taskschedule.TaskStatus{
		Phase:   taskschedule.Failed,
		Message: err.Error(),
	}
	var serialNumber string
	if err := ctx.Spec().Args.Get(constants.NodeSerialNumber, &serialNumber); err != nil {
		serialNumber = ""
	}
	if err := ctx.UpdateStatus(errStatus); err != nil {
		if serialNumber != "" {
			hwlog.RunLog.Errorf("failed to update sub task status for edge(%s), %v", serialNumber, err)
		} else {
			hwlog.RunLog.Errorf("failed to update task status, %v", err)
		}
	}
	if serialNumber != "" {
		hwlog.RunLog.Errorf("sub task for edge(%s) failed, %v", serialNumber, err)
	} else {
		hwlog.RunLog.Errorf("task failed, %v", err)
	}
}
