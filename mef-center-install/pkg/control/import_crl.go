// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package control

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"
)

// ImportCrlFlow [struct] for import crl from north
type ImportCrlFlow struct {
	crlPath  string
	savePath string
	caPath   string
	uid      uint32
	gid      uint32
}

// NewImportCrlFlow [method] return a flow for import crl
func NewImportCrlFlow(crlPath, savePath, caPath string, uid, gid uint32) *ImportCrlFlow {
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
	const maxCrlSizeInMb = 1

	if !fileutils.IsExist(icf.caPath) {
		hwlog.RunLog.Error("import crl check failed: ca is not be imported")
		return errors.New("ca is not be imported")
	}

	if _, err := fileutils.RealFileCheck(icf.crlPath, false, false, maxCrlSizeInMb); err != nil {
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
	if fileutils.IsLexist(icf.savePath) {
		if err := fileutils.DeleteFile(icf.savePath); err != nil {
			hwlog.RunLog.Errorf("delete original crl [%s] failed: %s", icf.savePath, err.Error())
			return errors.New("delete original crl failed")
		}
	}

	if err := fileutils.CopyFile(icf.crlPath, icf.savePath); err != nil {
		hwlog.RunLog.Errorf("copy temp crl to dst failed: %s", err.Error())
		return errors.New("copy temp crl to dst failed")
	}
	ownerPram := fileutils.SetOwnerParam{
		Path:         icf.savePath,
		Uid:          icf.uid,
		Gid:          icf.gid,
		Recursive:    false,
		IgnoreFile:   false,
		CheckerParam: nil,
	}
	if err := fileutils.SetPathOwnerGroup(ownerPram); err != nil {
		hwlog.RunLog.Errorf("set crl owner failed: %s", err.Error())
		return errors.New("set crl owner failed")
	}

	if err := fileutils.SetPathPermission(icf.savePath, fileutils.Mode400, false, false); err != nil {
		hwlog.RunLog.Errorf("set save crl right failed: %s", err.Error())
		if err = fileutils.DeleteFile(icf.savePath); err != nil {
			hwlog.RunLog.Warnf("delete crl [%s] failed: %s", filepath.Base(icf.savePath), err.Error())
		}
		return errors.New("set save crl right failed")
	}

	if err := icf.backupCrlWithMEFUser(); err != nil {
		hwlog.RunLog.Warnf("create crl backup file failed: %s", err.Error())
	}

	return nil
}

func (icf *ImportCrlFlow) backupCrlWithMEFUser() error {
	if err := backuputils.BackUpFiles(icf.savePath); err != nil {
		return fmt.Errorf("back up crl failed: %s", err.Error())
	}
	ownerPram := fileutils.SetOwnerParam{
		Path: icf.savePath + backuputils.BackupSuffix,
		Uid:  icf.uid,
		Gid:  icf.gid,
	}
	if err := fileutils.SetPathOwnerGroup(ownerPram); err != nil {
		hwlog.RunLog.Errorf("set crl backup file owner failed: %s", err.Error())
		return errors.New("set crl backup file owner failed")
	}
	return nil
}
