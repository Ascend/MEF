// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package modeltask

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/edge-main/common"
)

const percentMultiple = 100
const percentBase = 10

// DownloadingStatus the downloading status of a task
type DownloadingStatus struct {
	*BaseTaskStatus
	cancelFunc context.CancelFunc
	ca         []byte
}

// NewDownloadingStatus create a new downloading status
func NewDownloadingStatus(task *ModelFileTask, ca []byte) *DownloadingStatus {
	status := &DownloadingStatus{
		BaseTaskStatus: &BaseTaskStatus{
			Task: task,
		},
		ca: ca,
	}
	return status
}

func (t *DownloadingStatus) start() error {
	if !t.checkAvailableSpace() {
		hwlog.RunLog.Error("not enough space to download model file, please clear first")
		return fmt.Errorf("not enough space to download model file, please clear first")
	}
	err := fileutils.CreateDir(filepath.Join(constants.ModeFileDownloadDir, t.Task.Uuid), fileutils.Mode700)
	if err != nil {
		hwlog.RunLog.Errorf("create dir %s fail", filepath.Join(constants.ModeFileDownloadDir, t.Task.Uuid))
		return err
	}
	saveDir := filepath.Join(constants.ModeFileDownloadDir, t.Task.Uuid, t.Task.ModelFile.Name)
	ctx, cancelFunc := context.WithCancel(context.Background())
	t.cancelFunc = cancelFunc
	downloader := NewUrlDownloader(t.Task.Uuid, saveDir, t.Task.Size, t.Task.ModelFile, ctx)
	downloader.setCa(t.ca)
	go downloader.download()
	return nil
}

// GetStatusType get the actual status type
func (t *DownloadingStatus) getStatusType() types.ModelStatusType {
	return types.StatusDownloading
}

// BuildReportData build the data send to remote
func (t *DownloadingStatus) buildReportData(m ModelFileTask) *DownloadInfo {
	percent := strconv.FormatInt(int64((m.DownloadedSize*percentMultiple)/m.Size), percentBase) + "%"
	downInfo := DownloadInfo{
		Name:           m.ModelFile.Name,
		Version:        m.ModelFile.Version,
		Reason:         "",
		DownloadedTime: "",
		Status:         "inactive",
		DownloadStatus: "downloading",
		Percentage:     percent,
	}
	return &downInfo
}

func (t *DownloadingStatus) buildRecord(m ModelFileTask) *ModelDBRecord {
	return nil
}

// OnEvent handle the outside event
func (t *DownloadingStatus) onEvent(event IModelEvent) {
	if event.GetEventType() == TypeProgress {
		if pEvent, ok := event.(ProgressEvent); ok {
			t.Task.DownloadedSize = pEvent.Progress
			return
		}
		hwlog.RunLog.Errorf("event data not right: %v", event.GetEventType())
		return
	} else if event.GetEventType() == TypeDownloadFinish {
		if t.cancelFunc != nil {
			t.cancelFunc()
		}
		pEvent, ok := event.(DownloadFinishEvent)
		if !ok {
			hwlog.RunLog.Errorf("event data not right: %v", event.GetEventType())
			return
		}
		fdIp, err := common.GetFdIp()
		if err != nil {
			hwlog.RunLog.Warnf("get fd ip failed: %s", err.Error())
		}
		if pEvent.Success {
			hwlog.OpLog.Infof("[%s@%s] %s %s %s, the message is forwarded from [%s:%s]", constants.DeviceOmModule,
				constants.LocalIp, "download", constants.ResourceTypeModelFile, constants.Success, constants.FD, fdIp)
			t.Task.DownloadedTime = time.Now().Format(time.RFC3339)
			inactiveStatus := BuildModelStatus(&InactiveStatus{}, t.Task)
			t.Task.setCurrentStatus(inactiveStatus)
		} else {
			hwlog.OpLog.Errorf("[%s@%s] %s %s %s, the message is forwarded from [%s:%s]", constants.DeviceOmModule,
				constants.LocalIp, "download", constants.ResourceTypeModelFile, constants.Failed, constants.FD, fdIp)
			failStatus := buildFailStatus(t.Task, pEvent.Reason)
			t.Task.setCurrentStatus(failStatus)
			if err := t.Task.start(); err != nil {
				// currently unreachable branch. The start method of failedStatus always succeeds.
				hwlog.RunLog.Errorf("failed to start failed status, %v", err)
			}
		}
		return
	}
	hwlog.RunLog.Errorf("task in %s status, cannot process event: %v", t.getStatusType(), event.GetEventType())
}

// Cancel cancel the task
func (t *DownloadingStatus) cancel() {
	hwlog.RunLog.Errorf("cancel task %s_%s", t.Task.Uuid, t.Task.ModelFile.Name)
	if t.cancelFunc != nil {
		t.cancelFunc()
	}
}

func (t *DownloadingStatus) canUpgrade(m types.ModelFile) bool {
	if t.Task.ModelFile.Compare(m) {
		hwlog.RunLog.Infof("has same %s status task, cannot update", t.getStatusType())
		return false
	}
	return true
}

func (t *DownloadingStatus) setBaseStatus(b *BaseTaskStatus) {
	t.BaseTaskStatus = b
}
