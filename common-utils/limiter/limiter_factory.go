// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package limiter for limiter factory
package limiter

import (
	"time"
)

// IndependentLimiterFactory for single limiter factory
type IndependentLimiterFactory interface {
	Create() IndependentLimiter
}

// TimeWindowLimiterFactory for limiter request in expired time
type TimeWindowLimiterFactory struct {
	expiredTime time.Duration
}

// Create to init TimeWindowLimiterFactory
func (twf *TimeWindowLimiterFactory) Create() IndependentLimiter {
	return &TimeWindowLimiter{
		expiredTime: twf.expiredTime,
	}
}

// BurstMgsLimiterFactory for burst request
type BurstMgsLimiterFactory struct {
	limit float64
	burst int
}

// Create to init BurstMgsLimiterFactory
func (twf *BurstMgsLimiterFactory) Create() IndependentLimiter {
	return &RpsLimiter{
		rps:   twf.limit,
		burst: twf.burst,
	}
}
