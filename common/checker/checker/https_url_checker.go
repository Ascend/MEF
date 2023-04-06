// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"fmt"
	"net/http"
	"strings"

	"huawei.com/mindxedge/base/common/checker/valuer"
)

// HttpsUrlChecker [struct] for url checker
type HttpsUrlChecker struct {
	field    string
	required bool
	valuer   valuer.StringValuer
}

// GetHttpsUrlChecker [method] for get url checker
func GetHttpsUrlChecker(filed string, required bool) *HttpsUrlChecker {
	return &HttpsUrlChecker{
		field:    filed,
		required: required,
		valuer:   valuer.StringValuer{},
	}
}

// Check [method] for do url check
func (hc *HttpsUrlChecker) Check(data interface{}) CheckResult {
	const (
		urlSegmentCount = 2
		urlMaxLength    = 512
	)
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

	if segments[0] != http.MethodPost && segments[0] != http.MethodGet {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: method invalid", hc.field))
	}
	if len(segments[1]) > urlMaxLength {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: url length up to limit",
			hc.field))
	}

	if !strings.HasPrefix(segments[1], "https") {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: in not https url", hc.field))
	}

	if strings.ContainsAny(segments[1], "\n!\\; $<>") {
		return NewFailedResult(fmt.Sprintf("https url checker Check [%s] failed: contain invalide char",
			hc.field))
	}

	return NewSuccessResult()
}
