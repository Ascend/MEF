// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package msgchecker is a public util for validation if a Pod or Pod patch message meets requirement of
package msgchecker

import (
	"errors"
	"fmt"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

const (
	modelFilePrefix = "/var/lib/docker/modelfile/"
	modelFileDir    = "/var/lib/docker/modelfile"
)

type podCheckerIntf interface {
	check(podInfo *types.Pod) error
}

type podChecker struct {
	operation string
	isPatch   bool
}

func (mv *MsgValidator) getPodChecker(isPatch bool) podCheckerIntf {
	var checker podCheckerIntf

	c := podChecker{
		operation: mv.operation,
		isPatch:   isPatch,
	}

	if mv.netType == constants.FDWithOM {
		checker = &fdPodChecker{
			podChecker: c,
		}
	} else {
		checker = &mefPodChecker{
			podChecker: c,
		}
	}

	return checker
}

func (mv *MsgValidator) auxCheckPod(podInfo *types.Pod) error {
	if err := validateStruct(podInfo); err != nil {
		return err
	}
	checker := mv.getPodChecker(false)
	if err := checker.check(podInfo); err != nil {
		return fmt.Errorf("check pod error, %v", err)
	}

	return nil
}

// auxCheckPodPatch [method] provides a function to be invoked externally for this package.
func (mv *MsgValidator) auxCheckPodPatch(patch *types.PodPatch) error {
	if err := validateStruct(&patch.Object); err != nil {
		return err
	}

	checker := mv.getPodChecker(true)
	if err := checker.check(&patch.Object); err != nil {
		return fmt.Errorf("check pod patch error, %v", err)
	}

	return nil
}

func (pc *podChecker) checkPodResources(pod *types.Pod) error {
	excludePods := utils.NewSet(pod.Name)
	podResourcesNeeds := make(map[v1.ResourceName]resource.Quantity)
	for _, c := range pod.Spec.Containers {
		for resName, request := range c.Resources.Req {
			requestOrLimit := request
			// 在资源limit和request中选择较大的一个
			if request.Cmp(c.Resources.Lim[resName]) < 0 {
				requestOrLimit = c.Resources.Lim[resName]
			}
			totalNeed, ok := podResourcesNeeds[resName]
			if !ok {
				totalNeed = resource.Quantity{}
			}
			totalNeed.Add(requestOrLimit)
			podResourcesNeeds[resName] = totalNeed
		}
	}
	availableRes, err := statusmanager.GetAvailableRes(excludePods)
	if err != nil {
		return fmt.Errorf("get available resource of node failed, %v", err)
	}
	for name, quantity := range podResourcesNeeds {
		if name == constants.CenterNpuName && quantity.Sign() > 0 {
			npuName, ok := common.LoadNpuFromDb()
			if ok {
				name = v1.ResourceName(npuName)
			}
		}
		available, ok := availableRes[name]
		if !ok {
			available = resource.Quantity{}
		}
		if available.Cmp(quantity) < 0 {
			return fmt.Errorf("check pod resources failed, [%s] is out of the node's available quantities",
				string(name))
		}
	}
	return nil
}

func getPodHostPath(volumes []types.Volume) []string {
	var hostPaths []string
	for _, v := range volumes {
		if v.VolumeSource.HostPath == nil {
			continue
		}
		hostPaths = append(hostPaths, v.VolumeSource.HostPath.Path)
	}
	return hostPaths
}

func (pc *podChecker) checkHostNetwork(podInfo *types.Pod) error {
	if configpara.GetPodConfig().UseHostNetwork == false && podInfo.Spec.HostNetwork {
		return errors.New("cur config not support pod host network")
	}
	return nil
}

func (pc *podChecker) checkHostPid(podInfo *types.Pod) error {
	if configpara.GetPodConfig().HostPid == false && podInfo.Spec.HostPID {
		return errors.New("cur config not support pod host pid")
	}
	return nil
}

func (pc *podChecker) checkPodPorts(podInfo *types.Pod) error {
	tcpPorts := map[int32]struct{}{}
	udpPorts := map[int32]struct{}{}
	for _, c := range podInfo.Spec.Containers {
		ports := map[int32]struct{}{}
		for _, port := range c.Ports {
			var exist bool
			switch port.Protocol {
			case "TCP":
				_, exist = tcpPorts[port.HostPort]
				tcpPorts[port.HostPort] = struct{}{}
			case "UDP":
				_, exist = udpPorts[port.HostPort]
				udpPorts[port.HostPort] = struct{}{}
			default:
				return errors.New("unsupported protocol")
			}
			if exist {
				return errors.New("duplicated host port")
			}

			if _, ok := ports[port.ContainerPort]; ok {
				return errors.New("duplicated port of container")
			}
			ports[port.ContainerPort] = struct{}{}

		}
	}
	return nil
}

func isContainerNameChanged(oldPod, newPod *types.Pod) bool {
	var setA = utils.NewSet()
	var setB = utils.NewSet()
	for idx := range oldPod.Spec.Containers {
		setA.Add(oldPod.Spec.Containers[idx].Name)
	}

	for idx := range newPod.Spec.Containers {
		setB.Add(newPod.Spec.Containers[idx].Name)
	}

	if len(setA.List()) != len(setB.List()) {
		return true
	}

	if len(setA.Difference(setB).List()) != 0 {
		return true
	}

	return false
}

func (pc *podChecker) checkPodVolumeNameDuplicate(podInfo *types.Pod) error {
	var volumeNames = map[string]struct{}{}
	for _, v := range podInfo.Spec.Volumes {
		if _, ok := volumeNames[v.Name]; ok {
			return fmt.Errorf("volume name is not unique")
		}
		volumeNames[v.Name] = struct{}{}
	}

	return nil
}
