// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package monitors for npu monitor
package monitors

import (
	"errors"
	"strings"
	"time"

	"huawei.com/mindx/common/envutils"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
)

const (
	npuMonitorInterval = 1 * time.Minute
	npuMonitorName     = "npu"
	shouldReadLine     = 2
	chipHealthBit      = 2
	chipHealthOk       = "OK"
	maxLineCount       = 100
)

var npuTask = &cronTask{
	alarmId:         almutils.NPUAbnormal,
	name:            npuMonitorName,
	interval:        npuMonitorInterval,
	checkStatusFunc: checkNpuStatus,
}

func stringContainsAny(originStr string, targets []string) bool {
	for _, target := range targets {
		if strings.Contains(originStr, target) {
			return true
		}
	}
	return false
}

func checkNpuStatus() error {
	var npuRet string
	var err error
	if npuRet, err = envutils.RunCommand(constants.NpuSmiCmd, envutils.DefCmdTimeoutSec, "info"); err != nil {
		return err
	}
	lines := strings.Split(npuRet, "\n")
	excludeWords := []string{"+", "NPU", "Chip", "npu-smi"}
	var npuContent []string
	var npuReadLineCount int
	iterationCount := 1
	for _, line := range lines {
		if iterationCount > maxLineCount {
			break
		}
		iterationCount++
		if stringContainsAny(line, excludeWords) {
			continue
		}
		npuContent = append(npuContent, strings.Fields(strings.ReplaceAll(line, "|", ""))...)
		if npuReadLineCount%shouldReadLine == 0 {
			npuReadLineCount++
			continue
		}
		if len(npuContent) <= chipHealthBit {
			continue
		}
		if npuContent[chipHealthBit] != chipHealthOk {
			return errors.New("chip health is not ok")
		}
		npuContent = npuContent[:0]
	}
	return nil
}
