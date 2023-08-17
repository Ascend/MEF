// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/constants"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"
)

// CleanTempFiles clean temp files
func CleanTempFiles() (bool, error) {
	dirs := []string{constants.LogDumpTempDir, constants.LogDumpPublicDir}
	for _, dir := range dirs {
		if !fileutils.IsLexist(dir) {
			return false, nil
		}
		_, entries, err := fileutils.ReadDir(dir,
			fileutils.NewFileModeChecker(false, common.Umask077, false, false),
			fileutils.NewFileOwnerChecker(false, true, 0, 0))
		if err != nil {
			return false, fmt.Errorf("failed to read dir %s, %v", dir, err)
		}
		for _, entry := range entries {
			if err := fileutils.DeleteFile(filepath.Join(dir, entry.Name())); err != nil {
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
	if err := ctx.UpdateStatus(errStatus); err != nil {
		hwlog.RunLog.Errorf("failed to update task status, %v", err)
	}
	hwlog.RunLog.Errorf("task failed, %v", err)
}
