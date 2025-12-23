// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils offer the some utils for certificate handling
package utils

import (
	"context"
	"errors"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	dnsReqTimeoutForCheck = time.Second
	resolveNetwork        = "ip"
	domainReg             = "^[a-zA-Z0-9][a-zA-Z0-9.-]{1,61}[a-zA-Z0-9]$"
)

// ClientIP try to get the clientIP
func ClientIP(r *http.Request) string {
	// get forward ip fistly
	var ip string
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	forwardSlice := strings.Split(xForwardedFor, ",")
	if len(forwardSlice) >= 1 {
		if ip = strings.TrimSpace(forwardSlice[0]); ip != "" {
			return ip
		}
	}
	// try get ip from "X-Real-Ip"
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	var err error
	if ip, _, err = net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// IsLocalIp for check local ip
func IsLocalIp(ip string) bool {
	return net.ParseIP(ip).IsLoopback()
}

// CheckInterfaceAddressIp for check ip or url include interfaceAddressIp
func CheckInterfaceAddressIp(ip string) error {
	addrList, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}
	for _, addr := range addrList {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			return errors.New("get ipNet failed")
		}
		if strings.Contains(ip, ipNet.IP.String()) && !ipNet.IP.IsLoopback() {
			return errors.New("can't contain local ip")
		}
	}
	return nil
}

// CheckDomain check domain which by regex and blacklist
// Note 1: If parameter 'forLocalUsage' is true, which indicate url in this check act is used by the local, the checker
// returns error when any hostname in url is equivalent to localhost.
// Warning: this func may call DNS (configured in file /etc/resolv.conf) by UDP (port: 53).
// !! Make sure this net chain is added in Communication Matrix !!
// Note 2: When a new domain name is configured, the IP address corresponding to the domain name cannot be resolved, so
// the parsing error can be ignored. If the domain is used for configuration, the 'ignoreLookupIPErr' value can be true.
// If the domain is used for usage, the value can be false.
func CheckDomain(domain string, forLocalUsage bool, ignoreLookupIPErr bool) error {
	matched, err := regexp.MatchString(domainReg, domain)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("domain does not match allowed regex")
	}
	if !forLocalUsage {
		return nil
	}
	if IsDigitString(domain) {
		return errors.New("domain can not be all digits")
	}
	if strings.Contains(domain, "localhost") {
		return errors.New("domain can not contain localhost")
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), dnsReqTimeoutForCheck)
	defer cancelFunc()
	ips, err := net.DefaultResolver.LookupIP(ctx, resolveNetwork, domain)
	if err != nil {
		// When a new domain name is configured, the IP address corresponding to the domain name cannot be resolved, so
		// the parsing error can be ignored.
		if ignoreLookupIPErr {
			return nil
		}
		return errors.New("domain resolve failed")
	}
	for _, ip := range ips {
		if IsLocalIp(ip.String()) {
			return errors.New("domain is not allowed to be a loop back address")
		}
	}
	return nil
}
