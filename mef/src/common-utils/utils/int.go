// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
