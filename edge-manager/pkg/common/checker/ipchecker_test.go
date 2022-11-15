package checker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsIpValid(t *testing.T) {
	testIp := "11.111.11.1"
	assert.True(t, IsIpValid(testIp))

	testIp = "127.0.0.1"
	assert.True(t, IsIpValid(testIp))

	testIp = "255.255.255.255"
	assert.False(t, IsIpValid(testIp))

	testIp = "0.0.0.0"
	assert.False(t, IsIpValid(testIp))
}
