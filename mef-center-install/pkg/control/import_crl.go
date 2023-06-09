// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package control

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// ImportCrlFlow [struct] for import crl from north
type ImportCrlFlow struct {
	crlPath  string
	savePath string
	caPath   string
	uid      int
	gid      int
}

// NewImportCrlFlow [method] return a flow for import crl
func NewImportCrlFlow(crlPath, savePath, caPath string, uid, gid int) *ImportCrlFlow {
	return &ImportCrlFlow{
		crlPath:  crlPath,
		savePath: savePath,
		caPath:   caPath,
		uid:      uid,
		gid:      gid,
	}
}

// DoImportCrl [method] do actual jobs of flow
func (icf *ImportCrlFlow) DoImportCrl() error {
	var upgradeTasks = []func() error{
		icf.checkParam,
		icf.checkCrl,
		icf.importCrl,
	}

	for _, function := range upgradeTasks {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (icf *ImportCrlFlow) checkParam() error {
	const maxCrlSizeInMb = 10

	if !utils.IsExist(icf.caPath) {
		hwlog.RunLog.Error("import crl check failed: ca is not be imported")
		return errors.New("ca is not be imported")
	}

	if _, err := utils.RealFileChecker(icf.crlPath, false, false, maxCrlSizeInMb); err != nil {
		hwlog.RunLog.Errorf("crl path [%s] check failed: %s", icf.crlPath, err.Error())
		return errors.New("crl path check failed")
	}

	return nil
}

func (icf *ImportCrlFlow) checkCrl() error {
	hwlog.RunLog.Infof("start to check [%s] crl", icf.crlPath)
	if _, err := x509.CheckCrlsChainReturnContent(icf.crlPath, icf.caPath); err != nil {
		return fmt.Errorf("check crl failed, crl check error: %s", err.Error())
	}
	return nil
}

func (icf *ImportCrlFlow) importCrl() error {
	if utils.IsLexist(icf.savePath) {
		if err := common.DeleteFile(icf.savePath); err != nil {
			hwlog.RunLog.Errorf("delete original crl [%s] failed: %s", icf.savePath, err.Error())
			return errors.New("delete original crl failed")
		}
	}

	if err := utils.CopyFile(icf.crlPath, icf.savePath); err != nil {
		hwlog.RunLog.Errorf("copy temp crl to dst failed: %s", err.Error())
		return errors.New("copy temp crl to dst failed")
	}

	if err := util.SetPathOwnerGroup(icf.savePath, icf.uid, icf.gid, false, false); err != nil {
		hwlog.RunLog.Errorf("set crl owner failed: %s", err.Error())
		return errors.New("set crl owner failed")
	}

	if err := common.SetPathPermission(icf.savePath, common.Mode600, false, false); err != nil {
		hwlog.RunLog.Errorf("set save crl right failed: %s", err.Error())
		if err = common.DeleteFile(icf.savePath); err != nil {
			hwlog.RunLog.Warnf("delete crl [%s] failed: %s", filepath.Base(icf.savePath), err.Error())
		}
		return errors.New("set save crl right failed")
	}
	return nil
}
