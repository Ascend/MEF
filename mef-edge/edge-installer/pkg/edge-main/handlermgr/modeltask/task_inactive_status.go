// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package modeltask

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/types"
)

// InactiveStatus the inactive status of a task.download success
type InactiveStatus struct {
	*BaseTaskStatus
}

func (t *InactiveStatus) start() error {
	return nil
}

// GetStatusType get the actual status type
func (t *InactiveStatus) getStatusType() types.ModelStatusType {
	return types.StatusInactive
}

// BuildReportData build the data send to remote
func (t *InactiveStatus) buildReportData(m ModelFileTask) *DownloadInfo {
	downInfo := DownloadInfo{
		Name:           m.ModelFile.Name,
		Version:        m.ModelFile.Version,
		Reason:         "",
		DownloadedTime: m.DownloadedTime,
		Status:         "inactive",
		DownloadStatus: "downloaded",
		Percentage:     "100%",
	}
	return &downInfo
}

func (t *InactiveStatus) buildRecord(m ModelFileTask) *ModelDBRecord {
	record := ModelDBRecord{
		Uuid:           m.Uuid,
		Name:           m.ModelFile.Name,
		Version:        m.ModelFile.Version,
		DownloadedTime: m.DownloadedTime,
		Status:         t.getStatusType().String(),
		CheckType:      m.ModelFile.CheckType,
		CheckCode:      m.ModelFile.CheckCode,
	}
	return &record
}

// OnEvent handle the outside event
func (t *InactiveStatus) onEvent(event IModelEvent) {
	hwlog.RunLog.Errorf("task in %s status, cannot process event: %v", t.getStatusType(), event.GetEventType())
}

// Cancel cancel the task
func (t *InactiveStatus) cancel() {
	// do nothing stub
}

func (t *InactiveStatus) canUpgrade(m types.ModelFile) bool {
	if t.Task.ModelFile.Compare(m) {
		hwlog.RunLog.Infof("has same %s status task, cannot update", t.getStatusType())
		return false
	}
	return true
}

func (t *InactiveStatus) setBaseStatus(b *BaseTaskStatus) {
	t.BaseTaskStatus = b
}
