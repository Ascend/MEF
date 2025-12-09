// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package msgconv
package msgconv

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"

	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

// setKind the field `Kind` in content is required (from version 1.12)
func setKind(_ *model.Message, content interface{}) error {
	rv := reflect.ValueOf(content)
	if content == nil || rv.IsNil() {
		return errors.New("nil pointer content is not allowed")
	}
	switch value := content.(type) {
	case *v1.Secret:
		value.Kind = "Secret"
	case *v1.ConfigMap:
		value.Kind = "Configmap"
	case *v1.Pod:
		value.Kind = "Pod"
	case *v1.Node:
		value.Kind = "Node"
	default:
		return errors.New("not supported type")
	}
	return nil
}

// setSourceToEdgeController the module `controller` was renamed to `edgecontorller` (from version 1.12)
func setSourceToEdgeController(message *model.Message, _ interface{}) error {
	message.KubeEdgeRouter.Source = constants.EdgeControllerModule
	return nil
}

// setSourceToController the module `controller` was renamed to `edgecontorller` (from version 1.12)
func setSourceToController(message *model.Message, _ interface{}) error {
	message.KubeEdgeRouter.Source = constants.ControllerModule
	return nil
}

// setPodSpecForUpdate
// the field `NodeName` is required (from version 1.12)
// the field `EnableServiceLinks` is required (from version 1.12)
func setPodSpecForUpdate(_ *model.Message, content interface{}) error {
	pod, ok := content.(*v1.Pod)
	if !ok {
		return errors.New("unsupported type")
	}
	if pod == nil {
		return errors.New("nil pointer content is not allowed")
	}

	nodes, err := statusmanager.GetNodeStatusMgr().GetAll()
	if err != nil {
		return err
	}
	var nodeName string
	for key := range nodes {
		nodeName = filepath.Base(key)
		break
	}

	pod.Spec.NodeName = nodeName
	trueValue := true
	pod.Spec.EnableServiceLinks = &trueValue

	return setNpuForDownstreamPod(pod)
}

// setPodSpecForDelete edgecore requires an empty `PodSpec` to clean pod directory (from version 1.12)
func setPodSpecForDelete(_ *model.Message, content interface{}) error {
	pod, ok := content.(*v1.Pod)
	if !ok {
		return errors.New("unsupported type")
	}
	if pod == nil {
		return errors.New("nil pointer content is not allowed")
	}
	pod.Spec = v1.PodSpec{}
	return nil
}

// setAsync edgecore only accepts an async updating pod message (from version 1.12)
func setAsync(message *model.Message, _ interface{}) error {
	message.Header.Sync = false
	return nil
}

// setPodRestartRoute rewrites route for pod restart message
func setPodRestartRoute(message *model.Message, content interface{}) error {
	str, ok := content.(*string)
	if !ok {
		return errors.New("invalid content type")
	}
	if str == nil {
		return errors.New("nil pointer content is not allowed")
	}

	resourceSplit := strings.Split(message.KubeEdgeRouter.Resource, "/")
	if len(resourceSplit) == 0 {
		return errors.New("failed to get podName from msg")
	}
	podName := resourceSplit[len(resourceSplit)-1]
	message.KubeEdgeRouter.Resource = constants.ActionPod
	message.Router.Resource = constants.ActionPod
	*str = podName
	return nil
}
