//  Copyright(c) 2023. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils provides the util func about int
package utils

import (
	"math"
)

// MaxInt get max value between input
func MaxInt(input ...int) int {
	var res = math.MinInt
	for _, v := range input {
		if v > res {
			res = v
		}
	}
	return res
}

// MinInt get min value between input
func MinInt(input ...int) int {
	var res = math.MaxInt
	for _, v := range input {
		if v < res {
			res = v
		}
	}
	return res
}
