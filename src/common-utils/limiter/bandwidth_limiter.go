// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package limiter
package limiter

import (
	"os"
	"time"
)

// ClientBandwidthLimiter the rps limiter for client side
type ClientBandwidthLimiter struct {
	bandwidthLimiterBase
}

// ServerBandwidthLimiter the rps limiter for server side
type ServerBandwidthLimiter struct {
	bandwidthLimiterBase
}

// BandwidthLimiterConfig rps limiter config
type BandwidthLimiterConfig struct {
	MaxThroughput int
	Period        time.Duration
	ReserveRate   float64
}

// Allow allows the traffic
func (t *ClientBandwidthLimiter) Allow(size int) bool {
	if t == nil || t.maxThroughout == 0 {
		return false
	}

	return t.allow("", size)
}

// Allow allows the traffic for conn
func (t *ServerBandwidthLimiter) Allow(conn string, size int) bool {
	if t == nil || t.maxThroughout == 0 {
		return false
	}

	return t.allow(conn, size)
}

// RegisterConn registers con
func (t *ServerBandwidthLimiter) RegisterConn(conn string) {
	if t == nil || t.maxThroughout == 0 {
		return
	}

	t.mu.Lock()
	if _, exist := t.privateThroughputStats[conn]; !exist {
		t.privateThroughputStats[conn] = &limiterStats{}
	}
	t.mu.Unlock()
}

// UnregisterConn unregisters con
func (t *ServerBandwidthLimiter) UnregisterConn(conn string) {
	if t == nil || t.maxThroughout == 0 {
		return
	}

	t.mu.Lock()
	delete(t.privateThroughputStats, conn)
	t.mu.Unlock()
}

// NewClientBandwidthLimiter creates a new bandwidth limiter for client
func NewClientBandwidthLimiter(cfg *BandwidthLimiterConfig) *ClientBandwidthLimiter {
	if cfg == nil {
		return nil
	}
	limiter := &ClientBandwidthLimiter{
		bandwidthLimiterBase: *newBandwidthLimiter(cfg.MaxThroughput, cfg.Period, 0, false),
	}

	go limiter.start()
	return limiter
}

// NewServerBandwidthLimiter creates a new bandwidth limiter for server
func NewServerBandwidthLimiter(cfg *BandwidthLimiterConfig) *ServerBandwidthLimiter {
	if cfg == nil || cfg.MaxThroughput == 0 || cfg.Period == 0 {
		return nil
	}
	limiter := &ServerBandwidthLimiter{
		bandwidthLimiterBase: *newBandwidthLimiter(cfg.MaxThroughput, cfg.Period, cfg.ReserveRate, true),
	}

	go limiter.start()
	return limiter
}

func newBandwidthLimiter(maxThroughput int, period time.Duration, reserveRate float64, isServer bool,
) *bandwidthLimiterBase {
	if period < statsPhases {
		period = defaultPeriod
	}

	limiter := &bandwidthLimiterBase{
		maxThroughout:          maxThroughput,
		period:                 period,
		reserveRate:            reserveRate,
		isServer:               isServer,
		privateThroughputStats: make(map[string]*limiterStats),
		quit:                   make(chan os.Signal),
	}
	return limiter
}
