// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package statusmanager
package statusmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/common/database"
)

const (
	containerCreatingTimeout = 10 * time.Minute
	timeoutExitCode          = 65535
	unknownErrorExitCode     = 65534
)

// FdNode defines node struct for fd
type FdNode struct {
	Status FdNodeStatus `json:"Status"`
}

// FdNodeStatus defines node status struct for fd
type FdNodeStatus struct {
	Addresses   []v1.NodeAddress `json:"addresses"`
	Allocatable v1.ResourceList  `json:"allocatable"`
	Capacity    v1.ResourceList  `json:"capacity"`
	Content     v1.Node          `json:"content"`
}

// FdPod defines pod struct for fd
type FdPod struct {
	Name     string            `json:"Name"`
	UID      string            `json:"UID"`
	Kind     string            `json:"kind"`
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     v1.PodSpec        `json:"spec"`
	Status   v1.PodStatus      `json:"status"`
}

// LoadPodsDataForFd loads pods data for fd
func LoadPodsDataForFd() ([]FdPod, error) {
	allStatuses, err := GetPodStatusMgr().GetAll()
	if err != nil {
		return nil, err
	}

	edgeMaxPodNumber := configpara.GetPodConfig().MaxContainerNumber
	if len(allStatuses) > edgeMaxPodNumber {
		return nil, fmt.Errorf("pod num in db exceeds limit[%d]", edgeMaxPodNumber)
	}

	// fd only accepts a non-nil empty slice if there is no containers
	result := make([]FdPod, 0)
	for _, podStatus := range allStatuses {
		var pod v1.Pod
		if err := json.Unmarshal([]byte(podStatus), &pod); err != nil {
			return nil, err
		}
		result = append(result, *convertToFdPod(&pod))
	}
	return result, nil
}

// LoadNodeDataForFd loads node data for fd
func LoadNodeDataForFd() (*FdNode, error) {
	nodeStatuses, err := GetNodeStatusMgr().GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get node status: %v", err)
	}
	if len(nodeStatuses) == 0 {
		return nil, errors.New("failed to get node status, record not found")
	}
	var nodeStatus string
	for _, nodeStatus = range nodeStatuses {
		break
	}
	var node v1.Node
	if err := json.Unmarshal([]byte(nodeStatus), &node); err != nil {
		return nil, fmt.Errorf("failed to unmarshal node status: %v", err)
	}

	return convertToFdNode(&node), nil
}

// DeleteNodeStatus delete node status in database
func DeleteNodeStatus() error {
	nodeStatuses, err := GetNodeStatusMgr().GetAll()
	if err != nil {
		return fmt.Errorf("failed to get node status: %v", err)
	}
	for resourceName := range nodeStatuses {
		if err = GetNodeStatusMgr().Delete(resourceName); err != nil {
			return fmt.Errorf("failed to delete node status: %v", err)
		}
	}
	return nil
}

// GetAvailableRes [method] for getting all available resources of node expect those used by pods
func GetAvailableRes(excludePods *utils.Set) (v1.ResourceList, error) {
	metas, err := database.GetMetaRepository().GetByType(constants.ResourceTypeNode)
	if err != nil || len(metas) != 1 {
		return nil, errors.New("get node failed, db error")
	}
	nodeMeta := metas[0]
	var node v1.Node
	if err = json.Unmarshal([]byte(nodeMeta.Value), &node); err != nil {
		return nil, errors.New("get node failed, unmarshal node error")
	}
	available := node.Status.Allocatable

	totalAllocated, err := getTotalAllocatedRes(excludePods)
	if err != nil {
		return nil, fmt.Errorf("get total allocated resources failed, %s", err.Error())
	}
	for resName, quantity := range totalAllocated {
		availableRes := available[resName]
		if availableRes.Cmp(quantity) < 0 {
			return nil, errors.New("get node available resource error")
		}
		availableRes.Sub(quantity)
		availableRes = config.GetScaledNpu(resName, availableRes)
		available[resName] = availableRes
	}
	return available, nil
}

func getTotalAllocatedRes(excludePods *utils.Set) (v1.ResourceList, error) {
	totalAllocated := make(map[v1.ResourceName]resource.Quantity)
	podMaps, err := GetPodStatusMgr().GetAll()
	if err == ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.New("get pods failed, db error")
	}
	for _, podString := range podMaps {
		var pod v1.Pod
		if err := json.Unmarshal([]byte(podString), &pod); err != nil {
			return nil, errors.New("unmarshal pod error")
		}
		if excludePods.Find(pod.Name) {
			continue
		}
		for _, container := range pod.Spec.Containers {
			for resName, quantity := range container.Resources.Limits {
				allocatedRes := totalAllocated[resName]
				quantity.Add(allocatedRes)
				totalAllocated[resName] = quantity
			}
		}
	}
	return totalAllocated, nil
}

