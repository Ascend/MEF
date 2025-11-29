// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package common define common variable
package common

import "math"

const (
	// InvaidVal InvalidVal for NPU Invalid vaule
	InvaidVal = 0
	// Success for interface return code
	Success = 0
	// RetError return error when the function failed
	RetError = -1
	// Percent constant of 100
	Percent = 100
	// MaxErrorCodeCount number of error codes
	MaxErrorCodeCount = 128
	// UnRetError return unsigned int error
	UnRetError = math.MaxUint32

	// HiAIMaxCardID max card id for Ascend chip
	HiAIMaxCardID = math.MaxInt32

	// HiAIMaxCardNum max card number
	HiAIMaxCardNum = 64

	// HiAIMaxDeviceNum max device number
	HiAIMaxDeviceNum = 4

	// NpuType present npu chip
	NpuType = 0

	// Ascend310 ascend 310 chip
	Ascend310 = "Ascend310"
	// Ascend310B ascend 310B chip
	Ascend310B = "Ascend310B"
	// Ascend310P ascend 310P chip
	Ascend310P = "Ascend310P"
)

const (
	rootUID       = 0
	maxPathDepth  = 20
	maxPathLength = 1024
	// DefaultWriteFileMode  default file mode for write permission check
	DefaultWriteFileMode = 0022

	ldSplitLen     = 2
	ldLibNameIndex = 0
	ldLibPathIndex = 1
	ldCommand      = "/sbin/ldconfig"
	ldParam        = "--print-cache"
	ldLibPath      = "LD_LIBRARY_PATH"
	grepCommand    = "/bin/grep"
)
