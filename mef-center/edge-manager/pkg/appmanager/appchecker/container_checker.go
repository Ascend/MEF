// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appchecker container checker
package appchecker

import (
	"fmt"

	"huawei.com/mindx/common/checker"
)

// GetContainerChecker [method] for get container checker
func GetContainerChecker(field string) *ContainerChecker {
	return &ContainerChecker{
		modelChecker: checker.ModelChecker{Field: field, Required: true},
	}
}

// ContainerChecker [struct] for Container checker
type ContainerChecker struct {
	modelChecker checker.ModelChecker
}

func (cc *ContainerChecker) init() {
	cc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("Name", nameReg, true),
		checker.GetRegChecker("Image", imageNameReg, true),
		checker.GetRegChecker("ImageVersion", imageVerReg, true),
		checker.GetFloatChecker("CpuRequest", minCpuQuantity, maxCpuQuantity, true),
		checker.GetFloatChecker("CpuLimit", minCpuQuantity, maxCpuQuantity, false),
		checker.GetIntChecker("MemRequest", minMemoryQuantity, maxMemoryQuantity, true),
		checker.GetIntChecker("MemLimit", minMemoryQuantity, maxMemoryQuantity, false),
		checker.GetIntChecker("Npu", minNpuQuantity, maxNpuQuantity, false),
		checker.GetListChecker("Command", checker.GetRegChecker("", cmdReg, true),
			minCmdCount, maxCmdCount, true,
		),
		checker.GetListChecker("Args", checker.GetRegChecker("", argsReg, true),
			minArgsCount, maxArgsCount, true,
		),
		checker.GetListChecker("Env", GetEnvVarChecker(""), minEnvCount, maxEnvCount, true),
		checker.GetListChecker("Ports", GetContainerPortChecker(""),
			minPortMapCount, maxPortMapCount, true,
		),
		checker.GetIntChecker("UserID", minUserId, maxUserId, false),
		checker.GetIntChecker("GroupID", minGroupId, maxGroupId, false),
		checker.GetListChecker("HostPathVolumes", GetHostPathVolumeChecker(""),
			minVolumeMountsCount, maxVolumeMountsCount, true),
	)
}

// Check [method] for check container parameters
func (cc *ContainerChecker) Check(data interface{}) checker.CheckResult {
	cc.init()
	checkResult := cc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("container checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
