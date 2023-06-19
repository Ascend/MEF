// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package certchecker

import (
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type importCrlChecker struct {
	crlChecker checker.ModelChecker
}

// NewImportCrlChecker [method] for getting issue cert checker struct
func NewImportCrlChecker() *importCrlChecker {
	return &importCrlChecker{
		crlChecker: checker.ModelChecker{
			Checker: checker.GetAndChecker(
				checker.GetStringChoiceChecker("CrlName", []string{common.NorthernCertName}, true),
				GetStringChecker("Crl", crlContentChecker, true),
			)},
	}
}

func crlContentChecker(crlContent string) error {
	lock.Lock()
	defer lock.Unlock()

	bytes, err := base64.StdEncoding.DecodeString(crlContent)
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
	if err = mgr.CheckCrl(filepath.Join(util.RootCaMgrDir, common.NorthernCertName, util.RootCaFileName)); err != nil {
		hwlog.RunLog.Errorf("check crl content failed, error: %s", err.Error())
		return err
	}
	return nil
}

func (icc *importCrlChecker) Check(data interface{}) checker.CheckResult {
	checkResult := icc.crlChecker.Check(data)
	if !checkResult.Result {
		return checker.NewFailedResult(fmt.Sprintf("crl checker check failed: %s", checkResult.Reason))
	}
	return checker.NewSuccessResult()
}
