// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package msgchecker

import (
	"fmt"

	"edge-installer/pkg/edge-main/common/checker/msgchecker/types"
)

type fdContainerChecker struct {
	containerChecker
}

func (c fdContainerChecker) check(pod *types.Pod) error {
	if isPodGraceDelete(pod.DeletionTimestamp) {
		return nil
	}

	checkFuncs := []func(c *types.Container) error{
		c.checkPortMappingPara,
		c.checkContainerEnv,
		c.checkContainerResource,
		c.checkContainerVolumeMount,
		c.checkContainerProbePara,
		c.checkContainerSecurityContext,
	}

	var containerNames = map[string]struct{}{}
	for _, container := range pod.Spec.Containers {
		for _, function := range checkFuncs {
			if err := function(&container); err != nil {
				return err
			}
		}
		if _, ok := containerNames[container.Name]; ok {
			return fmt.Errorf("container name [%s] is not unique", container.Name)
		}
		containerNames[container.Name] = struct{}{}
	}
	return nil
}

func (c fdContainerChecker) checkContainerProbePara(container *types.Container) error {
	if err := checkProbePara(container.LivenessProbe); err != nil {
		return err
	}

	if err := checkProbePara(container.ReadinessProbe); err != nil {
		return err
	}

	return nil
}
