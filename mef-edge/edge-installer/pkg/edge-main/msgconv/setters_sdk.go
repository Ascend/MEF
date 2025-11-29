// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.
//go:build MEFEdge_SDK

// Package msgconv
package msgconv

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"
)

func setNodePatchRequest(_ *model.Message, content interface{}) error {
	patch, ok := content.(*string)
	if !ok {
		return errors.New("bad content type")
	}
	if patch == nil {
		return errors.New("nil pointer content is not allowed")
	}
	return setNpuForUpstreamNodePatch(patch)
}

func setNodePatchResponse(_ *model.Message, content interface{}) error {
	node, ok := content.(*NodeResp)
	if !ok {
		return errors.New("bad content type")
	}
	if node == nil {
		return errors.New("nil pointer content is not allowed")
	}
	if node.Object == nil {
		return nil
	}
	return setNpuForDownstreamNode(node.Object)
}

func setNodeInsertRequest(_ *model.Message, content interface{}) error {
	node, ok := content.(*v1.Node)
	if !ok {
		return errors.New("bad content type")
	}
	if node == nil {
		return errors.New("nil pointer content is not allowed")
	}
	return setNpuForUpstreamNode(node)
}

// setNodeResponse the response for node-query and node-insertion contains the same router and different content type
func setNodeResponse(_ *model.Message, content interface{}) error {
	respPtr, ok := content.(*map[string]interface{})
	if !ok {
		return errors.New("bad content type")
	}
	if respPtr == nil {
		return errors.New("nil pointer content is not allowed")
	}

	contentPtr, nodePtr, err := parseNodeResponseContent(*respPtr)
	if err != nil {
		return err
	}
	// no need to modify response if content doesn't contains a valid node object
	if nodePtr == nil {
		return nil
	}
	if err := setNpuForDownstreamNode(nodePtr); err != nil {
		return err
	}

	dataBytes, err := json.Marshal(contentPtr)
	if err != nil {
		return err
	}
	var newResp map[string]interface{}
	if err := json.Unmarshal(dataBytes, &newResp); err != nil {
		return err
	}
	*respPtr = newResp
	return nil
}

func setPodPatchRequest(_ *model.Message, content interface{}) error {
	patch, ok := content.(*string)
	if !ok {
		return errors.New("bad content type")
	}
	if patch == nil {
		return errors.New("nil pointer content is not allowed")
	}
	return setNpuForUpstreamPodPatch(patch)
}

func setPodPatchResponse(_ *model.Message, content interface{}) error {
	podResp, ok := content.(*PodResp)
	if !ok {
		return errors.New("bad content type")
	}
	if podResp == nil {
		return errors.New("nil pointer content is not allowed")
	}
	if podResp.Object == nil {
		return nil
	}
	return setNpuForDownstreamPod(podResp.Object)
}

func setPodUpdateRequest(_ *model.Message, content interface{}) error {
	pod, ok := content.(*v1.Pod)
	if !ok {
		return errors.New("bad content type")
	}
	if pod == nil {
		return errors.New("nil pointer content is not allowed")
	}
	return setNpuForDownstreamPod(pod)
}

func parseNodeResponseContent(resp map[string]interface{}) (interface{}, *v1.Node, error) {
	dataBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, nil, err
	}

	var (
		_, errFieldExists    = resp["Err"]
		_, objectFieldExists = resp["Object"]
	)
	// response for query operation
	if !(errFieldExists || objectFieldExists) {
		var node v1.Node
		if err := json.Unmarshal(dataBytes, &node); err != nil {
			return nil, nil, err
		}
		return &node, &node, nil
	}

	// response for insert operation
	var nodeResp NodeResp
	if err := json.Unmarshal(dataBytes, &nodeResp); err != nil {
		return nil, nil, err
	}
	// unsuccessful response, such as 409(AlreadyExists)
	if nodeResp.Object == nil || nodeResp.Object.Name == "" {
		return &nodeResp, nil, nil
	}

	// successful response
	return &nodeResp, nodeResp.Object, nil
}
