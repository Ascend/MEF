// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils string set operations
package utils

import (
	"sync"
)

// Set [struct] for a set
type Set struct {
	ele  map[string]struct{}
	lock sync.RWMutex
}

// NewSet new a set
func NewSet(items ...string) *Set {
	s := &Set{
		ele: make(map[string]struct{}, len(items)),
	}
	s.Add(items...)
	return s
}

// Add insert element to set
func (s *Set) Add(items ...string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, v := range items {
		s.ele[v] = struct{}{}
	}
}

// Delete delete element from set
func (s *Set) Delete(items ...string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, v := range items {
		delete(s.ele, v)
	}
}

// Find whether the set contain item or not
func (s *Set) Find(item string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if _, ok := s.ele[item]; ok {
		return true
	}
	return false
}

// List set element list
func (s *Set) List() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.listInner()
}

func (s *Set) listInner() []string {
	set := make([]string, 0, len(s.ele))
	for item := range s.ele {
		set = append(set, item)
	}
	return set
}

// Intersection sets intersection
func (s *Set) Intersection(set *Set) *Set {
	s.lock.RLock()
	defer s.lock.RUnlock()
	res := NewSet()
	if set == nil {
		return res
	}
	for e := range s.ele {
		if _, ok := set.ele[e]; ok {
			res.Add(e)
		}
	}
	return res
}

// Union sets union
func (s *Set) Union(set *Set) *Set {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if set == nil {
		return s
	}
	res := NewSet(s.listInner()...)
	for e := range set.ele {
		res.ele[e] = struct{}{}
	}
	return res
}

// Difference sets difference, s have, set have not
func (s *Set) Difference(set *Set) *Set {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if set == nil {
		return s
	}
	res := NewSet(s.listInner()...)
	for e := range set.ele {
		if _, ok := s.ele[e]; ok {
			delete(res.ele, e)
		}
	}
	return res
}
