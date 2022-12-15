// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package util this file for run command
package util

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"huawei.com/mindx/common/hwlog"
)

func shellDefenceCheck(cmds []string) bool {
	for _, cmd := range cmds {
		if strings.ContainsAny(cmd, IllegalChars) {
			return true
		}
	}
	return false
}

// RunCommand run command, return the output of command
func RunCommand(name string, arg ...string) (string, error) {
	if shellDefenceCheck(append(arg, name)) {
		fmt.Printf("shell check arg: %v\n", arg)
		hwlog.RunLog.Error("exec command check failed: contain illegal chars")
		return "", errors.New("exec command check failed")
	}
	cmd := exec.Command(name, arg...)
	ret, err := cmd.Output()
	if err != nil {
		return string(ret), err
	}
	return strings.Trim(string(ret), "\n"), nil
}
