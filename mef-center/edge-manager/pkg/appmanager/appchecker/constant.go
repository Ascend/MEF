// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package appchecker constant
package appchecker

import (
	"math"
)

const (
	minContainerCountInPod = 1
	maxContainerCountInPod = 10
	minPortMapCount        = 0
	maxPortMapCount        = 16
	minEnvCount            = 0
	maxEnvCount            = 256
	minVolumeMountsCount   = 0
	maxVolumeMountsCount   = 256
	minContainerPort       = 1
	maxContainerPort       = 65535
	minHostPort            = 1024
	maxHostPort            = 65535
	minUserId              = 1
	maxUserId              = 65535
	minGroupId             = 1
	maxGroupId             = 65535
	minCmdCount            = 0
	maxCmdCount            = 16
	minArgsCount           = 0
	maxArgsCount           = 16
	minCpuQuantity         = 0.01
	maxCpuQuantity         = 1000
	minMemoryQuantity      = 4           // 4 MB
	maxMemoryQuantity      = 1000 * 1024 // 1000GB
	minNpuQuantity         = 0
	maxNpuQuantity         = 32
	nameReg                = "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$"
	imageNameReg           = "^[a-zA-Z0-9:_/.-]{1,256}$"
	imageVerReg            = "^[a-zA-Z0-9_.-]{1,32}$"
	cmdReg                 = "^[a-zA-Z0-9 _./-]{0,255}[a-zA-Z0-9]$"
	argsReg                = "^[a-zA-Z0-9 =_./-]{0,255}[a-zA-Z0-9]$"
	descriptionReg         = "^[\\S ]{0,512}$"
	envNameReg             = "^[a-zA-Z][a-zA-Z0-9._-]{0,30}[a-zA-Z0-9]$"
	envValueReg            = "^[a-zA-Z0-9 _./:-]{1,512}$"

	minAppId       = 1
	maxAppId       = math.MaxUint32
	minNodeGroupId = 1
	maxNodeGroupId = math.MaxUint32
	minList        = 1
	maxList        = 1024
)
