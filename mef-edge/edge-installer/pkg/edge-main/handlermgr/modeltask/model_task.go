// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package modeltask

import (
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

// BaseTaskStatus super class of a task status
type BaseTaskStatus struct {
	Task *ModelFileTask
}

func (b *BaseTaskStatus) checkAvailableSpace() bool {
	if b.Task.Size > maxModelFileSize {
		return false
	}
	osFree, err := envutils.GetDiskFree(constants.ModelFileRootPath)
	if err != nil {
		return false
	}
	if b.Task.Size > int(osFree)-b.Task.RemainSize-reserveModelFileSize {
		return false
	}
	return true
}

func (m *ModelFileTask) start() error {
	return m.CurrentStatus.start()
}

func (m *ModelFileTask) setCurrentStatus(status StatusIntf) {
	hwlog.RunLog.Infof("%s_%s_%s change to status : %s",
		m.Uuid, m.ModelFile.Name, m.ModelFile.Version, status.getStatusType())
	m.CurrentStatus = status
	GetModelReporter().Notify()
}

func (m *ModelFileTask) buildReportData() *DownloadInfo {
	return m.CurrentStatus.buildReportData(*m)
}

func (m *ModelFileTask) buildRecord() *ModelDBRecord {
	return m.CurrentStatus.buildRecord(*m)
}

func (m *ModelFileTask) onEvent(event IModelEvent) {
	m.CurrentStatus.onEvent(event)
}

func (m *ModelFileTask) cancel() {
	m.CurrentStatus.cancel()
}

// GetStatusType get the status type of task
func (m *ModelFileTask) GetStatusType() types.ModelStatusType {
	return m.CurrentStatus.getStatusType()
}

func (m *ModelFileTask) canUpgrade(mo types.ModelFile) bool {
	return m.CurrentStatus.canUpgrade(mo)
}

func (m *ModelFileTask) buildBrief() types.ModelBrief {
	brief := types.ModelBrief{
		Uuid:   m.Uuid,
		Name:   m.ModelFile.Name,
		Status: m.CurrentStatus.getStatusType().String(),
	}
	return brief
}
