package checker

import (
	"huawei.com/mindx/common/hwlog"
	"net"
)

// IsIpValid check ip is valid
func IsIpValid(ip string) bool {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		hwlog.RunLog.Errorf("ip parse failed")
		return false
	}
	if parsedIp.To4() == nil && parsedIp.To16() == nil {
		hwlog.RunLog.Errorf("IP must be a valid IP address")
		return false
	}
	if parsedIp.IsMulticast() {
		hwlog.RunLog.Errorf("IP can't be a multicast address")
		return false
	}
	if parsedIp.IsLinkLocalUnicast() {
		hwlog.RunLog.Errorf("IP can't be a link-local unicast address")
		return false
	}
	if parsedIp.IsUnspecified() {
		hwlog.RunLog.Errorf("IP can't be an all zeros address")
		return false
	}
	return true
}

// IsIpInHost check whether the IP address is on the host
func IsIpInHost(ip string) bool {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		hwlog.RunLog.Errorf("ip parse failed")
		return false
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		hwlog.RunLog.Errorf("get host ip list fail")
		return false
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil && ip.Equal(parsedIp) {
			return true
		}
	}
	hwlog.RunLog.Errorf("ip %s not found in the host's network interfaces", ip)
	return false
}
