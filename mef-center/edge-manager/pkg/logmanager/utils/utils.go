// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils
package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
)

func deleteTempSubFiles(dir string, entries []os.DirEntry) error {
	if len(entries) > common.MaxLoopNum {
		return errors.New("the number of dir entries exceed the upper limit")
	}
	for _, entry := range entries {
		if err := fileutils.DeleteFile(filepath.Join(dir, entry.Name())); err != nil {
			return fmt.Errorf("failed to delete file %s, %v", entry.Name(), err)
		}
	}
	return nil
}

// CleanTempFiles clean temp files
func CleanTempFiles() (bool, error) {
	dirs := []string{constants.LogDumpTempDir, constants.LogDumpPublicDir}
	for _, dir := range dirs {
		if !fileutils.IsLexist(dir) {
			return false, nil
		}
		fileHandle, entries, err := fileutils.ReadDir(dir,
			fileutils.NewFileLinkChecker(false),
			fileutils.NewFileModeChecker(true, fileutils.DefaultWriteFileMode, false, false),
			fileutils.NewFileOwnerChecker(true, true, 0, 0))
		if err != nil {
			return false, fmt.Errorf("failed to read dir %s, %v", dir, err)
		}
		if err = deleteTempSubFiles(dir, entries); err != nil {
			if closeErr := fileHandle.Close(); closeErr != nil {
				hwlog.RunLog.Errorf("closer dir %s's handle failed: %v", dir, closeErr)
			}
			return false, err
		}

		if err = fileHandle.Close(); err != nil {
			hwlog.RunLog.Errorf("closer dir %s's handle failed: %v", dir, err)
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
	if err := ctx.Spec().Args.Get(constants.NodeSnAndIp, &serialNumber); err != nil {
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
