// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
