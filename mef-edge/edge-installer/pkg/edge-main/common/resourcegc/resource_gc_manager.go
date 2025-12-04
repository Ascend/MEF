// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package resourcegc do gc job for edge-main and edge-core
package resourcegc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/database"
)

const gcInterval = 5 * time.Minute

// ResGCManager a manager for garbage collection of resources
type ResGCManager struct {
	historyPods      *utils.Set
	historyResources *utils.Set
	failedCases      *utils.Set
}

// NewResourceGCManager do resource garbage collection
func NewResourceGCManager() *ResGCManager {
	gcManager := ResGCManager{
		historyPods:      utils.NewSet(),
		historyResources: utils.NewSet(),
		failedCases:      utils.NewSet(),
	}
	return &gcManager
}

// StartGcJob do resource garbage collect
func (gc *ResGCManager) StartGcJob(ctx context.Context) {
	hwlog.RunLog.Info("resource gc manager start")
	ticker := time.NewTicker(gcInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			hwlog.RunLog.Info("context done, resource garbage collection closed")
			return
		case _, ok := <-ticker.C:
			if !ok {
				hwlog.RunLog.Warn("resource gc stopped")
				return
			}
			gc.startGC()
		}
	}
}

func (gc *ResGCManager) startGC() {
	hwlog.RunLog.Info("resource garbage collection start to recycle unused resources")
	unusedCmResources := gc.computeUnusedResources()
	unusedPods := gc.getUnusedPods()

	unusedAll := unusedCmResources.Union(unusedPods)
	if gc.failedCases != nil {
		unusedAll = unusedCmResources.Union(gc.failedCases)
	}
	gc.cleanUnusedResource(unusedAll)
}

func (gc *ResGCManager) computeUnusedResources() *utils.Set {
	unusedResources := utils.NewSet()

	usedResources, err := getUsedResources()
	if err != nil {
		hwlog.RunLog.Errorf("get all pod used resources failed, error: %v", err)
		return unusedResources
	}

	allResources, err := getAllResources()
	if err != nil {
		hwlog.RunLog.Errorf("get all resources failed, error: %v", err)
		return unusedResources
	}

	unusedResources = allResources.Difference(usedResources)
	// now historyResources has past gc.period time
	// compute resource after gc.period time still unused
	unusedResources = unusedResources.Intersection(gc.historyResources)
	gc.historyResources = allResources

	hwlog.RunLog.Info("compute unused pod successful")
	return unusedResources
}

func (gc *ResGCManager) getUnusedPods() *utils.Set {
	unusedPods := utils.NewSet()

	pods, err := database.GetMetaRepository().GetByType(constants.ResourceTypePod)
	if err != nil {
		hwlog.RunLog.Error("get pod metas from db failed")
		return unusedPods
	}

	for _, podInfo := range pods {
		pod := v1.Pod{}
		if err = json.Unmarshal([]byte(podInfo.Value), &pod); err != nil {
			hwlog.RunLog.Errorf("unmarshal pod info failed, error: %v", err)
			return unusedPods
		}

		if pod.DeletionTimestamp != nil || len(pod.Spec.Containers) == 0 {
			unusedPods.Add(podInfo.Key)
		}
	}

	res := unusedPods.Intersection(gc.historyPods)
	gc.historyPods = unusedPods
	return res
}

func (gc *ResGCManager) cleanUnusedResource(unusedResource *utils.Set) {
	if unusedResource == nil || len(unusedResource.List()) == 0 {
		hwlog.RunLog.Info("unused resource set is empty, skip clean operation")
		return
	}
	gc.failedCases = utils.NewSet()
	for _, resourceName := range unusedResource.List() {
		hwlog.RunLog.Infof("start to clear unused resource, %s", resourceName)
		if err := gc.sendClearMsgToEdgeCore(resourceName); err != nil {
			gc.failedCases.Add(resourceName)
		}
	}
}

func (gc *ResGCManager) sendClearMsgToEdgeCore(resourceName string) error {
	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("gc manager new message for edgecore failed, error: %v", err)
		return err
	}

	msg.SetRouter(constants.ModCloudCore, constants.ModEdgeCore, constants.OptDelete, resourceName)
	msg.Header.Timestamp = time.Now().UnixMilli()
	msg.Header.ID = msg.Header.Id
	msg.Header.Sync = false
	if err = msg.FillContent(map[string]interface{}{}); err != nil {
		hwlog.RunLog.Errorf("fill map into content failed: %v", err)
		return errors.New("fill map into content failed")
	}

	msg.SetKubeEdgeRouter(constants.ModEdgeMain, constants.ResourceModule, constants.OptDelete, resourceName)
	resp, err := modulemgr.SendSyncMessage(msg, constants.WsSycMsgWaitTime)
	if err != nil {
		hwlog.RunLog.Errorf("gc manager send message to egdecore failed, resourceName=%s, error: %v", resourceName, err)
		return err
	}
	var respCntStr string
	if err = resp.ParseContent(&respCntStr); err != nil {
		hwlog.RunLog.Errorf("get resp content failed: %v", err)
		return errors.New("get resp content failed")
	}
	if respCntStr != constants.OK {
		hwlog.RunLog.Errorf("bad response type, resourceName=%s, type=%T", resourceName, resp.Content)
		return err
	}
	return nil
}

func getUsedResources() (*utils.Set, error) {
	podResources := utils.NewSet()
	podMetas, err := database.GetMetaRepository().GetByType(constants.ResourceTypePod)
	if err != nil {
		hwlog.RunLog.Error("get pod metas from db failed")
		return nil, errors.New("get pod metas from db failed")
	}

	for _, podMeta := range podMetas {
		pod := v1.Pod{}
		if err = json.Unmarshal([]byte(podMeta.Value), &pod); err != nil {
			hwlog.RunLog.Errorf("unmarshal pod info failed, error: %v", err)
			return nil, fmt.Errorf("unmarshal pod info failed, %v", err)
		}
		podResources = podResources.Union(parsePodResources(pod))
	}

	return podResources, nil
}

func parsePodResources(pod v1.Pod) *utils.Set {
	podResources := utils.NewSet()
	for _, volume := range pod.Spec.Volumes {
		if volume.ConfigMap != nil {
			configmapResName := fmt.Sprintf("%s/%s/%s",
				pod.Namespace, constants.ResourceTypeConfigMap, volume.ConfigMap.Name)
			podResources.Add(configmapResName)
		}
	}
	return podResources
}

func getAllResources() (*utils.Set, error) {
	allResources := utils.NewSet()
	// Currently, only configmap is supported.
	resourceList := []string{constants.ResourceTypeConfigMap}
	for _, resourceType := range resourceList {
		resources, err := getResourcesByType(resourceType)
		if err != nil {
			return nil, err
		}
		allResources = allResources.Union(resources)
	}
	return allResources, nil
}

func getResourcesByType(resourceType string) (*utils.Set, error) {
	keys, err := database.GetMetaRepository().GetKeyByType(resourceType)
	if err != nil {
		hwlog.RunLog.Error("get resource metas from db failed")
		return nil, errors.New("get resource metas from db failed")
	}

	return utils.NewSet(keys...), nil
}
