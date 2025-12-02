/* Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
*/

// Package common a series of common function
package common

import "sync/atomic"

// AtomicBool is an atomic Boolean.
type AtomicBool struct{ v uint32 }

// NewAtomicBool creates a AtomicBool.
func NewAtomicBool(initial bool) *AtomicBool {
	return &AtomicBool{v: boolToUint(initial)}
}

// Load atomically loads the Boolean.
func (b *AtomicBool) Load() bool {
	return atomic.LoadUint32(&b.v) == 1
}

// Store atomically stores the passed value.
func (b *AtomicBool) Store(new bool) {
	atomic.StoreUint32(&b.v, boolToUint(new))
}

func boolToUint(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
