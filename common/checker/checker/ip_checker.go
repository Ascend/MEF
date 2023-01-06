// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"
	"net"

	"huawei.com/mindxedge/base/common/checker/valuer"
)

// IpChecker [struct] for ip checker
type IpChecker struct {
	field    string
	required bool
	valuer   valuer.StringValuer
}

// GetIpChecker [method] for get ip checker
func GetIpChecker(filed string, required bool) *IpChecker {
	return &IpChecker{
		field:    filed,
		required: required,
		valuer:   valuer.StringValuer{},
	}
}

// Check [method] for do ip checker
func (ic *IpChecker) Check(data interface{}) CheckResult {
	ipString, err := ic.valuer.GetValue(data, ic.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !ic.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("Int checker get ipString failed:%v", err))
	}
	return ic.isIpValid(ipString)
}

func (ic *IpChecker) isIpValid(ip string) CheckResult {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return NewFailedResult("ip parse failed")
	}
	if parsedIp.To4() == nil && parsedIp.To16() == nil {
		return NewFailedResult("IP must be a valid IP address")
	}
	if parsedIp.Equal(net.IPv4bcast) {
		return NewFailedResult("IP can't be a broadcast address")
	}
	if parsedIp.IsMulticast() {
		return NewFailedResult("IP can't be a multicast address")
	}
	if parsedIp.IsLinkLocalUnicast() {
		return NewFailedResult("IP can't be a link-local unicast address")
	}
	if parsedIp.IsUnspecified() {
		return NewFailedResult("IP can't be an all zeros address")
	}
	return NewSuccessResult()
}
