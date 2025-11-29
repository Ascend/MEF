// Copyright (c)  2024. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package securitysetters
package securitysetters

import (
	"errors"

	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"
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
