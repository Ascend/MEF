// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package handlermgr for deal every handler
package handlermgr

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/constants"
)

const (
	minHostPort        = 1
	maxHostPort        = 65535
	imageSeparator     = ":"
	minImageAddressLen = 2
	dnsIpAddr          = 0
	portAddr           = 1
)

type certSdkParaChecker struct {
	certChecker checker.ModelChecker
}

func newCertSdkParaChecker() *certSdkParaChecker {
	return &certSdkParaChecker{}
}

func (idc *certSdkParaChecker) init() {
	idc.certChecker.Checker = checker.GetOrChecker(
		checker.GetAndChecker(
			checker.GetStringChoiceChecker("CertName",
				[]string{constants.ImageCertName, constants.SoftwareCertName}, true),
			checker.GetStringChoiceChecker("CertOpt",
				[]string{constants.OptUpdate}, true),
			getCaCertChecker("CertContent", checkCaContent, true),
			checker.GetOrChecker(
				checker.GetStringChoiceChecker("ImageAddress", []string{""}, true),
				getCaCertChecker("ImageAddress", checkImageAddress, true),
			),
		),
		checker.GetAndChecker(
			checker.GetStringChoiceChecker("CertName",
				[]string{constants.ImageCertName, constants.SoftwareCertName}, true),
			checker.GetStringChoiceChecker("CertOpt",
				[]string{constants.OptDelete}, true),
			checker.GetOrChecker(
				checker.GetStringChoiceChecker("ImageAddress", []string{""}, true),
				getCaCertChecker("ImageAddress", checkImageAddress, true),
			),
		),
	)
}

// Check [implement interface method] for certSdkParaChecker
func (idc *certSdkParaChecker) Check(data interface{}) checker.CheckResult {
	idc.init()
	checkResult := idc.certChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("cert checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}

func checkCaContent(certContent string) error {
	if len(certContent) == 0 {
		return nil
	}
	if len(certContent) > constants.CertSizeLimited {
		hwlog.RunLog.Error("verify ca file size failed")
		return errors.New("verify ca file size failed")
	}
	if err := x509.CheckPemCertChain([]byte(certContent)); err != nil {
		hwlog.RunLog.Errorf("verify ca file failed: %v", err)
		return errors.New("verify ca file failed")
	}
	return nil
}

func checkImageAddress(imagesAddress string) error {
	imagesAddressSplit := strings.Split(imagesAddress, imageSeparator)
	if len(imagesAddressSplit) != minImageAddressLen {
		hwlog.RunLog.Error("verify imagesAddress failed")
		return errors.New("verify imagesAddress failed")
	}

	portChecker := checker.GetIntChecker("", minHostPort, maxHostPort, true)
	imagesAddressPort, err := strconv.Atoi(imagesAddressSplit[portAddr])
	if err != nil {
		hwlog.RunLog.Errorf("strconv imagesAddress port failed, error:%v", err)
		return err
	}
	if portCheckerResult := portChecker.Check(imagesAddressPort); !portCheckerResult.Result {
		hwlog.RunLog.Errorf("verify port failed, error:%s", portCheckerResult.Reason)
		return errors.New(portCheckerResult.Reason)
	}

	imagesAddressDnsIpAddress := imagesAddressSplit[dnsIpAddr]
	ip := net.ParseIP(imagesAddressDnsIpAddress)
	if ip != nil {
		ipChecker := checker.GetIpV4Checker("", true)
		if ipCheckerResult := ipChecker.Check(imagesAddressDnsIpAddress); !ipCheckerResult.Result {
			hwlog.RunLog.Errorf("check ip failed, error: %s", ipCheckerResult.Reason)
			return errors.New(ipCheckerResult.Reason)
		}
		if utils.IsLocalIp(imagesAddressDnsIpAddress) {
			hwlog.RunLog.Error("check ip failed, cannot be loopBack address")
			return errors.New("check ip failed, cannot be loopBack address")
		}
		if err = utils.CheckInterfaceAddressIp(imagesAddressDnsIpAddress); err != nil {
			hwlog.RunLog.Errorf("check ip failed, error: %v", err)
			return fmt.Errorf("checkInterfaceAddressIp failed, error: %v", err)
		}
		return nil
	}

	if err := utils.CheckDomain(imagesAddressDnsIpAddress, true, true); err != nil {
		hwlog.RunLog.Errorf("check domain failed, error: %v", err)
		return errors.New("check domain failed")
	}
	return nil
}
