// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package util
package util

import (
	"encoding/json"
	"errors"
)

type jsonNode = interface{}
type jsonObject = map[string]interface{}

var (
	errNullPointer  = errors.New("nil map is not allowed")
	errBadPatchType = errors.New("patch type is not allowed")
)

// MergePatch implements JSON Merge Patch [rfc7386]
func MergePatch(targetBytes, patchBytes []byte) ([]byte, error) {
	var (
		target interface{}
		patch  interface{}
	)

	if err := json.Unmarshal(targetBytes, &target); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(patchBytes, &patch); err != nil {
		return nil, err
	}
	if patch == nil {
		return nil, errBadPatchType
	}
	if _, ok := patch.(string); ok {
		return nil, errBadPatchType
	}

	result, err := doMergePatch(target, patch)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func doMergePatch(target, patch jsonNode) (jsonNode, error) {
	objPatch, ok := patch.(jsonObject)
	if !ok {
		return patch, nil
	}
	if objPatch == nil {
		return nil, errNullPointer
	}

	objTarget, ok := target.(jsonObject)
	if !ok {
		objTarget = make(jsonObject)
	}
	if objTarget == nil {
		return nil, errNullPointer
	}

	for name, value := range objPatch {
		if value == nil {
			delete(objTarget, name)
			continue
		}
		mergedValue, err := doMergePatch(objTarget[name], value)
		if err != nil {
			return nil, err
		}
		objTarget[name] = mergedValue
	}
	return objTarget, nil
}
