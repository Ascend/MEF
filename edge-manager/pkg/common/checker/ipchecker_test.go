package checker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsIpValid(t *testing.T) {
	ipTestCase := []struct{
		ip	string
		expResult bool
	} {
		{
			ip: "11.111.11.1",
			expResult: true,
		},
		{
			ip: "127.0.0.1",
			expResult: true,
		},
		{
			ip: "255.255.255.255",
			expResult: false,
		},
		{
			ip: "0.0.0.0",
			expResult: false,
		},
	}

	for index := range ipTestCase {
		valid, _ := IsIpValid(ipTestCase[index].ip)
		assert.Equal(t, valid, ipTestCase[index].expResult)
	}
}
