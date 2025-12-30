// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr for deal every handler
package handlermgr

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/checker/valuer"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindx/common/utils"
)

type certParaChecker struct {
	certChecker checker.ModelChecker
}

func newCertParaChecker() *certParaChecker {
	return &certParaChecker{}
}

func (icc *certParaChecker) init() {
	icc.certChecker.Checker = checker.GetAndChecker(
		checker.GetAndChecker(
			checker.GetIpV4Checker("Ip", true),
			&localIpChecker{},
		),
		checker.GetRegChecker("Port", "^[0-9]{1,5}$", true),
		checker.GetDomainChecker("Domain", true, true, true),
		getCaCertChecker("CaContent", checkCaCertContent, true),
	)
}

type localIpChecker struct{}

func (i *localIpChecker) Check(data interface{}) checker.CheckResult {
	strValuer := valuer.StringValuer{}
	value, err := strValuer.GetValue(data, "Ip")
	if err != nil {
		return checker.NewFailedResult(fmt.Sprintf("get string value failed, error: %v", err))
	}
	if utils.IsLocalIp(value) {
		return checker.NewFailedResult("ip cannot be loopBack address")
	}
	if err = utils.CheckInterfaceAddressIp(value); err != nil {
		return checker.NewFailedResult(fmt.Sprintf("checkInterfaceAddressIp failed, error: %v", err))
	}
	return checker.NewSuccessResult()
}

func (icc *certParaChecker) Check(data interface{}) checker.CheckResult {
	icc.init()
	checkResult := icc.certChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("cert checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

type caCertChecker struct {
	filed    string
	f        func(string) error
	required bool
	valuer   valuer.StringValuer
}

func checkCaCertContent(certContent string) error {
	caMgr, err := x509.NewCaChainMgr([]byte(certContent))
	if err != nil {
		hwlog.RunLog.Errorf("create and init ca chain Mgr failed: %s", err.Error())
		return errors.New("create and init ca chain Mgr failed")
	}

	if err = caMgr.CheckCertChain(); err != nil {
		hwlog.RunLog.Errorf("check importing certs failed: %s", err.Error())
		return errors.New("check importing certs failed")
	}

	return nil
}

func getCaCertChecker(field string, f func(string) error, required bool) *caCertChecker {
	return &caCertChecker{
		filed:    field,
		f:        f,
		required: required,
		valuer:   valuer.StringValuer{},
	}
}

// Check [method] for do ca cert check
func (cc *caCertChecker) Check(data interface{}) checker.CheckResult {
	targetString, err := cc.valuer.GetValue(data, cc.filed)
	if err != nil {
		if valuer.CheckIsFieldNotExistErr(err) && !cc.required {
			return checker.NewSuccessResult()
		}
		return checker.NewFailedResult(fmt.Sprintf("ca cert checker get field [%s] value failed:%v", cc.filed, err))
	}
	if err = cc.f(targetString); err != nil {
		return checker.NewFailedResult(fmt.Sprintf("ca cert checker check [%s] failed: %v",
			cc.filed, err))
	}
	return checker.NewSuccessResult()

}
