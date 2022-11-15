package checker

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPortInRange(t *testing.T) {
	port := 1000
	minPort := 900
	maxPort := 2000

	inRange, err := IsPortInRange(minPort, maxPort, port)
	assert.True(t, inRange)
	assert.Nil(t, err)
}

func TestPortSmallerThanMinPort(t *testing.T) {
	port := 800
	minPort := 900
	maxPort := 2000

	inRange, err := IsPortInRange(minPort, maxPort, port)
	assert.False(t, inRange)
	assert.Errorf(t, fmt.Errorf("port %d is not in range [%d, %d]", port ,minPort, maxPort), err.Error())
}

func TestPortBiggerThanMaxPort(t *testing.T) {
	port := 3000
	minPort := 900
	maxPort := 2000

	inRange, err := IsPortInRange(minPort, maxPort, port)
	assert.False(t, inRange)
	assert.Errorf(t, fmt.Errorf("port %d is not in range [%d, %d]", port ,minPort, maxPort), err.Error())
}
