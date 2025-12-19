// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appchecker volume mounts variable checker
package appchecker

import (
	"fmt"

	"huawei.com/mindx/common/checker"

	"edge-manager/pkg/util"
)

// GetHostPathVolumeChecker [method] for get volume mounts variable checker
func GetHostPathVolumeChecker(field string) *HostPathVolumeChecker {
	return &HostPathVolumeChecker{
		modelChecker: checker.ModelChecker{Field: field, Required: true},
	}
}

// HostPathVolumeChecker [struct] for checking hostPath volume
type HostPathVolumeChecker struct {
	modelChecker checker.ModelChecker
}

func (hc *HostPathVolumeChecker) init() {
	hc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Name", nameReg, true),
		util.GetPathChecker("HostPath", true),
		util.GetPathChecker("MountPath", true),
	)
}

// Check [method] for check host path volume mounts variable parameters
func (hc *HostPathVolumeChecker) Check(data interface{}) checker.CheckResult {
	hc.init()
	checkResult := hc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("check hostPath volume mounts failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
