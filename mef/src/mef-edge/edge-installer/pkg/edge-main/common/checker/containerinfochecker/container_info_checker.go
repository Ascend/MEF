// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package containerinfochecker for container info checker
package containerinfochecker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/checker/valuer"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

const (
	minContainerListLen = 1
	maxContainerListLen = 10
	minModelFileListLen = 0
	maxModelFileListLen = 256
)

var modelFileSuffixList = []string{".om", ".tar.gz", ".zip"}

// CheckContainerInfo [method] do actual job to check update content of container info
func CheckContainerInfo(content []byte) error {
	var err error
	var containerInfo types.UpdateContainerInfo

	if err = json.Unmarshal(content, &containerInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal update container info message failed, error: %v", err)
		return errors.New("unmarshal update container info message failed")
	}

	if checkResult := newContainerInfoChecker().Check(containerInfo); !checkResult.Result {
		hwlog.RunLog.Errorf("container info check failed, error: %s", checkResult.Reason)
		return errors.New(checkResult.Reason)
	}

	return nil
}

type containerInfoChecker struct {
	modelChecker checker.ModelChecker
}

func newContainerInfoChecker() *containerInfoChecker {
	return &containerInfoChecker{}
}

func (c *containerInfoChecker) init() {
	c.modelChecker.Checker = checker.GetAndChecker(
		checker.GetStringChoiceChecker("Operation", []string{"update"}, true),
		checker.GetStringChoiceChecker("Source", []string{"all"}, true),
		checker.GetRegChecker("PodName", constants.PodNameReg, true),
		checker.GetRegChecker("PodUid", "^"+constants.UUIDRegex+"$", true),
		checker.GetRegChecker("Uuid", "^"+constants.UUIDRegex+"$", true),
		checker.GetListChecker("Container",
			checker.GetListChecker("ModelFile", &modelFileChecker{}, minModelFileListLen, maxModelFileListLen,
				true),
			minContainerListLen, maxContainerListLen, true),
	)
}

func (c *containerInfoChecker) Check(data interface{}) checker.CheckResult {
	c.init()
	checkResult := c.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(
			fmt.Sprintf("container info checker check failed, error: %s", checkResult.Reason))
	}
	hwlog.RunLog.Info("container info checker check success")
	return checker.NewSuccessResult()
}

type modelFileChecker struct {
	modelChecker checker.ModelChecker
}

func (m *modelFileChecker) init() {
	m.modelChecker.Checker = checker.GetAndChecker(
		checker.GetStringChoiceChecker("ActiveType", []string{"cold_update", "hot_update"}, true),
		checker.GetAndChecker(
			checker.GetRegChecker("Name", constants.ModelFileNameReg, true),
			checker.GetStringExcludeChecker("Name", []string{".."}, true),
			&modelFileNameChecker{checkFunc: checkNameSuffix},
		),
		checker.GetAndChecker(
			checker.GetRegChecker("Version", constants.VersionLenReg, true),
		),
	)
}

func (m *modelFileChecker) Check(data interface{}) checker.CheckResult {
	m.init()
	checkResult := m.modelChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(
			fmt.Sprintf("model file checker check failed, error: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

type modelFileNameChecker struct {
	checkFunc func(data interface{}) checker.CheckResult
}

func (m *modelFileNameChecker) Check(data interface{}) checker.CheckResult {
	return m.checkFunc(data)
}

func checkNameSuffix(data interface{}) checker.CheckResult {
	stringValuer := valuer.StringValuer{}
	name, err := stringValuer.GetValue(data, "Name")
	if err != nil {
		return checker.NewFailedResult(fmt.Sprintf("get model file name value failed, error: %v", err))
	}
	for _, suffix := range modelFileSuffixList {
		if strings.HasSuffix(name, suffix) {
			return checker.NewSuccessResult()
		}
	}
	return checker.NewFailedResult(fmt.Sprintf("name suffix is invalid"))
}
