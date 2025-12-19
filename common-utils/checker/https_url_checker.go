// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package checker

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"huawei.com/mindx/common/checker/valid"
	"huawei.com/mindx/common/checker/valuer"
	"huawei.com/mindx/common/utils"
)

const (
	urlSegmentCount = 2
	urlMaxLength    = 512
)

// HttpsUrlChecker [struct] for url checker
type HttpsUrlChecker struct {
	field         string
	required      bool
	valuer        valuer.StringValuer
	forLocalUsage bool
}

// GetHttpsUrlChecker [method] for get url checker.
// Note: If parameter 'forLocalUsage' is true, which indicate url in this check act is used by the local, the checker
// returns error when any hostname in url is equivalent to localhost.
// Warning: If parameter 'forLocalUsage' is true, this checker may call DNS (configured in file /etc/resolv.conf) by UDP
// (port: 53), so make sure this net chain is added in Communication Matrix !!
func GetHttpsUrlChecker(filed string, required bool, forLocalUsage bool) *HttpsUrlChecker {
	return &HttpsUrlChecker{
		field:         filed,
		required:      required,
		valuer:        valuer.StringValuer{},
		forLocalUsage: forLocalUsage,
	}
}

// Check [method] for do url check
func (hc *HttpsUrlChecker) Check(data interface{}) CheckResult {
	value, err := hc.valuer.GetValue(data, hc.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !hc.required {
			return NewSuccessResult()
		}

		return NewFailedResult(fmt.Sprintf("https url checker get field [%s] value failed:%v", hc.field, err))
	}

	if value == "" && !hc.required {
		return NewSuccessResult()
	}

	segments := strings.Split(value, " ")
	if len(segments) != urlSegmentCount {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: the value segment in not 2",
			hc.field))
	}

	if segments[0] != http.MethodGet {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: method invalid", hc.field))
	}
	if len(segments[1]) > urlMaxLength {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: url length up to limit",
			hc.field))
	}

	if !strings.HasPrefix(segments[1], "https://") {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: is not https url", hc.field))
	}

	if strings.ContainsAny(segments[1], "\n!\\|; $<>@`") {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: contain invalid char",
			hc.field))
	}
	httpsUrl, err := url.Parse(segments[1])
	if err != nil {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: parse url failed", hc.field))
	}
	if err = hc.checkHostNameValid(httpsUrl.Hostname()); err != nil {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: %v", hc.field, err))
	}
	return NewSuccessResult()
}

func (hc *HttpsUrlChecker) checkHostNameValid(hostname string) error {
	parsedIp := net.ParseIP(hostname)
	if parsedIp != nil {
		if _, err := valid.IsIpValid(hostname); err != nil {
			return err
		}
		if !hc.forLocalUsage {
			return nil
		}
		if utils.IsLocalIp(hostname) {
			return errors.New("url IP can't be loopBack address")
		}
		return utils.CheckInterfaceAddressIp(hostname)
	}
	return utils.CheckDomain(hostname, hc.forLocalUsage, false)
}
