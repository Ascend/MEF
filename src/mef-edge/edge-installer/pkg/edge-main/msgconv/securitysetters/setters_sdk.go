// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package securitysetters
package securitysetters

import (
	"errors"

	"k8s.io/api/core/v1"

	"huawei.com/mindx/common/modulemgr/model"
)

// SetPodUpdate set pod update message
func SetPodUpdate(_ *model.Message, content interface{}) error {
	pod, ok := content.(*v1.Pod)
	if !ok {
		return errors.New("invalid content type")
	}
	if pod == nil {
		return errors.New("nil pointer content is not allowed")
	}

	for idx := range pod.Spec.Containers {
		tempContainer := v1.Container{}
		setContainerSecurityContext(&pod.Spec.Containers[idx], &tempContainer)
		pod.Spec.Containers[idx].SecurityContext = tempContainer.SecurityContext

		pod.Spec.Containers[idx].TerminationMessagePath = ""
	}
	return nil
}
