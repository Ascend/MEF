// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

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
