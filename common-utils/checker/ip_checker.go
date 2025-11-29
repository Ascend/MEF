// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"
	"net"

	"huawei.com/mindx/common/checker/valuer"
)

type ipChecker struct {
	field    string
	required bool
	valuer   valuer.StringValuer
}

// Ipv4Checker [struct] of checker for ipv4
type Ipv4Checker struct {
	ipChecker
}

// GetIpV4Checker [method] get ipv4 checker
func GetIpV4Checker(filed string, required bool) *Ipv4Checker {
	return &Ipv4Checker{
		ipChecker: ipChecker{
			field:    filed,
			required: required,
			valuer:   valuer.StringValuer{},
		},
	}
}

// Check [method] actually do check job
func (ic *Ipv4Checker) Check(data interface{}) CheckResult {
	ipString, err := ic.valuer.GetValue(data, ic.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !ic.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("Int checker get ipString failed:%v", err))
	}
	return ic.isIpValid(ipString)
}

func (ic *Ipv4Checker) isIpValid(ip string) CheckResult {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return NewFailedResult("IP parse failed")
	}
	if parsedIp.To4() == nil {
		return NewFailedResult("IP is not a valid IPv4 address")
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
