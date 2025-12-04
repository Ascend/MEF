// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package checker to test port checker
package valid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testMinPort = 900
	testMaxPort = 2000
	inRange     = 1000
)

func TestPortInRange(t *testing.T) {
	port := inRange
	minPort := testMinPort
	maxPort := testMaxPort

	inRange := IsPortInRange(minPort, maxPort, port)
	assert.True(t, inRange)
}

func TestPortSmallerThanMinPort(t *testing.T) {
	port := testMinPort - 1
	minPort := testMinPort
	maxPort := testMaxPort

	inRange := IsPortInRange(minPort, maxPort, port)
	assert.False(t, inRange)
}

func TestPortBiggerThanMaxPort(t *testing.T) {
	port := testMaxPort + 1
	minPort := testMinPort
	maxPort := testMaxPort

	inRange := IsPortInRange(minPort, maxPort, port)
	assert.False(t, inRange)
}
