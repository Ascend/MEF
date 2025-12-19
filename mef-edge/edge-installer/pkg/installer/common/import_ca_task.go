// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common for log path permissions manager
package common

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
)

// ImportCaTask is a struct to check and save a CA file
type ImportCaTask struct {
	importPath string
	savePath   string
	certName   string
	uid        uint32
	gid        uint32
}

// InitImportCaTask is the func to init an ImportCaTask struct
func InitImportCaTask(importPath, savePath, certName string, uid, gid uint32) *ImportCaTask {
	return &ImportCaTask{
		importPath: importPath,
		savePath:   savePath,
		certName:   certName,
		uid:        uid,
		gid:        gid,
	}
}

// RunTask is to start to task
func (ict *ImportCaTask) RunTask() error {
	var tasks = []func() error{
		ict.checkCa,
		ict.prepareSavePath,
		ict.importCa,
	}
	for _, function := range tasks {
		if err := function(); err != nil {
			return err
		}
	}

	hwlog.RunLog.Info("import ca success")
	return nil
}

func (ict *ImportCaTask) checkCa() error {
	hwlog.RunLog.Infof("start to check [%s] cert", ict.certName)

	if _, err := x509.CheckCertsChainReturnContent(ict.importPath); err != nil {
		hwlog.RunLog.Errorf("check importing cert failed: %s", err.Error())
		return fmt.Errorf("check importing cert failed")
	}

	hash, err := fileutils.GetFileSha256(ict.importPath)
	if err != nil {
		hwlog.RunLog.Errorf("get file sha256 sum failed: %s", err.Error())
		return errors.New("get file sha256 sum failed")
	}
	fmt.Printf("the sha256sum of the importing cert file is: %s\n", hash)
	hwlog.RunLog.Infof("the sha256sum of the importing cert file is: %s\n", hash)

	hwlog.RunLog.Infof("check [%s] cert success", ict.certName)
	return nil
}

func (ict *ImportCaTask) prepareSavePath() error {
	if fileutils.IsExist(ict.savePath) {
		return nil
	}

	if _, err := envutils.RunCommandWithUser(constants.MkdirCmd, envutils.DefCmdTimeoutSec, ict.uid, ict.gid,
		ict.savePath); err != nil {
		hwlog.RunLog.Errorf("create dir [%s] failed: %s", ict.savePath, err.Error())
		return fmt.Errorf("create dir [%s] failed", ict.savePath)
	}

	if _, err := envutils.RunCommandWithUser(constants.ChmodCmd, envutils.DefCmdTimeoutSec, ict.uid, ict.gid,
		strconv.FormatInt(constants.Mode700, constants.Base8), ict.savePath); err != nil {
		hwlog.RunLog.Errorf("set path [%s] right failed: %s", ict.savePath, err.Error())
		return fmt.Errorf("set path [%s] right failed", ict.savePath)
	}

	return nil
}

func (ict *ImportCaTask) importCa() error {
	configPathMgr, err := path.GetConfigPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("get config path manager failed, error: %v", err)
		return errors.New("get config path manager failed")
	}

	tempPath := configPathMgr.GetTempCertsDir()
	defer func() {
		if err = fileutils.DeleteAllFileWithConfusion(tempPath); err != nil {
			hwlog.RunLog.Warnf("delete tempPath [%s] failed: %s", tempPath, err.Error())
		}
	}()

	if err = ict.copyCaToTemp(tempPath); err != nil {
		return err
	}

	saveCrt := filepath.Join(ict.savePath, ict.certName)
	if fileutils.IsExist(saveCrt) {
		if err = fileutils.DeleteAllFileWithConfusion(saveCrt); err != nil {
			hwlog.RunLog.Errorf("delete original crt [%s] failed: %s", saveCrt, err.Error())
			return errors.New("delete original crt failed")
		}
	}

	if err = ict.copyCaToEdgeMain(tempPath); err != nil {
		return err
	}

	if err = util.CreateBackupWithMefOwner(saveCrt); err != nil {
		hwlog.RunLog.Warnf("creat backup for imported ca failed, %v", err)
	}

	hwlog.RunLog.Infof("import [%s] cert success", ict.certName)
	return nil
}

func (ict *ImportCaTask) copyCaToTemp(tempPath string) error {
	if !fileutils.IsExist(tempPath) {
		if err := fileutils.CreateDir(tempPath, constants.Mode700); err != nil {
			hwlog.RunLog.Errorf("create temp crt path failed: %v", err)
			return errors.New("create temp crt path failed")
		}

		if err := fileutils.SetPathPermission(tempPath, constants.Mode755, false, false); err != nil {
			hwlog.RunLog.Errorf("set temp dir right failed: %s", err.Error())
			return errors.New("set temp dir right failed")
		}
	}

	tempCrt := filepath.Join(tempPath, ict.certName)
	if err := fileutils.CopyFile(ict.importPath, tempCrt); err != nil {
		hwlog.RunLog.Errorf("import [%s] cert failed: %s", ict.certName, err.Error())
		return fmt.Errorf("import [%s] cert failed, error: %s", ict.certName, err.Error())
	}

	if err := fileutils.SetPathPermission(tempCrt, constants.Mode444, false, false); err != nil {
		hwlog.RunLog.Errorf("set temp crt right failed: %s", err.Error())
		return errors.New("set temp crt right failed")
	}

	return nil
}

func (ict *ImportCaTask) copyCaToEdgeMain(tempPath string) error {
	tempCrt := filepath.Join(tempPath, ict.certName)
	saveCrt := filepath.Join(ict.savePath, ict.certName)
	if _, err := envutils.RunCommandWithUser(constants.CpCmd, envutils.DefCmdTimeoutSec, ict.uid, ict.gid, tempCrt,
		saveCrt); err != nil {
		hwlog.RunLog.Errorf("copy temp crt to dst failed: %v", err)
		return errors.New("copy temp crt to dst failed")
	}

	if _, err := envutils.RunCommandWithUser(constants.ChmodCmd, envutils.DefCmdTimeoutSec, ict.uid, ict.gid,
		strconv.FormatInt(constants.Mode400, constants.Base8), saveCrt); err != nil {
		hwlog.RunLog.Errorf("set save crt right failed: %s", err.Error())
		if err = fileutils.DeleteFile(saveCrt); err != nil {
			hwlog.RunLog.Warnf("delete crt [%s] failed: %s", saveCrt, err.Error())
		}
		return errors.New("set save crt right failed")
	}
	return nil
}
