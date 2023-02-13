// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package install this package is for handle mef center install
package install

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/certutils"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type certPrepareCtl struct {
	certPathMgr *util.ConfigPathMgr
	components  []string
}

func (cpc *certPrepareCtl) doPrepare() error {
	var prepareCertsTasks = []func() error{
		cpc.prepareCertsDir,
		cpc.prepareCerts,
	}

	fmt.Println("start to prepare certs")
	for _, function := range prepareCertsTasks {
		if err := function(); err != nil {
			return err
		}
	}
	fmt.Println("prepare certs success")
	return nil
}

func (cpc *certPrepareCtl) prepareCertsDir() error {
	hwlog.RunLog.Info("start to prepare component certs directories")
	if cpc.certPathMgr == nil {
		hwlog.RunLog.Error("pointer cpc.certPathMgr is nil")
		return errors.New("pointer cpc.certPathMgr is nil")
	}

	certPath := cpc.certPathMgr.GetConfigPath()
	if err := common.MakeSurePath(certPath); err != nil {
		hwlog.RunLog.Errorf("create cert path [%s] failed: %v", certPath, err.Error())
		return errors.New("create cert path failed")
	}

	rootCertPath := cpc.certPathMgr.GetRootCaCertDirPath()
	if err := common.MakeSurePath(rootCertPath); err != nil {
		hwlog.RunLog.Errorf("create root certs path [%s] failed: %v", certPath, err.Error())
		return errors.New("create root certs path failed")
	}

	rootKeyPath := cpc.certPathMgr.GetRootCaKeyDirPath()
	if err := common.MakeSurePath(rootKeyPath); err != nil {
		hwlog.RunLog.Errorf("create root key path [%s] failed: %v", certPath, err.Error())
		return errors.New("create root key path failed")
	}

	// prepare component's certs directory
	for _, component := range cpc.components {
		componentMgr := util.GetComponentMgr(component)
		if err := componentMgr.PrepareComponentCertDir(certPath); err != nil {
			hwlog.RunLog.Errorf("prepare component [%s]'s cert dir failed: %v", component, err.Error())
			return fmt.Errorf("prepare component [%s]'s cert dir failed", component)
		}
	}
	hwlog.RunLog.Info("prepare component certs directories successful")
	return nil
}

func (cpc *certPrepareCtl) prepareCerts() error {
	hwlog.RunLog.Info("start to prepare certs")

	var (
		err         error
		rootCertMgr *certutils.RootCertMgr
	)

	if rootCertMgr, err = cpc.prepareCA(); err != nil {
		return err
	}

	for _, component := range cpc.components {
		componentMgr := util.GetComponentMgr(component)
		if err = componentMgr.PrepareComponentCert(rootCertMgr, cpc.certPathMgr); err != nil {
			hwlog.RunLog.Errorf("prepare %s component cert failed: %s", component, err.Error())
			return errors.New("prepare single component cert failed")
		}
	}

	if err = cpc.setCertsOwner(); err != nil {
		return err
	}

	hwlog.RunLog.Info("prepare certs successful")

	return nil
}

func (cpc *certPrepareCtl) setCertsOwner() error {
	var err error
	if cpc.certPathMgr == nil {
		hwlog.RunLog.Error("pointer cpc.certPathMgr is nil")
		return errors.New("pointer cpc.certPathMgr is nil")
	}

	mefUid, mefGid, err := util.GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get mef uid or gid failed: %s", err.Error())
		return errors.New("get mef uid or gid failed")
	}

	if err = util.SetPathOwnerGroup(cpc.certPathMgr.GetConfigPath(), mefUid,
		mefGid, true, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			cpc.certPathMgr.GetConfigPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}

	if err = util.SetPathOwnerGroup(cpc.certPathMgr.GetConfigPath(), util.RootUid,
		util.RootGid, false, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			cpc.certPathMgr.GetConfigPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}

	if err = util.SetPathOwnerGroup(cpc.certPathMgr.GetRootKmcDirPath(), util.RootUid,
		util.RootGid, true, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			cpc.certPathMgr.GetRootKmcDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}

	if err = util.SetPathOwnerGroup(cpc.certPathMgr.GetRootCaKeyDirPath(), util.RootUid,
		util.RootGid, true, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			cpc.certPathMgr.GetRootCaKeyDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}

	if err = util.SetPathOwnerGroup(cpc.certPathMgr.GetRootCaDirPath(), util.RootUid,
		util.RootGid, false, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			cpc.certPathMgr.GetRootCaDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	return nil
}

func (cpc *certPrepareCtl) prepareCA() (*certutils.RootCertMgr, error) {
	if cpc.certPathMgr == nil {
		hwlog.RunLog.Error("pointer cpc.certPathMgr is nil")
		return nil, errors.New("pointer cpc.certPathMgr is nil")
	}

	rootCaFilePath := cpc.certPathMgr.GetRootCaCertPath()
	rootPrivFilePath := cpc.certPathMgr.GetRootCaKeyPath()
	kmcKeyPath := cpc.certPathMgr.GetRootMasterKmcPath()
	kmcBackKeyPath := cpc.certPathMgr.GetRootBackKmcPath()

	rootKmcCfg := common.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)
	initCertMgr := certutils.InitRootCertMgr(rootCaFilePath, rootPrivFilePath, util.CaCommonName, rootKmcCfg)
	if _, err := initCertMgr.NewRootCa(); err != nil {
		hwlog.RunLog.Errorf("init root ca info failed: %v", err)
		return nil, errors.New("init root ca info failed")
	}
	return initCertMgr, nil
}
