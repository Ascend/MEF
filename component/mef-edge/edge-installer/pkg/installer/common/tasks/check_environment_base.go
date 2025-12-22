// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks this file for check environment base task
package tasks

import (
	"fmt"
	"os/exec"

	"huawei.com/mindx/common/hwlog"
)

var (
	necessaryTools = [...]string{
		"sh",
		"cat",
		"docker",
		"dmidecode",
		"systemctl",
		"useradd",
	}
	recommendTool = "haveged"
)

// CheckEnvironmentBaseTask the task for check environment base
type CheckEnvironmentBaseTask struct{}

// CheckNecessaryTools check necessary tools
func (cet *CheckEnvironmentBaseTask) CheckNecessaryTools() error {
	for _, tool := range necessaryTools {
		if _, err := exec.LookPath(tool); err != nil {
			fmt.Printf("[%s] not found, please check whether the tool is installed\n", tool)
			return fmt.Errorf("look path of [%s] failed, error: %v", tool, err)
		}
	}

	if _, err := exec.LookPath(recommendTool); err != nil {
		fmt.Printf("warning: [%s] not found, system may be slow to read random numbers without it\n", recommendTool)
		hwlog.RunLog.Warnf("[%s] not found, system may be slow to read random numbers without it", recommendTool)
	}

	hwlog.RunLog.Info("check necessary tools success")
	return nil
}
