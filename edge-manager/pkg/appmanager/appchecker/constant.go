// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package appchecker constant
package appchecker

const (
	minContainerCountInPod = 1
	maxContainerCountInPod = 10
	minPortMapCount        = 0
	maxPortMapCount        = 16
	minEnvCount            = 0
	maxEnvCount            = 256
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
	minNpuQuantity         = 0.01
	maxNpuQuantity         = 32
	nameReg                = "^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9]){0,1}$"
	imageNameReg           = "^[a-z0-9]([a-z0-9_./-]{0,30}[a-z0-9]){0,1}$"
	imageVerReg            = "^[a-zA-Z0-9_.-]{1,32}$"
	cmdAndArgsReg          = "^[a-zA-Z0-9 _./-]{0,255}[a-zA-Z0-9]$"
	descriptReg            = "^[\\S ]{0,512}$"
	envNameReg             = "^[a-zA-Z][a-zA-z0-9._-]{0,30}[a-zA-Z0-9]$"
	envValueReg            = "^[a-zA-Z0-9 _./-]{1,512}$"
)
