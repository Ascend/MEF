// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package modeltask

import (
	"path/filepath"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

// FailStatus the fail status of a task
type FailStatus struct {
	*BaseTaskStatus
	reason string
}

func (t *FailStatus) start() error {
	filePath := filepath.Join(constants.ModeFileDownloadDir, t.Task.Uuid, t.Task.ModelFile.Name)
	if err := fileutils.DeleteAllFileWithConfusion(filePath); err != nil {
		hwlog.RunLog.Warnf("remove model file [%s] failed", filePath)
	}
	return nil
}

// GetStatusType get the actual status type
func (t *FailStatus) getStatusType() types.ModelStatusType {
	return types.StatusFail
}

// BuildReportData build the data send to remote
func (t *FailStatus) buildReportData(m ModelFileTask) *DownloadInfo {
	t.Task.ReportCount++
	downInfo := DownloadInfo{
		Name:           m.ModelFile.Name,
		Version:        m.ModelFile.Version,
		Reason:         t.reason,
		DownloadedTime: "",
		Status:         "inactive",
		DownloadStatus: "failed",
		Percentage:     "0%",
	}
	return &downInfo
}

func (t *FailStatus) buildRecord(m ModelFileTask) *ModelDBRecord {
	return nil
}

// OnEvent handle the outside event
func (t *FailStatus) onEvent(event IModelEvent) {
	hwlog.RunLog.Errorf("task in %s status, cannot process event: %v", t.getStatusType(), event.GetEventType())
}

// Cancel cancel the task
func (t *FailStatus) cancel() {
	// do nothing stub
}

func (t *FailStatus) canUpgrade(m types.ModelFile) bool {
	return true
}

func (t *FailStatus) setBaseStatus(b *BaseTaskStatus) {
	t.BaseTaskStatus = b
}
