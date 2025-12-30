// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package msgconv
package msgconv

import (
	"encoding/json"
	"fmt"
	"strconv"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

func setNpuForUpstreamNode(node *v1.Node) error {
	setNpuForUpstreamResourceLists(node.Status.Allocatable, node.Status.Capacity)
	return nil
}

func setNpuForUpstreamNodePatch(patch *string) error {
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(*patch), &v); err != nil {
		return err
	}

	wrapper := util.NewWrapper(v)
	allocatable := wrapper.GetObject("status").GetObject("allocatable").GetData()
	capacity := wrapper.GetObject("status").GetObject("capacity").GetData()
	if err := setNpuForUpstreamResourceMaps(allocatable, capacity); err != nil {
		return err
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	*patch = string(data)
	return nil
}

// setNpuForDownstreamNode replaces resourceList by local data to ensure that the full resources are reported
func setNpuForDownstreamNode(node *v1.Node) error {
	localNodeStr, err := statusmanager.GetNodeStatusMgr().Get(constants.ActionDefaultNodeStatus + node.Name)
	if err != nil {
		return err
	}
	var localNode v1.Node
	if err := json.Unmarshal([]byte(localNodeStr), &localNode); err != nil {
		return err
	}
	node.Status.Capacity = localNode.Status.Capacity
	node.Status.Allocatable = localNode.Status.Allocatable
	return nil
}

func setNpuForUpstreamPodPatch(patch *string) error {
	var v map[string]interface{}
	if err := json.Unmarshal([]byte(*patch), &v); err != nil {
		return err
	}

	var resourceMaps []interface{}
	wrapper := util.NewWrapper(v)
	for _, w := range wrapper.GetObject("spec").GetSlice("containers") {
		resourceMaps = append(resourceMaps, w.GetObject("resources").GetObject("limits").GetData())
		resourceMaps = append(resourceMaps, w.GetObject("resources").GetObject("requests").GetData())
	}
	if err := setNpuForUpstreamResourceMaps(resourceMaps...); err != nil {
		return err
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	*patch = string(data)
	return nil
}

func setNpuForDownstreamPod(pod *v1.Pod) error {
	var resourceLists []v1.ResourceList
	for _, c := range pod.Spec.Containers {
		resourceLists = append(resourceLists, c.Resources.Limits)
		resourceLists = append(resourceLists, c.Resources.Requests)
	}
	setNpuForDownstreamResourceLists(resourceLists...)
	return nil
}

func setNpuForUpstreamResourceMaps(resourceMaps ...interface{}) error {
	npuRealName, ok := common.LoadNpuFromDb()
	if !ok {
		npuRealName = ""
	}
	for _, obj := range resourceMaps {
		if obj == nil {
			continue
		}
		resourceMap, ok := obj.(map[string]interface{})
		if !ok {
			return fmt.Errorf("resource list's type(%T) is invalid", obj)
		}
		resourceList := v1.ResourceList{}
		for k, v := range resourceMap {
			quantity, ok := parseQuantity(v)
			if !ok {
				continue
			}
			resourceList[v1.ResourceName(k)] = quantity
			delete(resourceMap, k)
		}
		setNpuForUpstreamResourceList(v1.ResourceName(npuRealName), resourceList)
		for k, v := range resourceList {
			resourceMap[string(k)] = v
		}
	}
	return nil
}

func setNpuForUpstreamResourceLists(resourceLists ...v1.ResourceList) {
	npuRealName, ok := common.LoadNpuFromDb()
	if !ok {
		npuRealName = ""
	}
	for _, list := range resourceLists {
		setNpuForUpstreamResourceList(v1.ResourceName(npuRealName), list)
	}
}

func setNpuForUpstreamResourceList(npuRealName v1.ResourceName, resourceList v1.ResourceList) {
	if resourceList == nil {
		return
	}
	config.ModifyNpuRes(resourceList, true)
	if npuRealName == constants.CenterNpuName {
		return
	}
	oldNpuQuantity, ok := resourceList[npuRealName]
	if !ok {
		return
	}
	resourceList[constants.CenterNpuName] = oldNpuQuantity
	delete(resourceList, npuRealName)
}

func setNpuForDownstreamResourceLists(resourceLists ...v1.ResourceList) {
	npuName, foundNpu := common.LoadNpuFromDb()
	for _, resourceList := range resourceLists {
		if resourceList == nil {
			continue
		}
		config.ModifyNpuRes(resourceList, false)
		if !foundNpu || npuName == constants.CenterNpuName {
			continue
		}
		oldNpuQuantity, ok := resourceList[constants.CenterNpuName]
		if !ok {
			continue
		}
		resourceList[v1.ResourceName(npuName)] = oldNpuQuantity
		delete(resourceList, constants.CenterNpuName)
	}
}

func parseQuantity(v interface{}) (resource.Quantity, bool) {
	const smallestPrec = -1
	var valStr string
	switch val := v.(type) {
	case float64:
		valStr = strconv.FormatFloat(val, 'f', smallestPrec, constants.BitSize64)
	case string:
		valStr = val
	default:
		return resource.Quantity{}, false
	}
	quantity, err := resource.ParseQuantity(valStr)
	if err != nil {
		return resource.Quantity{}, false
	}
	return quantity, true
}
