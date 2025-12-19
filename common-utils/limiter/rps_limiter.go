// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package limiter, limit handle msg
package limiter

import (
	"sync"
	"time"
)

// RpsLimiter - request frequency limiter
type RpsLimiter struct {
	rps    float64 // represents how many requests are allowed in ONE SECOND
	burst  int     // represents maximum requests are allowed at the same time
	tokens float64
	last   time.Time
	mu     sync.Mutex
}

// RpsLimiterCfg rps limiter config
type RpsLimiterCfg struct {
	Rps   float64
	Burst int
}

// NewRpsLimiter returns a RpsLimiter.
// rps: request per second
// burst: max size for tokens bucket.
func NewRpsLimiter(rps float64, burst int) *RpsLimiter {
	return &RpsLimiter{
		rps:   rps,
		burst: burst,
	}
}

func (limiter *RpsLimiter) calculateTokens(t time.Time) float64 {
	last := limiter.last
	if t.Before(last) {
		last = t
	}
	elapsedTime := t.Sub(last)
	deltaTokens := elapsedTime.Seconds() * limiter.rps
	availableTokens := limiter.tokens + deltaTokens
	if burst := float64(limiter.burst); availableTokens > burst {
		availableTokens = burst
	}
	return availableTokens
}

// Allow shows whether event can happen at now
func (limiter *RpsLimiter) Allow() bool {
	if limiter.rps <= 0 || limiter.burst <= 0 {
		return false
	}
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	timeNow := time.Now()
	tokens := limiter.calculateTokens(timeNow)
	tokens -= 1
	if tokens < 0 {
		return false
	}
	// status of limiter updated only when this event allowed to be happened
	limiter.last = timeNow
	limiter.tokens = tokens
	return true
}

// TimeWindowLimiter expiredTime limiter
type TimeWindowLimiter struct {
	lock           sync.Mutex
	lastRequstTime int64
	expiredTime    time.Duration
	log            bool
}

// Allow TimeWindowLimiter request
func (twl *TimeWindowLimiter) Allow() bool {
	twl.lock.Lock()
	defer twl.lock.Unlock()
	if twl.lastRequstTime == 0 {
		twl.lastRequstTime = time.Now().UnixNano()
		return true
	}

	if time.Now().UnixNano() < twl.lastRequstTime+int64(twl.expiredTime) {
		return false
	}

	twl.lastRequstTime = time.Now().UnixNano()
	return true
}
