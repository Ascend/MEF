// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package control

import (
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// ExchangeCaFlow is used to exchange root ca with north module
type ExchangeCaFlow struct {
	pathMgr        *util.InstallDirPathMgr
	importPath     string
	exportPath     string
	savePath       string
	saveBackupPath string
	certName       string
	uid            uint32
	gid            uint32
}

// NewExchangeCaFlow an ExchangeCaFlow struct
func NewExchangeCaFlow(importPath, exportPath string, pathMgr *util.InstallDirPathMgr,
	uid, gid uint32) *ExchangeCaFlow {
	savePath := pathMgr.ConfigPathMgr.GetNorthernCertPath()
	return &ExchangeCaFlow{
		pathMgr:        pathMgr,
		importPath:     importPath,
		exportPath:     exportPath,
		savePath:       savePath,
		saveBackupPath: savePath + backuputils.BackupSuffix,
		certName:       filepath.Base(savePath),
		uid:            uid,
		gid:            gid,
	}
}

// DoExchange is the main func to exchange certs
func (ecf *ExchangeCaFlow) DoExchange() error {
	var upgradeTasks = []func() error{
		ecf.checkParam,
		ecf.checkCa,
		ecf.exportCa,
		ecf.importCa,
	}

	for _, function := range upgradeTasks {
		if err := function(); err != nil {
			return err
		}
	}

	return nil
}

func (ecf *ExchangeCaFlow) checkParam() error {
	const maxCertSizeInMb = 1

	if ecf.importPath == ecf.exportPath {
		hwlog.RunLog.Error("import path cannot equal export path")
		return errors.New("import path cannot equal export path")
	}

	if _, err := utils.RealFileChecker(ecf.importPath, false, false, maxCertSizeInMb); err != nil {
		hwlog.RunLog.Errorf("importPath [%s] check failed: %s", ecf.importPath, err.Error())
		return errors.New("importPath check failed")
	}

	exportDir := filepath.Dir(ecf.exportPath)
	if _, err := utils.RealDirChecker(exportDir, true, false); err != nil {
		hwlog.RunLog.Errorf("exportPath [%s] check failed: %s", ecf.exportPath, err.Error())
		return errors.New("exportPath check failed")
	}

	if !utils.IsLexist(ecf.exportPath) {
		return nil
	}

	if _, err := utils.RealFileChecker(ecf.exportPath, false, false, maxCertSizeInMb); err != nil {
		hwlog.RunLog.Errorf("exportPath [%s] check failed: %s", ecf.exportPath, err.Error())
		return errors.New("exportPath check failed")
	}
	return nil
}

func (ecf *ExchangeCaFlow) checkCa() error {
	hwlog.RunLog.Infof("start to check [%s] cert", ecf.certName)

	if _, err := x509.CheckCertsChainReturnContent(ecf.importPath); err != nil {
		hwlog.RunLog.Errorf("check importing cert failed: %s", err.Error())
		return fmt.Errorf("check importing cert failed")
	}

	hash, err := utils.GetFileSha256(ecf.importPath)
	if err != nil {
		hwlog.RunLog.Errorf("get file sha256 sum failed: %s", err.Error())
		return errors.New("get file sha256 sum failed")
	}
	fmt.Printf("the sha256sum of the importing cert file is: %s\n", hash)
	hwlog.RunLog.Infof("the sha256sum of the importing cert file is: %s\n", hash)

	hwlog.RunLog.Infof("check [%s] cert success", ecf.certName)
	return nil
}

func (ecf *ExchangeCaFlow) importCa() error {
	if utils.IsLexist(ecf.savePath) {
		if err := common.DeleteFile(ecf.savePath); err != nil {
			hwlog.RunLog.Errorf("delete original crt [%s] failed: %s", ecf.savePath, err.Error())
			return errors.New("delete original crt failed")
		}
	}

	if err := ecf.copyCaToCertManager(); err != nil {
		return err
	}

	// delete old crl and backup of old crl
	crl := ecf.pathMgr.ConfigPathMgr.GetNorthernCrlPath()
	crlBackup := crl + backuputils.BackupSuffix
	for _, filePath := range []string{crl, crlBackup} {
		if err := utils.DeleteFile(filePath); err != nil {
			return fmt.Errorf("clear old crl failed, error: %s", err.Error())
		}
	}

	hwlog.RunLog.Infof("import [%s] cert success", ecf.certName)
	return nil
}

func (ecf *ExchangeCaFlow) copyCaToCertManager() error {
	if err := utils.MakeSureDir(ecf.savePath); err != nil {
		hwlog.RunLog.Errorf("create cert dst dir failed: %s", err.Error())
		return errors.New("create cert dst dir failed")
	}
	if err := utils.SetPathOwnerGroup(filepath.Dir(filepath.Dir(ecf.savePath)),
		ecf.uid, ecf.gid, false, false); err != nil {
		hwlog.RunLog.Errorf("set root-ca dir owner failed: %s", err.Error())
		return errors.New("set root-ca dir owner failed")
	}
	if err := utils.SetPathOwnerGroup(filepath.Dir(ecf.savePath),
		ecf.uid, ecf.gid, false, false); err != nil {
		hwlog.RunLog.Errorf("set crt dir owner failed: %s", err.Error())
		return errors.New("set crt dir owner failed")
	}

	if err := utils.CopyFile(ecf.importPath, ecf.savePath); err != nil {
		hwlog.RunLog.Errorf("copy temp crt to dst failed: %s", err.Error())
		return errors.New("copy temp crt to dst failed")
	}
	if err := backuputils.BackUpFiles(ecf.savePath); err != nil {
		hwlog.RunLog.Errorf("create backup of cert failed: %s", err.Error())
		return errors.New("create backup of cert failed")
	}

	if err := ecf.setDirOwnerAndPermission(ecf.savePath, ecf.saveBackupPath); err != nil {
		hwlog.RunLog.Errorf("set save crt right failed: %s", err.Error())
		if err = common.DeleteFile(ecf.savePath); err != nil {
			hwlog.RunLog.Warnf("delete crt [%s] failed: %s", ecf.certName, err.Error())
		}
		if err = common.DeleteFile(ecf.saveBackupPath); err != nil {
			hwlog.RunLog.Warnf("delete crt backup [%s] failed: %s", ecf.certName, err.Error())
		}
		return errors.New("set save crt right failed")
	}
	return nil
}

func (ecf *ExchangeCaFlow) exportCa() error {
	hwlog.RunLog.Info("start to export ca")

	srcPath := ecf.pathMgr.ConfigPathMgr.GetApigRootPath()
	srcBackupPath := srcPath + backuputils.BackupSuffix
	if !utils.IsExist(srcPath) {
		fmt.Println("the root ca has not yet generated, plz start cert manager first")
		hwlog.RunLog.Errorf("the root ca has not yet generated, plz start cert manager first")
		return errors.New(util.NotGenCertErrorStr)
	}

	if err := common.IsSoftLink(srcPath); err != nil {
		hwlog.RunLog.Errorf("check path [%s] failed: %s, cannot export", srcPath, err.Error())
		return fmt.Errorf("check path [%s] failed", srcPath)
	}

	if _, err := certutils.GetCertContentWithBackup(srcPath); err != nil {
		hwlog.RunLog.Errorf("check cert [%s] failed: %s, cannot export", srcPath, err.Error())
		return fmt.Errorf("check cert [%s] failed", srcPath)
	}
	if err := ecf.setDirOwnerAndPermission(srcPath, srcBackupPath); err != nil {
		hwlog.RunLog.Errorf("reset apig crt file right failed: %s", err.Error())
		return errors.New("reset apig crt file right failed")
	}

	if err := utils.CopyFile(srcPath, ecf.exportPath); err != nil {
		hwlog.RunLog.Errorf("export ca failed: %s", err.Error())
		return errors.New("export ca failed")
	}

	hwlog.RunLog.Info("export ca success")
	return nil
}

func (ecf *ExchangeCaFlow) setDirOwnerAndPermission(paths ...string) error {
	for _, path := range paths {
		if err := utils.SetPathOwnerGroup(path, ecf.uid, ecf.gid, true, false); err != nil {
			return err
		}
		if err := utils.SetPathPermission(path, common.Mode600, true, false); err != nil {
			return err
		}
	}
	return nil
}
