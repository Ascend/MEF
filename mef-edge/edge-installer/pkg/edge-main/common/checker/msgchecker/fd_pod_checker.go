// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package msgchecker

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
	"edge-installer/pkg/edge-main/common/configpara"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

type fdPodChecker struct {
	podChecker
}

func (pc *fdPodChecker) checkHostPath(podInfo *types.Pod) error {
	hostPaths := getPodHostPath(podInfo.Spec.Volumes)
	whiteListSet := utils.NewSet(configpara.GetPodConfig().HostPath...)
	for _, hostPath := range hostPaths {
		if strings.HasPrefix(hostPath, modelFilePrefix) && whiteListSet.Find(constants.ModeFileActiveDir) {
			continue
		}

		if !whiteListSet.Find(filepath.Clean(hostPath)) {
			return fmt.Errorf("hostpath [%s] verification failed: not in whitelist", hostPath)
		}
	}
	return nil
}
func (pc *fdPodChecker) checkContainersNumber(podInfo *types.Pod) error {
	containerNumbers := len(podInfo.Spec.Containers)

	if containerNumbers == 0 {
		return fmt.Errorf("there's no container in pod")
	}

	podMap, err := statusmanager.GetPodStatusMgr().GetAll()
	if err != nil {
		return fmt.Errorf("get depolyed pod failed: %s", err.Error())
	}
	var deployedContainerCount int
	for _, podString := range podMap {
		var podInDb v1.Pod
		if err = json.Unmarshal([]byte(podString), &podInDb); err != nil {
			return errors.New("unmarshal pod failed")
		}
		// skip count container if the pod is already in db
		if podInDb.Name == podInfo.Name {
			continue
		}
		deployedContainerCount += len(podInDb.Spec.Containers)
	}

	edgeMaxContainerNumber := configpara.GetPodConfig().MaxContainerNumber
	if deployedContainerCount+containerNumbers > edgeMaxContainerNumber {
		return fmt.Errorf("container num is out of limit[%d]", edgeMaxContainerNumber)
	}

	return nil
}

func (pc *fdPodChecker) checkContainerWhetherChanged(podInfo *types.Pod) error {
	// only when graceful delete pod need to check whether container info is changed no not
	if !isPodGraceDelete(podInfo.DeletionTimestamp) {
		return nil
	}

	content, err := statusmanager.GetPodStatusMgr().Get(constants.ActionPod + podInfo.Name)
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
func (pc *fdPodChecker) check(podInfo *types.Pod) error {
	var cc = fdContainerChecker{containerChecker: containerChecker{operation: pc.operation}}
	var configCheckers = []func(*types.Pod) error{
		pc.checkContainersNumber,
		pc.checkContainerWhetherChanged,
		pc.checkPodResources,
		pc.checkHostPath,
		pc.checkHostNetwork,
		pc.checkHostPid,
		pc.checkPodPorts,
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
