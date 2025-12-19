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
	"sync"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
)

const (
	statsPhases = 10
	// the default GC period of golang
	defaultPeriod = 2 * time.Minute
)

// bandwidthLimiterBase limits rps
type bandwidthLimiterBase struct {
	maxThroughout int
	period        time.Duration
	// the percentage of exclusive throughput
	reserveRate float64
	isServer    bool

	phase int
	// for all connections
	sharedThroughoutStats limiterStats
	// for one conn
	privateThroughputStats map[string]*limiterStats

	quit     chan os.Signal
	stopOnce sync.Once
	mu       sync.Mutex
}

// Stop stops the limiter
func (t *bandwidthLimiterBase) Stop() {
	if t == nil {
		return
	}

	t.stopOnce.Do(func() {
		close(t.quit)
	})
}

// checks whether limiter can accept and record the throughput
func (t *bandwidthLimiterBase) allow(conn string, size int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if size < 0 {
		return false
	}

	var availablePrivateThroughput int
	if t.isServer {
		stats, ok := t.privateThroughputStats[conn]
		if !ok {
			return false
		}
		if len(t.privateThroughputStats) <= 0 || stats == nil {
			hwlog.RunLog.Error("stats is nil or length of stats is non-positive")
			return false
		}
		availablePrivateThroughput = int(float64(t.maxThroughout)*t.reserveRate) / len(t.privateThroughputStats)
		availablePrivateThroughput -= stats.sum()
	}

	availableSharedThroughput := int(float64(t.maxThroughout) * (1 - t.reserveRate))
	availableSharedThroughput -= t.sharedThroughoutStats.sum()

	if availablePrivateThroughput+availableSharedThroughput < size {
		return false
	}

	if t.isServer {
		// available private throughput could be negative
		throughput := utils.MinInt(availablePrivateThroughput, size, 0)
		t.privateThroughputStats[conn].add(t.phase, throughput)
		size -= throughput
	}
	t.sharedThroughoutStats.add(t.phase, size)
	return true
}

func (t *bandwidthLimiterBase) start() {
	ticker := time.NewTicker(t.period / statsPhases)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-ticker.C:
			if !ok {
				return
			}
			t.roll()
		case <-t.quit:
			return
		}
	}
}

// roll clears the expired throughput and update the phase
func (t *bandwidthLimiterBase) roll() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.phase++
	if t.phase >= statsPhases {
		t.phase = 0
	}
	t.sharedThroughoutStats.reset(t.phase)
	for conn := range t.privateThroughputStats {
		t.privateThroughputStats[conn].reset(t.phase)
	}
}

type limiterStats [statsPhases]int

func (s *limiterStats) reset(phase int) {
	if phase < 0 || phase >= statsPhases {
		hwlog.RunLog.Errorf("reset: index of out of bound %d", phase)
		return
	}
	s[phase] = 0
}

func (s *limiterStats) add(phase, delta int) {
	if phase < 0 || phase >= statsPhases {
		hwlog.RunLog.Errorf("add: index of out of bound %d", phase)
		return
	}
	s[phase] += delta
}

func (s limiterStats) sum() int {
	var total int
	for _, throughput := range s {
		total += throughput
	}
	return total
}