func convertToFdNode(node *v1.Node) *FdNode {
	var fdNode FdNode
	fdNode.Status.Addresses = node.Status.Addresses
	fdNode.Status.Capacity = convertToFdResourceList(node.Status.Capacity)
	fdNode.Status.Allocatable = convertToFdResourceList(node.Status.Allocatable)
	fdNode.Status.Content = *node
	return &fdNode
}

func convertToFdPod(pod *v1.Pod) *FdPod {
	terminatedContainerStates := fixContainerStateTerminated(pod)
	if len(terminatedContainerStates) > 0 {
		updatePodStatus(pod, terminatedContainerStates[0])
	}
	var fdPod FdPod
	fdPod.Name = filepath.Base(pod.ObjectMeta.Name)
	fdPod.UID = string(pod.ObjectMeta.UID)
	fdPod.Kind = "Pod"
	fdPod.Metadata = pod.ObjectMeta
	fdPod.Spec = pod.Spec
	fdPod.Status = pod.Status
	for _, container := range fdPod.Spec.Containers {
		config.ModifyNpuRes(container.Resources.Limits, true)
		config.ModifyNpuRes(container.Resources.Requests, true)
	}
	return &fdPod
}

func convertToFdResourceList(rawResList v1.ResourceList) v1.ResourceList {
	var (
		npuName  = v1.ResourceName(constants.SharableNpuName)
		npuValue resource.Quantity
	)
	for name, value := range rawResList {
		if util.IsWholeNpu(string(name)) {
			npuName = name
			npuValue = config.GetScaledNpu(name, value)
			break
		}
	}

	convertedResList := map[v1.ResourceName]resource.Quantity{
		npuName: npuValue,
	}
	resNames := []v1.ResourceName{v1.ResourceCPU, v1.ResourceMemory, v1.ResourcePods}
	for _, name := range resNames {
		quantity, ok := rawResList[name]
		if !ok {
			quantity = resource.Quantity{}
		}
		convertedResList[name] = quantity
	}
	return convertedResList
}

func fixContainerStateTerminated(pod *v1.Pod) []*v1.ContainerStateTerminated {
	var terminatedContainerStates []*v1.ContainerStateTerminated
	for idx := range pod.Status.ContainerStatuses {
		status := &pod.Status.ContainerStatuses[idx]
		if status.State.Terminated != nil {
			if status.State.Terminated.ExitCode != 0 {
				hwlog.RunLog.Errorf("container terminate, pod=%s, container=%s, reason=%s",
					pod.Name, status.Name, status.State.Terminated.Reason)
				terminatedContainerStates = append(terminatedContainerStates, status.State.Terminated)
			}
		} else if status.LastTerminationState.Terminated != nil && status.State.Waiting != nil {
			if status.LastTerminationState.Terminated.ExitCode != 0 {
				hwlog.RunLog.Errorf("container terminate and restart, pod=%s, container=%s, reason=%s",
					pod.Name, status.Name, status.LastTerminationState.Terminated.Reason)
				status.State = status.LastTerminationState
				terminatedContainerStates = append(terminatedContainerStates, status.LastTerminationState.Terminated)
			}
		} else if status.State.Waiting != nil && status.RestartCount == 0 &&
			pod.Status.StartTime != nil && time.Now().Sub(pod.Status.StartTime.Time) > containerCreatingTimeout {
			hwlog.RunLog.Errorf("container create timeout, pod=%s, container=%s, reason=%s",
				pod.Name, status.Name, status.State.Waiting.Reason)
			status.State.Terminated = &v1.ContainerStateTerminated{
				Reason:   fmt.Sprintf("ContainerCreatingTimeout(%s)", status.State.Waiting.Reason),
				Message:  fmt.Sprintf("timeout to create container: %s", status.State.Waiting.Message),
				ExitCode: timeoutExitCode,
			}
			status.State.Waiting = nil
			terminatedContainerStates = append(terminatedContainerStates, status.State.Terminated)
		} else if status.State.Waiting != nil && status.RestartCount == 0 &&
			strings.Contains(strings.ToLower(status.State.Waiting.Reason), "error") {
			hwlog.RunLog.Errorf("container create unexpected error, pod=%s, container=%s, reason=%s",
				pod.Name, status.Name, status.State.Waiting.Reason)
			status.State.Terminated = &v1.ContainerStateTerminated{
				Reason:   status.State.Waiting.Reason,
				Message:  status.State.Waiting.Message,
				ExitCode: unknownErrorExitCode,
			}
			status.State.Waiting = nil
			terminatedContainerStates = append(terminatedContainerStates, status.State.Terminated)
		}
	}
	return terminatedContainerStates
}

func updatePodStatus(pod *v1.Pod, state *v1.ContainerStateTerminated) {
	if pod.Status.Phase == v1.PodPending || pod.Status.Phase == v1.PodRunning {
		pod.Status.Phase = v1.PodFailed
		pod.Status.Reason = state.Reason
		pod.Status.Message = state.Message
	}
}
