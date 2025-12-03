// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package certchecker

import (
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/checker/valuer"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type importCrlChecker struct {
}

// NewImportCrlChecker [method] for getting crl validation struct
func NewImportCrlChecker() *importCrlChecker {
	return &importCrlChecker{}
}

// NewImportCrlNameChecker [method] for getting crl name validation struct
func NewImportCrlNameChecker() *checker.ModelChecker {
	return &checker.ModelChecker{
		Required: true,
		Checker: checker.GetStringChoiceChecker("",
			[]string{common.NorthernCertName, common.ImageCertName, common.SoftwareCertName}, true),
	}
}

func checkCrlContent(name, content string) error {
	lock.Lock()
	defer lock.Unlock()

	bytes, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		hwlog.RunLog.Errorf("base64 decode crl content failed, error: %s", err.Error())
		return err
	}
	if len(bytes) == 0 || len(bytes) > maxCertSize {
		hwlog.RunLog.Errorf("check crl file size failed, [%d] is out of limit [%d] ", len(bytes), maxCertSize)
		return errors.New("size of crl file is out of limit")
	}

	mgr, err := x509.NewCrlMgr(bytes)
	if err != nil {
		hwlog.RunLog.Errorf("check crl failed, new crl mgr error: %s", err.Error())
		return err
	}
	certData := x509.CertData{CertPath: filepath.Join(util.RootCaMgrDir, name, util.RootCaFileName)}
	if err = mgr.CheckCrl(certData); err != nil {
		hwlog.RunLog.Errorf("check crl content failed, error: %s", err.Error())
		return err
	}
	return nil
}

func (icc *importCrlChecker) Check(data interface{}) checker.CheckResult {
	var (
		strValuer  valuer.StringValuer
		crlName    string
		crlContent string
		err        error
	)
	if crlName, err = strValuer.GetValue(data, "CrlName"); err != nil {
		return checker.NewFailedResult("crl checker check failed: unable to get crl name")
	}
	if crlContent, err = strValuer.GetValue(data, "Crl"); err != nil {
		return checker.NewFailedResult("crl checker check failed: unable to get crl content")
	}

	if checkResult := NewImportCrlNameChecker().Check(crlName); !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("crl checker check name failed: %s", checkResult.Reason))
	}
	if err := checkCrlContent(crlName, crlContent); err != nil {
		return checker.NewFailedResult(fmt.Sprintf("crl checker check content failed: %s", err.Error()))
	}
	return checker.NewSuccessResult()
}
