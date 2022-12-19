// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package common

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"

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

func runCommandWithUser(uid uint32, name string, arg ...string) (string, error) {
	if shellDefenceCheck(append(arg, name)) {
		hwlog.RunLog.Error("exec command check failed: contain illegal chars")
		return "", errors.New("exec command check failed")
	}
	cmd := exec.Command(name, arg...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uid}}
	ret, err := cmd.Output()
	if err != nil {
		return string(ret), err
	}
	return strings.Trim(string(ret), "\n"), nil
}

// RunCommand run command, return the output of command
func RunCommand(name string, arg ...string) (string, error) {
	if shellDefenceCheck(append(arg, name)) {
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

// RunCommandWithUser run command with specified user, return the output of command
func RunCommandWithUser(uid uint32, name string, arg ...string) (string, error) {
	return runCommandWithUser(uid, name, arg...)
}

// IsProcessActive is process active
func IsProcessActive(pid int) (bool, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		hwlog.RunLog.Errorf("find process failed,error:%v", err)
		return false, err
	}
	if err = proc.Signal(syscall.Signal(0)); err != nil {
		hwlog.RunLog.Warnf("process (%d) is not active,error:%v", pid, err)
		return false, nil
	}
	return true, nil
}
