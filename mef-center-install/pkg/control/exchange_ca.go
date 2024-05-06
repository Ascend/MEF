// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package control

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// ExchangeCaFlow is used to exchange root ca with north module
type ExchangeCaFlow struct {
	pathMgr          *util.InstallDirPathMgr
	importPath       string
	exportPath       string
	savePath         string
	srcPath          string
	component        string
	uid              uint32
	gid              uint32
	printFingerprint bool
}

// NewExchangeCaFlow an ExchangeCaFlow struct
func NewExchangeCaFlow(importPath, exportPath, component string,
	pathMgr *util.InstallDirPathMgr) (*ExchangeCaFlow, error) {
	uid, gid, err := util.GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get MEF uid/gid failed: %s", err.Error())
		return nil, errors.New("get MEF uid/gid failed")
	}

	var (
		savePath, srcPath string
		printFingerprint  bool
	)
	switch component {
	case util.NginxManagerName:
		savePath = pathMgr.ConfigPathMgr.GetNorthernCertPath()
		srcPath = pathMgr.ConfigPathMgr.GetApigRootPath()
		printFingerprint = true
	default:
		return nil, errors.New("not support component to exchange ca")
	}

	return &ExchangeCaFlow{
		pathMgr:          pathMgr,
		importPath:       importPath,
		exportPath:       exportPath,
		component:        component,
		gid:              gid,
		uid:              uid,
		srcPath:          srcPath,
		savePath:         savePath,
		printFingerprint: printFingerprint,
	}, nil
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

	if _, err := fileutils.RealFileCheck(ecf.importPath, true, false, maxCertSizeInMb); err != nil {
		hwlog.RunLog.Errorf("importPath [%s] check failed: %s", ecf.importPath, err.Error())
		return errors.New("importPath check failed")
	}
	// forbid path traversal
	if strings.Contains(ecf.exportPath, "..") {
		return errors.New("the input path contains unsupported flag for parent directory")
	}
	exportDir := filepath.Dir(ecf.exportPath)
	if _, err := fileutils.RealDirCheck(exportDir, true, false); err != nil {
		hwlog.RunLog.Errorf("exportPath [%s] check failed: %s", ecf.exportPath, err.Error())
		return errors.New("exportPath check failed")
	}
	// exportPath should not be an existing file to avoid overwriting any sys file
	// should not be already existed file
	if fileutils.IsExist(ecf.exportPath) {
		hwlog.RunLog.Errorf("exportPath [%s] check failed, cannot overwrite existed file", ecf.exportPath)
		return fmt.Errorf("exportPath [%s] check failed, cannot overwrite existed file", ecf.exportPath)
	}
	return nil
}

func (ecf *ExchangeCaFlow) checkCa() error {
	hwlog.RunLog.Infof("start to check [%s] cert", ecf.component)

	if _, err := x509.CheckCertsChainReturnContent(ecf.importPath); err != nil {
		hwlog.RunLog.Errorf("check importing cert failed: %s", err.Error())
		return fmt.Errorf("check importing cert failed")
	}

	hash, err := fileutils.GetFileSha256(ecf.importPath)
	if err != nil {
		hwlog.RunLog.Errorf("get file sha256 sum failed: %s", err.Error())
		return errors.New("get file sha256 sum failed")
	}
	if ecf.printFingerprint {
		fmt.Printf("the sha256sum of the importing cert file is: %s\n", hash)
	}
	hwlog.RunLog.Infof("the sha256sum of the importing cert file is: %s\n", hash)

	hwlog.RunLog.Infof("check [%s] cert success", ecf.component)
	return nil
}

func (ecf *ExchangeCaFlow) importCa() error {
	if fileutils.IsLexist(ecf.savePath) {
		if err := fileutils.DeleteFile(ecf.savePath); err != nil {
			hwlog.RunLog.Errorf("delete original crt [%s] failed: %s", ecf.savePath, err.Error())
			return errors.New("delete original crt failed")
		}
	}
	if err := ecf.copyCaToCertManager(); err != nil {
		return err
	}
	if ecf.component == util.NginxManagerName {
		// delete old crl and backup of old crl
		crl := ecf.pathMgr.ConfigPathMgr.GetNorthernCrlPath()
		crlBackup := crl + backuputils.BackupSuffix
		for _, filePath := range []string{crl, crlBackup} {
			if err := fileutils.DeleteFile(filePath); err != nil {
				return fmt.Errorf("clear old crl failed, error: %s", err.Error())
			}
		}
	}
	hwlog.RunLog.Infof("import [%s] cert success", ecf.component)
	return nil
}

