// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr for deal every handler
package handlermgr

import (
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

const (
	containerIdIndex = 0
	podNamePrefix    = "k8s_POD_"
)

type restartPodHandler struct{}

// Handle entry
func (r *restartPodHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start restart pod")
	var podName string
	if err := msg.ParseContent(&podName); err != nil {
		r.sendResponse(msg, constants.Failed)
		hwlog.RunLog.Errorf("parse msg content failed: %v", err)
		return errors.New("parse")
	}

	if err := r.restartPod(podName); err != nil {
		r.sendResponse(msg, constants.Failed)
		hwlog.RunLog.Errorf("restart pod failed, error: %v", err)
		return errors.New("restart pod failed")
	}

	r.sendResponse(msg, constants.Success)
	hwlog.RunLog.Infof("restart pod [%s] success", podName)
	return nil
}

func (r *restartPodHandler) restartPod(podName string) error {
	cmdOut, err := envutils.RunCommand(constants.DockerCmd, envutils.DefCmdTimeoutSec, "ps")
	if err != nil {
		hwlog.RunLog.Errorf("get running container info failed, error: %v", err)
		return errors.New("get running container info failed")
	}

	var podId string
	namePrefix := podNamePrefix + podName
	containerInfos := strings.Split(cmdOut, "\n")
	iterationCount := 1
	for _, containerInfo := range containerInfos {
		if iterationCount > constants.MaxIterationCount {
			break
		}
		iterationCount++
		containerInfoSplit := strings.Fields(containerInfo)
		if len(containerInfoSplit) == 0 {
			continue
		}
		if strings.HasPrefix(containerInfoSplit[len(containerInfoSplit)-1], namePrefix) {
			podId = containerInfoSplit[containerIdIndex]
			break
		}
	}

	if podId == "" {
		return fmt.Errorf("didn't find running pod [%s]", podName)
	}

	if _, err = envutils.RunCommand(constants.DockerCmd, envutils.DefCmdTimeoutSec, "stop", podId); err != nil {
		hwlog.RunLog.Errorf("stop pod [%s] failed, error: %v", podName, err)
		return errors.New("stop pod failed")
	}
	return nil
}

func (r *restartPodHandler) sendResponse(msg *model.Message, respMsg string) {
	newResp, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("get new response message failed, error: %v", err)
		return
	}
	if err = newResp.FillContent(respMsg); err != nil {
		hwlog.RunLog.Errorf("fill resp into content failed: %v", err)
		return
	}
	if err = sendHandlerReplyMsg(newResp); err != nil {
		hwlog.RunLog.Errorf("send restart pod handler response failed, error: %v", err)
	}
}
