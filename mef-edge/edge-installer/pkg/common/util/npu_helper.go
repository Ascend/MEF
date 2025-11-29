// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package util

import (
	"edge-installer/pkg/common/checker"
	"edge-installer/pkg/common/constants"
)

const (
	wholeNpuResRegex  = "huawei.com/Ascend[0-9a-zA-Z]{1,64}"
	wholeNpuResFormat = "^" + wholeNpuResRegex + "$"
)

// IsWholeNpu indicate this resource is a physical npu
func IsWholeNpu(resName string) bool {
	return checker.RegexStringChecker(resName, wholeNpuResFormat)
}

// IsSharableNpu indicate a resource is a sharable npu
func IsSharableNpu(resName string) bool {
	return resName == constants.SharableNpuName
}

// IsNpu indicate if a resource is a npu: sharable npu,a whole npu or a virtual npu
func IsNpu(resName string) bool {
	return IsSharableNpu(resName) || IsWholeNpu(resName)
}

// FindMostQualifiedNpu to find the most qualified npu
func FindMostQualifiedNpu(resObj interface{}) (string, bool) {
	if resObj == nil {
		return "", false
	}
	resList, ok := resObj.(map[string]interface{})
	if !ok {
		return "", false
	}
	for key, _ := range resList {
		if IsNpu(key) {
			return key, true
		}
	}
	return "", false
}
