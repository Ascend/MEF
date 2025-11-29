// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package importcrl for import crl flow
package importcrl

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
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

// CertPair is a struct that contains the crl path and its peer cert path
type CertPair struct {
	crlPath  string
	certPath string
}

// CrlImportFlow is the flow to import crl
type CrlImportFlow struct {
	configPathMgr *pathmgr.ConfigPathMgr
	crlPath       string
	peer          string
	certPair      *CertPair
}

// NewCrlImportFlow is the func to create a CrlImportFlow instance
func NewCrlImportFlow(configPathMgr *pathmgr.ConfigPathMgr, crlPath, peer string) *CrlImportFlow {
	return &CrlImportFlow{
		configPathMgr: configPathMgr,
		crlPath:       crlPath,
		peer:          peer,
	}
}

func (cif *CrlImportFlow) getCertPairPath() (CertPair, error) {
	peerMap := map[string]CertPair{
		constants.MefCenterPeer: {
			certPath: cif.configPathMgr.GetHubSvrRootCertPath(),
			crlPath:  cif.configPathMgr.GetHubSvrCrlPath(),
		},
	}

	path, exists := peerMap[cif.peer]
	if !exists {
		return CertPair{}, errors.New("unsupported peer param")
	}

	return path, nil
}

// RunFlow is the func to start the import ctl flow
func (cif *CrlImportFlow) RunFlow() error {
	var tasks = []func() error{
		cif.setCertPair,
		cif.checkCrl,
		cif.crlImport,
	}

	for _, function := range tasks {
		if err := function(); err != nil {
			return err
		}
	}

	hwlog.RunLog.Info("import crl success")
	return nil
}

func (cif *CrlImportFlow) setCertPair() error {
	certPair, err := cif.getCertPairPath()
	if err != nil {
		return err
	}

	cif.certPair = &certPair
	return nil
}

func (cif *CrlImportFlow) checkCrl() error {
	if cif.certPair == nil {
		return errors.New("cert pair does not initialized")
	}

	if !fileutils.IsLexist(cif.certPair.certPath) {
		hwlog.RunLog.Errorf("peer %s's root cert has not imported yet, cannot import its crl", cif.peer)
		fmt.Printf("peer %s's root cert has not imported yet, cannot import its crl\n", cif.peer)
		return errors.New("peer's cert has not yet imported")
	}

	_, err := x509.CheckCrlsChainReturnContent(cif.crlPath, cif.certPair.certPath)
	if err != nil {
		return fmt.Errorf("check crl chain failed: %s", err.Error())
	}

	return nil
}

func (cif *CrlImportFlow) crlImport() error {
	if cif.certPair == nil {
		return errors.New("cert pair does not initialized")
	}

	tempPath := cif.configPathMgr.GetTempCrlPath()
	defer func() {
		if err := fileutils.DeleteAllFileWithConfusion(filepath.Dir(tempPath)); err != nil {
			hwlog.RunLog.Warnf("delete tempPath [%s] failed: %s", tempPath, err.Error())
		}
	}()

	if err := cif.copyCrlToTmp(tempPath); err != nil {
		return err
	}

	if err := cif.copyCrlToEdgeMain(tempPath); err != nil {
		return err
	}

	return nil
}

func (cif *CrlImportFlow) copyCrlToTmp(tmpPath string) error {
	tmpDir := filepath.Dir(tmpPath)

	if err := fileutils.CreateDir(tmpDir, constants.Mode755); err != nil {
		return fmt.Errorf("init tmp dir failed: %s", err.Error())
	}

	if err := fileutils.CopyFile(cif.crlPath, tmpPath); err != nil {
		return fmt.Errorf("copy file to tmp dir failed: %s", err.Error())
	}

	if err := fileutils.SetPathPermission(tmpPath, constants.Mode444, false, false); err != nil {
		return fmt.Errorf("set tmp dir path failed: %s", err.Error())
	}

	return nil
}

func (cif *CrlImportFlow) copyCrlToEdgeMain(tmpPath string) error {
	if err := fileutils.DeleteAllFileWithConfusion(cif.certPair.crlPath); err != nil {
		hwlog.RunLog.Errorf("delete original crl [%s] failed: %s", cif.certPair.crlPath, err.Error())
		return errors.New("delete original crl failed")
	}

	uid, err := envutils.GetUid(constants.EdgeUserName)
	if err != nil {
		return err
	}

	gid, err := envutils.GetGid(constants.EdgeUserGroup)
	if err != nil {
		return err
	}

	if _, err = envutils.RunCommandWithUser(constants.CpCmd, envutils.DefCmdTimeoutSec, uid, gid, tmpPath,
		cif.certPair.crlPath); err != nil {
		hwlog.RunLog.Errorf("copy temp crl to dst failed: %v", err)
		return errors.New("copy temp crl to dst failed")
	}

	if _, err := envutils.RunCommandWithUser(constants.ChmodCmd, envutils.DefCmdTimeoutSec, uid, gid,
		strconv.FormatInt(constants.Mode400, constants.Base8), cif.certPair.crlPath); err != nil {
		hwlog.RunLog.Errorf("set save crl right failed: %s", err.Error())
		if err = fileutils.DeleteFile(cif.certPair.crlPath); err != nil {
			hwlog.RunLog.Warnf("delete crl [%s] failed: %s", cif.certPair.crlPath, err.Error())
		}
		return errors.New("set save crl right failed")
	}

	if err = util.CreateBackupWithMefOwner(cif.certPair.crlPath); err != nil {
		hwlog.RunLog.Warnf("creat backup for crl failed, %v", err)
	}
	return nil
}
