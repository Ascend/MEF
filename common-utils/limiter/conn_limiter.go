// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package limiter for connection limiter
package limiter

import (
	"fmt"
	"sync"
)

// ConnLimiter implementation for connection limiter interface
type ConnLimiter struct {
	mu           sync.Mutex
	connNum      int
	connLimitNum int
}

// NewConnLimiter initialize an instance of conn limiter
func NewConnLimiter(maxConnNum int) (*ConnLimiter, error) {
	if maxConnNum <= 0 {
		return nil, fmt.Errorf("invalid connection limit value: %v", maxConnNum)
	}
	return &ConnLimiter{
		connLimitNum: maxConnNum,
	}, nil
}

// ConnAdd if check passed, atomically adds one to counter , otherwise do nothing and return false
func (limiter *ConnLimiter) ConnAdd() bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if limiter.connNum < limiter.connLimitNum {
		limiter.connNum++
		return true
	}
	return false
}

// ConnDone atomically sub one from the counter
func (limiter *ConnLimiter) ConnDone() {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if limiter.connNum <= 0 {
		return
	}
	limiter.connNum--
}
