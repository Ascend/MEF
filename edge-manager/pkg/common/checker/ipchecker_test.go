package checker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsIpValid(t *testing.T) {
	testIp := "11.111.11.1"
	valid, _ := IsIpValid(testIp)
	assert.True(t, valid)

	testIp = "127.0.0.1"
	valid, _ = IsIpValid(testIp)
	assert.True(t, valid)

	testIp = "255.255.255.255"
	valid, _ = IsIpValid(testIp)
	assert.False(t, valid)

	testIp = "0.0.0.0"
	valid, _ = IsIpValid(testIp)
	assert.False(t, valid)
}
