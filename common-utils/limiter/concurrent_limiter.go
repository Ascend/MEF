// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package limiter for concurrent limiter
package limiter

import (
	"context"
	"time"

	"huawei.com/mindx/common/hwlog"
)

// ConcurrentLimiter concurrent limiter struct
type ConcurrentLimiter struct {
	concurrency chan struct{}
	overdueTime time.Duration
}

// Init for limiter token
func (cl *ConcurrentLimiter) Init() {
	for i := 0; i < cap(cl.concurrency); i++ {
		cl.concurrency <- struct{}{}
	}
}

// Allow concurrent limiter request
func (cl *ConcurrentLimiter) Allow(ctx context.Context) bool {
	select {
	case _, ok := <-cl.concurrency:
		if !ok {
			//  channel closed and no need return token
			return false
		}
		go cl.returnToken(ctx)
		return true
	default:
		return false
	}
}

func (cl *ConcurrentLimiter) returnToken(ctx context.Context) {
	defer func() {
		cl.concurrency <- struct{}{}
		if err := recover(); err != nil {
			hwlog.RunLog.Errorf("go routine failed with %#v", err)
		}
	}()
	if cl.concurrency == nil {
		hwlog.RunLog.Error("Limiter meet error")
		return
	}
	timer := time.NewTimer(cl.overdueTime)
	defer timer.Stop()
	select {
	case _, ok := <-timer.C:
		if !ok {
			return
		}
		hwlog.RunLog.Debugf("recover token num：%d", len(cl.concurrency))
		return
	case _, ok := <-ctx.Done():
		err := ctx.Err()
		if !ok || err != nil {
			hwlog.RunLog.Debugf("%+v:%+v", err, ok)
		}
		return
	}
}
