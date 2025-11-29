// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package msgchecker

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

type mefPodChecker struct {
	podChecker
}

func (pc *mefPodChecker) checkHostPath(podInfo *types.Pod) error {
	hostPaths := getPodHostPath(podInfo.Spec.Volumes)
	whiteListSet := utils.NewSet(configpara.GetPodConfig().HostPath...)
	for _, hostPath := range hostPaths {
		// message from mef center do not allow to use model file yet
		if filepath.Clean(hostPath) == modelFileDir {
			return fmt.Errorf("model file path not permitted in host path")
		}
		if !whiteListSet.Find(filepath.Clean(hostPath)) {
			return fmt.Errorf("hostpath [%s] Verification failed: not in whitelist", hostPath)
		}
	}
	return nil
}

// checkContainersNumber [method] for checking if the total numbers of container is out of MEFEdge's limit.
func (pc *mefPodChecker) checkContainersNumber(podInfo *types.Pod) error {
	podMap, err := statusmanager.GetPodStatusMgr().GetAll()
	if err != nil {
		return fmt.Errorf("get depolyed pod failed: %v", err)
	}
	containerNumbers := len(podInfo.Spec.Containers)
	// only allows orphaned pods in the edgecore database in the podpatch response,
	// and delete messages do not have containers.
	if containerNumbers <= 0 && (!pc.isPatch && pc.operation != constants.OptDelete) {
		return errors.New("pod has no containers")
	}
	var deployedContainerCount int
	for _, podString := range podMap {
		var podInDb types.Pod
		if err = json.Unmarshal([]byte(podString), &podInDb); err != nil {
			return errors.New("unmarshal pod failed")
		}
		// skip count container if the pod is already in db
		if podInDb.Name == podInfo.Name {
			continue
		}
		deployedContainerCount += len(podInfo.Spec.Containers)
	}

	edgeMaxContainerNumber := configpara.GetPodConfig().MaxContainerNumber
	if deployedContainerCount+containerNumbers > edgeMaxContainerNumber {
		return fmt.Errorf("container num in mef edge is out of limit[%d]", edgeMaxContainerNumber)
	}
	return nil
}

func (pc *mefPodChecker) checkConfigMapVolume(podInfo *types.Pod) error {
	for _, v := range podInfo.Spec.Volumes {
		if v.ConfigMap != nil {
			return errors.New("cur config not support config map")
		}
	}

	return nil
}

func (pc *mefPodChecker) checkEmptyDirVolume(podInfo *types.Pod) error {
	for _, v := range podInfo.Spec.Volumes {
		if v.EmptyDir != nil {
			return errors.New("cur config not support empty dir")
		}
	}

	return nil
}
func (pc *mefPodChecker) checkContainerWhetherChanged(podInfo *types.Pod) error {
	// 仅优雅删除pod时需要校验容器名是否变更，如果存在变更，pod删除后会存在残留容器运行
	if !isPodGraceDelete(podInfo.DeletionTimestamp) {
		return nil
	}

	content, err := statusmanager.GetPodStatusMgr().Get(constants.ResMefPodPrefix + podInfo.Name)
	if err != nil {
		hwlog.RunLog.Warnf("get pod:[%s] from db failed: %v", podInfo.Name, err)
		return nil
	}

	var originPod types.Pod
	if err = json.Unmarshal([]byte(content), &originPod); err != nil {
		return errors.New("unmarshal pod error")
	}

	if isContainerNameChanged(&originPod, podInfo) {
		return errors.New("container name in pod has changed")
	}

	return nil
}

func (pc *mefPodChecker) check(podInfo *types.Pod) error {
	var cc = mefContainerChecker{containerChecker: containerChecker{operation: pc.operation}}

	var configCheckers = []func(*types.Pod) error{
		pc.checkContainersNumber,
		pc.checkContainerWhetherChanged,
		pc.checkPodResources,
		pc.checkHostPath,
		pc.checkHostNetwork,
		pc.checkHostPid,
		pc.checkPodPorts,
		pc.checkConfigMapVolume,
		pc.checkEmptyDirVolume,
		pc.checkPodVolumeNameDuplicate,
		cc.check,
	}

	for idx, cf := range configCheckers {
		if err := cf(podInfo); err != nil {
			hwlog.RunLog.Errorf("[%d] check pod [%v] failed, %v", idx, podInfo.Name, err)
			return err
		}
	}
	return nil
}
