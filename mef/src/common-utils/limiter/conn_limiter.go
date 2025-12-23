// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
