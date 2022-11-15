package checker

import "huawei.com/mindx/common/hwlog"

// IsPortInRange check port is in range or not
func IsPortInRange(minPort, maxPort, port int) bool {
	if port < minPort || port > maxPort {
		hwlog.RunLog.Errorf("port %d is not in range [%d, %d]", port, minPort, maxPort)
		return false
	}
	return true
}
