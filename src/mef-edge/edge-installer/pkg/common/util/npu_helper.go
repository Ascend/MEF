// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
