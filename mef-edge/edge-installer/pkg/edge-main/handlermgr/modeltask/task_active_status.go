// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package modeltask

import (
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/types"
)

// ActiveStatus the active status of a task
type ActiveStatus struct {
	*BaseTaskStatus
}

func (t *ActiveStatus) start() error {
	return nil
}

// GetStatusType get the actual status type
func (t *ActiveStatus) getStatusType() types.ModelStatusType {
	return types.StatusActive
}

// BuildReportData build the data send to remote
func (t *ActiveStatus) buildReportData(m ModelFileTask) *DownloadInfo {
	downInfo := DownloadInfo{
		Name:           m.ModelFile.Name,
		Version:        m.ModelFile.Version,
		Reason:         "",
		DownloadedTime: m.DownloadedTime,
		Status:         "active",
		DownloadStatus: "downloaded",
		Percentage:     "100%",
	}
	return &downInfo
}

func (t *ActiveStatus) buildRecord(m ModelFileTask) *ModelDBRecord {
	record := ModelDBRecord{
		PodUid:         m.PodUid,
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

func (t *ActiveStatus) onEvent(event IModelEvent) {
	hwlog.RunLog.Errorf("task in %s status, cannot process event: %v", t.getStatusType(), event.GetEventType())
}

func (t *ActiveStatus) cancel() {
	// do nothing stub
}

func (t *ActiveStatus) canUpgrade(m types.ModelFile) bool {
	if t.Task.ModelFile.Compare(m) {
		hwlog.RunLog.Infof("has same %s status task, cannot update", t.getStatusType())
		return false
	}
	return true
}

func (t *ActiveStatus) setBaseStatus(b *BaseTaskStatus) {
	t.BaseTaskStatus = b
}
