// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker to test port checker
package checker

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
