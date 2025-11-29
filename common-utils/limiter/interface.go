// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
