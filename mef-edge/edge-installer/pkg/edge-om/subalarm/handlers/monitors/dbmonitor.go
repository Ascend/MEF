// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