func (ecf *ExchangeCaFlow) copyCaToCertManager() error {
	if err := fileutils.MakeSureDir(ecf.savePath); err != nil {
		hwlog.RunLog.Errorf("create cert dst dir failed: %s", err.Error())
		return errors.New("create cert dst dir failed")
	}
	param := fileutils.SetOwnerParam{
		Path:       filepath.Dir(filepath.Dir(ecf.savePath)),
		Uid:        ecf.uid,
		Gid:        ecf.gid,
		Recursive:  false,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set root-ca dir owner failed: %s", err.Error())
		return errors.New("set root-ca dir owner failed")
	}
	param = fileutils.SetOwnerParam{
		Path:       filepath.Dir(ecf.savePath),
		Uid:        ecf.uid,
		Gid:        ecf.gid,
		Recursive:  false,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set crt dir owner failed: %s", err.Error())
		return errors.New("set crt dir owner failed")
	}
	if err := fileutils.CopyFile(ecf.importPath, ecf.savePath); err != nil {
		hwlog.RunLog.Errorf("copy temp crt to dst failed: %s", err.Error())
		return errors.New("copy temp crt to dst failed")
	}
	if err := backuputils.BackUpFiles(ecf.savePath); err != nil {
		hwlog.RunLog.Errorf("create backup of cert failed: %s", err.Error())
		return errors.New("create backup of cert failed")
	}

	if err := ecf.setDirOwnerAndPermission(common.Mode400, true, ecf.savePath); err != nil {
		hwlog.RunLog.Errorf("failed to set import root.crt permissions mode and owner,err:%s", err.Error())
		return fmt.Errorf("failed to set import root.crt permissions mode and owner,err:%s", err.Error())
	}

	if err := ecf.setBackupPathPermission(); err != nil {
		return err
	}

	return nil
}

func (ecf *ExchangeCaFlow) setBackupPathPermission() error {
	saveBackupPath := ecf.savePath + backuputils.BackupSuffix
	err := ecf.setDirOwnerAndPermission(common.Mode600, true, saveBackupPath)
	if err == nil {
		return nil
	}
	hwlog.RunLog.Errorf("set save crt right failed: %s", err.Error())

	if err = fileutils.DeleteFile(ecf.savePath); err != nil {
		hwlog.RunLog.Warnf("delete crt [%s] failed: %s", ecf.component, err.Error())
	}
	if err = fileutils.DeleteFile(saveBackupPath); err != nil {
		hwlog.RunLog.Warnf("delete crt backup [%s] failed: %s", ecf.component, err.Error())
	}

	return errors.New("set save crt right failed")
}

func (ecf *ExchangeCaFlow) exportCa() (err error) {
	hwlog.RunLog.Info("start to export ca")

	srcBackupPath := ecf.srcPath + backuputils.BackupSuffix
	if !fileutils.IsExist(ecf.srcPath) {
		fmt.Println("the root ca has not yet generated, please start cert manager first")
		hwlog.RunLog.Errorf("the root ca has not yet generated, please start cert manager first")
		return errors.New(util.NotGenCertErrorStr)
	}

	if err = fileutils.IsSoftLink(ecf.srcPath); err != nil {
		hwlog.RunLog.Errorf("check path [%s] failed: %s, cannot export", ecf.srcPath, err.Error())
		return fmt.Errorf("check path [%s] failed", ecf.srcPath)
	}

	if err = util.ReducePriv(); err != nil {
		hwlog.RunLog.Errorf("reduce euid/gid to MEFCenter failed: %s", err.Error())
		return errors.New("reduce priv failed")
	}

	defer func() {
		if resetErr := util.ResetPriv(); resetErr != nil {
			err = resetErr
			hwlog.RunLog.Errorf("reset euid/gid back to root failed: %s", err.Error())
		}
	}()

	if _, err = certutils.GetCertContentWithBackup(ecf.srcPath); err != nil {
		hwlog.RunLog.Errorf("check cert [%s] failed: %s, cannot export", ecf.srcPath, err.Error())
		return fmt.Errorf("check cert [%s] failed", ecf.srcPath)
	}
	if err = ecf.setDirOwnerAndPermission(common.Mode400, true, ecf.srcPath); err != nil {
		hwlog.RunLog.Errorf("reset apig crt file right failed: %s", err.Error())
		return errors.New("reset apig crt file right failed")
	}
	if err = ecf.setDirOwnerAndPermission(common.Mode600, true, srcBackupPath); err != nil {
		hwlog.RunLog.Errorf("reset apig crt file right failed: %s", err.Error())
		return errors.New("reset apig crt file right failed")
	}

	if err = util.ResetPriv(); err != nil {
		hwlog.RunLog.Errorf("reset euid/gid back to root failed: %s", err.Error())
		return errors.New("reset euid/gid back to root failed")
	}

	if err = fileutils.CopyFile(ecf.srcPath, ecf.exportPath); err != nil {
		hwlog.RunLog.Errorf("export ca failed: %s", err.Error())
		return errors.New("export ca failed")
	}

	hwlog.RunLog.Info("export ca success")
	return nil
}

func (ecf *ExchangeCaFlow) setDirOwnerAndPermission(mode os.FileMode, recursive bool, paths ...string) error {
	for _, path := range paths {
		ownerParam := fileutils.SetOwnerParam{
			Path:         path,
			Uid:          ecf.uid,
			Gid:          ecf.gid,
			Recursive:    recursive,
			IgnoreFile:   false,
			CheckerParam: nil,
		}
		if err := fileutils.SetPathOwnerGroup(ownerParam); err != nil {
			return err
		}
		if err := fileutils.SetPathPermission(path, mode, recursive, false); err != nil {
			return err
		}
	}
	return nil
}
