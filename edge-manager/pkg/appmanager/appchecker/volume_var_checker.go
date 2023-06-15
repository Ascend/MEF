// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appchecker volume mounts variable checker
package appchecker

import (
	"fmt"

	"huawei.com/mindx/common/checker"

	"edge-manager/pkg/util"
)

// GetVolumeVarChecker [method] for get volume mounts variable checker
func GetVolumeVarChecker(field string) *VolumeVarChecker {
	return &VolumeVarChecker{
		modelChecker: checker.ModelChecker{Field: field, Required: true},
	}
}

// VolumeVarChecker [struct] for volume mounts var checker
type VolumeVarChecker struct {
	modelChecker checker.ModelChecker
}

func (vvc *VolumeVarChecker) init() {
	vvc.modelChecker.Checker = checker.GetAndChecker(
		checker.GetRegChecker("LocalVolumeName", localVolumeReg, true),
		checker.GetRegChecker("MountPath", configmapMountPathReg, true),
		checker.GetRegChecker("ConfigmapName", configmapNameReg, true),
	)
}

// Check [method] for check volume mounts variable parameters
func (vvc *VolumeVarChecker) Check(data interface{}) checker.CheckResult {
	vvc.init()
	checkResult := vvc.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("volume mounts var checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

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
