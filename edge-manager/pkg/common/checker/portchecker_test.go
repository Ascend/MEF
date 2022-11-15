// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package checker to test port checker
package checker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPortInRange(t *testing.T) {
	port := 1000
	minPort := 900
	maxPort := 2000

	inRange := IsPortInRange(minPort, maxPort, port)
	assert.True(t, inRange)
}

func TestPortSmallerThanMinPort(t *testing.T) {
	port := 800
	minPort := 900
	maxPort := 2000

	inRange := IsPortInRange(minPort, maxPort, port)
	assert.False(t, inRange)
}

func TestPortBiggerThanMaxPort(t *testing.T) {
	port := 3000
	minPort := 900
	maxPort := 2000

	inRange := IsPortInRange(minPort, maxPort, port)
	assert.False(t, inRange)
}
