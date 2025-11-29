// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package monitors
package monitors

import (
	"time"

	"edge-installer/pkg/common/almutils"
	"edge-installer/pkg/common/util"
)

const (
	dbMonitorInterval = 5 * time.Minute
	dbMonitorName     = "database"
)

var dbTask = &cronTask{
	alarmId:         almutils.EdgeDBAbnormal,
	name:            dbMonitorName,
	interval:        dbMonitorInterval,
	checkStatusFunc: util.CheckEdgeDbIntegrity,
}
