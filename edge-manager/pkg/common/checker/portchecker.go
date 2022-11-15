package checker

import (
	"fmt"
)

// IsPortInRange check port is in range or not
func IsPortInRange(minPort, maxPort, port int) (bool,error) {
	if port < minPort || port > maxPort {
		return false, fmt.Errorf("port %d is not in range [%d, %d]", port, minPort, maxPort)
	}
	return true, nil
}
