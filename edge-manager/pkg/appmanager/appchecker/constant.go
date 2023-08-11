// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
	minCmMountsCount       = 0
	maxCmMountsCount       = 4
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
	envNameReg             = "^[a-zA-Z][a-zA-z0-9._-]{0,30}[a-zA-Z0-9]$"
	envValueReg            = "^[a-zA-Z0-9 _./:-]{1,512}$"

	minAppId       = 1
	maxAppId       = math.MaxInt64
	minTemplateId  = 1
	maxTemplateId  = math.MaxInt64
	minNodeGroupId = 1
	maxNodeGroupId = math.MaxInt64
	minList        = 1
	maxList        = 1024
)
