// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
