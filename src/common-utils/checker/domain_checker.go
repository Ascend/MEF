// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package checker for domain checker
package checker

import (
	"fmt"

	"huawei.com/mindx/common/utils"

	"huawei.com/mindx/common/checker/valuer"
)

// DomainChecker [struct] of checker for domain
type DomainChecker struct {
	field             string
	required          bool
	valuer            valuer.StringValuer
	forLocalUsage     bool
	ignoreLookupIPErr bool
}

// GetDomainChecker [method] get domain checker
// Note 1: If parameter 'forLocalUsage' is true, which indicate domain in this check act is used by the local, the
// checker returns error when any domain is equivalent to localhost. If 'forLocalUsage' is false, this checker prevents
// a domain in string 'localhost'.
// Warning: If parameter 'forLocalUsage' is true, this checker may call DNS (configured in file /etc/resolv.conf) by UDP
// (port: 53), so make sure this net chain is added in Communication Matrix !!
// Note 2: When a new domain name is configured, the IP address corresponding to the domain name cannot be resolved, so
// the parsing error can be ignored. If the domain is used for configuration, the 'ignoreLookupIPErr' value can be true.
// If the domain is used for usage, the value can be false.
func GetDomainChecker(filed string, required bool, forLocalUsage bool, ignoreLookupIPErr bool) *DomainChecker {
	return &DomainChecker{
		field:             filed,
		required:          required,
		valuer:            valuer.StringValuer{},
		forLocalUsage:     forLocalUsage,
		ignoreLookupIPErr: ignoreLookupIPErr,
	}
}

// Check [method] actually do check job
func (dc *DomainChecker) Check(data interface{}) CheckResult {
	value, err := dc.valuer.GetValue(data, dc.field)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !dc.required {
			return NewSuccessResult()
		}
		return NewFailedResult(fmt.Sprintf("domain checker get field [%s] value failed: %v", dc.field, err))
	}
	return dc.checkDomainValid(value)
}

func (dc *DomainChecker) checkDomainValid(domain string) CheckResult {
	if err := utils.CheckDomain(domain, dc.forLocalUsage, dc.ignoreLookupIPErr); err != nil {
		return NewFailedResult(fmt.Sprintf("domain checker check [%s] failed: %v", dc.field, err))
	}
	return NewSuccessResult()
}
