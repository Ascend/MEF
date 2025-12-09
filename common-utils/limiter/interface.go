// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package limiter, abstract interface definition for all kinds limiters
package limiter

import "context"

// ConnLimiterIntf abstract interface definition for connection limiter
type ConnLimiterIntf interface {
	ConnAdd() bool
	ConnDone()
}

// ContextAwareLimiter abstract interface definition for limiter contain context
type ContextAwareLimiter interface {
	Allow(ctx context.Context) bool
}

// IndependentLimiter abstract interface definition for websocket message limiter
type IndependentLimiter interface {
	Allow() bool
}

// ServerBandwidthLimiterIntf abstract interface definition for server side rps limiter
type ServerBandwidthLimiterIntf interface {
	Allow(string, int) bool
	RegisterConn(string)
	UnregisterConn(string)
	Stop()
}

// ClientBandwidthLimiterIntf abstract interface definition for client side rps limiter
type ClientBandwidthLimiterIntf interface {
	Allow(int) bool
	Stop()
}
