// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package monitors for docker monitor
package monitors

import (
	"time"

	"huawei.com/mindx/common/envutils"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/constants"
)

const (
	dockerMonitorInterval = 1 * time.Minute
	dockerMonitorName     = "docker"
)

var dockerTask = &cronTask{
	alarmId:         almutils.DockerAbnormal,
	name:            dockerMonitorName,
	interval:        dockerMonitorInterval,
	checkStatusFunc: checkDockerStatus,
}

func checkDockerStatus() error {
	_, err := envutils.RunCommand(constants.DockerCmd, envutils.DefCmdTimeoutSec, "version")
	return err
}
