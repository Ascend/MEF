// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common base process used
package common

// ClearSliceByteMemory clears slice in memory
func ClearSliceByteMemory(sliceByte []byte) {
	for i := 0; i < len(sliceByte); i++ {
		sliceByte[i] = 0
	}
}
