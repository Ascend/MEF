// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package control

import (
	"crypto/rsa"
	"encoding/pem"
	"errors"
	"fmt"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

// ExchangeCaFlow is used to exchange root ca with north module
type ExchangeCaFlow struct {
	pathMgr    *util.InstallDirPathMgr
	importPath string
	exportPath string
	savePath   string
	certName   string
	uid        int
	gid        int
}

// NewExchangeCaFlow an ExchangeCaFlow struct
func NewExchangeCaFlow(importPath, exportPath string, pathMgr *util.InstallDirPathMgr,
	uid, gid int) *ExchangeCaFlow {
	savePath := pathMgr.ConfigPathMgr.GetNginxNorthernCertPath()
	return &ExchangeCaFlow{
		pathMgr:    pathMgr,
		importPath: importPath,
		exportPath: exportPath,
		savePath:   savePath,
		certName:   filepath.Base(savePath),
		uid:        uid,
		gid:        gid,
	}
}

// DoExchange is the main func to exchange certs
func (ecf *ExchangeCaFlow) DoExchange() error {
	var upgradeTasks = []func() error{
		ecf.checkParam,
		ecf.checkCa,
		ecf.importCa,
		ecf.exportCa,
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
	caBytes, err := utils.LoadFile(ecf.importPath)
	if err != nil {
		return fmt.Errorf("load content of cert [%s] failed, error: %v", ecf.importPath, err)
	}
	if caBytes == nil {
		return fmt.Errorf("the content of cert [%s] is empty", ecf.importPath)
	}

	block, _ := pem.Decode(caBytes)
	if block == nil {
		return errors.New("the cert can only be in pem format")
	}

	if err = x509.VerifyCaCert(caBytes, x509.InvalidNum); err != nil {
		return fmt.Errorf("check cert [%s] failed, error: %v", ecf.importPath, err)
	}

	caCrt, err := x509.LoadCertsFromPEM(caBytes)
	if err != nil {
		return fmt.Errorf("load cert [%s] failed, error: %v", ecf.importPath, err)
	}

	if signAlg := caCrt.SignatureAlgorithm.String(); signAlg != common.SignAlg {
		return fmt.Errorf("signature algorithm [%s] is not supported, only support %s", signAlg, common.SignAlg)
	}

	pubKey, ok := caCrt.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("the public key convert to rsa public key failed")
	}

	if pubKeyLen := pubKey.N.BitLen(); pubKeyLen < common.MinPubKeyLen {
		return fmt.Errorf("the length of public key %d less than %d", pubKeyLen, common.MinPubKeyLen)
	}

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

	if err := ecf.copyCaToNginx(); err != nil {
		return err
	}

	hwlog.RunLog.Infof("import [%s] cert success", ecf.certName)
	return nil
}

func (ecf *ExchangeCaFlow) copyCaToNginx() error {
	if err := utils.CopyFile(ecf.importPath, ecf.savePath); err != nil {
		hwlog.RunLog.Errorf("copy temp crt to dst failed: %s", err.Error())
		return errors.New("copy temp crt to dst failed")
	}

	if err := util.SetPathOwnerGroup(ecf.savePath, ecf.uid, ecf.gid, false, false); err != nil {
		hwlog.RunLog.Errorf("set crt owner failed: %s", err.Error())
		return errors.New("set crt owner failed")
	}

	if err := common.SetPathPermission(ecf.savePath, common.Mode600, false, false); err != nil {
		hwlog.RunLog.Errorf("set save crt right failed: %s", err.Error())
		if err = common.DeleteFile(ecf.savePath); err != nil {
			hwlog.RunLog.Warnf("delete crt [%s] failed: %s", ecf.certName, err.Error())
		}
		return errors.New("set save crt right failed")
	}
	return nil
}

func (ecf *ExchangeCaFlow) exportCa() error {
	hwlog.RunLog.Info("start to export ca")

	srcPath := ecf.pathMgr.ConfigPathMgr.GetApigRootPath()
	if !utils.IsExist(srcPath) {
		fmt.Println("the root ca has not yet generated, plz start cert manager first")
		hwlog.RunLog.Errorf("the root ca has not yet generated, plz start cert manager first")
		return errors.New(util.NotGenCertErrorStr)
	}

	if err := common.IsSoftLink(srcPath); err != nil {
		hwlog.RunLog.Errorf("check path [%s] failed: %s, cannot export", srcPath, err.Error())
		return fmt.Errorf("check path [%s] failed", srcPath)
	}

	if err := utils.CopyFile(srcPath, ecf.exportPath); err != nil {
		hwlog.RunLog.Errorf("export ca failed: %s", err.Error())
		return errors.New("export ca failed")
	}

	hwlog.RunLog.Info("export ca success")
	return nil
}
