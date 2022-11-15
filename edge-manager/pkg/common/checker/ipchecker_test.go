package checker

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsIpValid(t *testing.T) {
	ipTestCase := []struct {
		ip        string
		expResult bool
		errMsg    error
	}{
		{
			ip:        "11.111.11.1",
			expResult: true,
			errMsg:    nil,
		},
		{
			ip:        "127.0.0.1",
			expResult: true,
			errMsg:    nil,
		},
		{
			ip:        "255.255.255.255",
			expResult: false,
			errMsg:    fmt.Errorf("IP can't be a broadcast address"),
		},
		{
			ip:        "0.0.0.0",
			expResult: false,
			errMsg:    fmt.Errorf("IP can't be an all zeros address"),
		},
		{
			ip:        "::",
			expResult: false,
			errMsg:    fmt.Errorf("IP can't be an all zeros address"),
		},
		{
			ip:        "224.1.1.1",
			expResult: false,
			errMsg:    fmt.Errorf("IP can't be a multicast address"),
		},
	}

	for index := range ipTestCase {
		valid, err := IsIpValid(ipTestCase[index].ip)
		assert.Equal(t, valid, ipTestCase[index].expResult)
		assert.Equal(t, err, ipTestCase[index].errMsg)
	}
}
