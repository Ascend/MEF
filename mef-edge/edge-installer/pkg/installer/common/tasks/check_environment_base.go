// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
