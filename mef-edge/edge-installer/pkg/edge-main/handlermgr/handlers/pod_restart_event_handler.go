// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers
package handlers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

type podRestartEventHandler struct {
}

// Handle handles pod-restart events
func (h podRestartEventHandler) Handle(message *model.Message) error {
	err := h.handle(message)
	response := constants.OK
	if err != nil {
		response = constants.Failed
	}
	if err := util.SendInnerMsgResponse(message, response); err != nil {
		hwlog.RunLog.Errorf("failed to send sync response, %v", err)
	}
	return err
}

func (h podRestartEventHandler) handle(message *model.Message) error {
	var (
		podName             = filepath.Base(message.KubeEdgeRouter.Resource)
		restartedContainers []string
	)
	oldPodStr, err := statusmanager.GetPodStatusMgr().Get(
		strings.Replace(message.KubeEdgeRouter.Resource, "/podpatch/", "/pod/", 1))
	if err != nil {
		return err
	}
	var oldPod v1.Pod
	if err := json.Unmarshal([]byte(oldPodStr), &oldPod); err != nil {
		return err
	}
	mergedBytes, err := util.MergePatch([]byte(oldPodStr), model.UnformatMsg(message.Content))
	if err != nil {
		return err
	}
	var newPod v1.Pod
	if err := json.Unmarshal(mergedBytes, &newPod); err != nil {
		return err
	}

	for _, oldStatus := range oldPod.Status.ContainerStatuses {
		for _, newStatus := range newPod.Status.ContainerStatuses {
			if oldStatus.Name == newStatus.Name && hasContainerRestart(oldStatus, newStatus) {
				podName = oldPod.Name
				restartedContainers = append(restartedContainers, oldStatus.Name)
				hwlog.RunLog.Infof("container %s(%s) restart, send event to cloud", oldPod.Name, oldStatus.Name)
			}
		}
	}
	if len(restartedContainers) == 0 {
		return nil
	}
	alm, err := almutils.CreateAlarm(almutils.ApplicationRestart, podName, almutils.NotifyTypeEvent)
	if err != nil {
		hwlog.RunLog.Error("failed to create application restart event")
		return nil
	}
	alm.DetailedInformation += fmt.Sprintf("\nRestarted pod=%s, containers=[%s]",
		podName, strings.Join(restartedContainers, ","))
	if err := almutils.SendAlarm(constants.EdgedModule, constants.AlarmManager, alm); err != nil {
		hwlog.RunLog.Error("failed to send application restart event to alarm manager")
	}
	return nil
}

func hasContainerRestart(oldStatus, newStatus v1.ContainerStatus) bool {
	const (
		containerIDSep = "://"
		substrIndex    = 1
		substrCount    = 2
	)
	if oldStatus.ContainerID == newStatus.ContainerID && oldStatus.RestartCount == newStatus.RestartCount {
		return false
	}
	ids := strings.Split(oldStatus.ContainerID, containerIDSep)
	if len(ids) != substrCount {
		return false
	}
	return ids[substrIndex] != ""
}
