// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
	components  map[string]*util.InstallComponent
}

func (cpc *certPrepareCtl) doPrepare() error {
	var prepareCertsTasks = []func() error{
		cpc.prepareCertsDir,
		cpc.prepareCerts,
	}

	for _, function := range prepareCertsTasks {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (cpc *certPrepareCtl) prepareCertsDir() error {
	hwlog.RunLog.Info("start to prepare component certs directories")
	if cpc.certPathMgr == nil {
		hwlog.RunLog.Error("pointer cpc.certPathMgr is nil")
		return errors.New("pointer cpc.certPathMgr is nil")
	}
	if cpc.components == nil {
		hwlog.RunLog.Error("pointer cpc.components is nil")
		return errors.New("pointer cpc.components is nil")
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
		if err := component.PrepareComponentCertDir(certPath); err != nil {
			hwlog.RunLog.Errorf("prepare component [%s]'s cert dir failed: %v", component.Name, err.Error())
			return fmt.Errorf("prepare component [%s]'s cert dir failed", component.Name)
		}
	}
	hwlog.RunLog.Info("prepare component certs directories successful")
	return nil
}

func (cpc *certPrepareCtl) prepareCerts() error {
	hwlog.RunLog.Info("start to prepare certs")
	if cpc.components == nil {
		hwlog.RunLog.Error("pointer cpc.components is nil")
		return errors.New("pointer cpc.components is nil")
	}

	var (
		err     error
		certMng *certutils.RootCertMgr
	)

	if err = cpc.prepareCA(); err != nil {
		return err
	}

	for _, component := range cpc.components {
		certMng, err = cpc.getComponentCertMgr(component.Name)
		if err != nil {
			hwlog.RunLog.Errorf("init %s component cert mgr failed: %s", component.Name, err.Error())
			return errors.New("init single component cert mgr failed")
		}
		if err = component.PrepareComponentCert(certMng, cpc.certPathMgr); err != nil {
			hwlog.RunLog.Errorf("prepare %s component cert failed: %s", component.Name, err.Error())
			return errors.New("prepare single component cert failed")
		}
	}

	if err = cpc.setCertsOwner(); err != nil {
		return nil
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

	if err = util.SetPathOwnerGroup(cpc.certPathMgr.GetRootKmcDirPath(), util.RootUid,
		util.RootGid, true, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			cpc.certPathMgr.GetRootKmcDirPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}

	if err = util.SetPathOwnerGroup(cpc.certPathMgr.GetRootCaKeyPath(), util.RootUid,
		util.RootGid, false, false); err != nil {
		hwlog.RunLog.Errorf("set path [%s] owner and group failed: %v",
			cpc.certPathMgr.GetRootCaKeyPath(), err.Error())
		return errors.New("set cert root path owner and group failed")
	}
	return nil
}

func (cpc *certPrepareCtl) prepareCA() error {
	if cpc.certPathMgr == nil {
		hwlog.RunLog.Error("pointer cpc.certPathMgr is nil")
		return errors.New("pointer cpc.certPathMgr is nil")
	}

	rootCaFilePath := cpc.certPathMgr.GetRootCaCertPath()
	rootPrivFilePath := cpc.certPathMgr.GetRootCaKeyPath()
	rootKmcCfg := common.KmcCfg{
		SdpAlgID:       common.Aes256gcm,
		PrimaryKeyPath: cpc.certPathMgr.GetRootMasterKmcPath(),
		StandbyKeyPath: cpc.certPathMgr.GetRootBackKmcPath(),
		DoMainId:       common.DoMainId,
	}
	initCertMgr := certutils.InitRootCertMgr(rootCaFilePath, rootPrivFilePath, util.CaCommonName, &rootKmcCfg)
	if _, err := initCertMgr.NewRootCa(); err != nil {
		hwlog.RunLog.Errorf("init root ca info failed: %v", err)
		return errors.New("init root ca info failed")
	}
	return nil
}

func (cpc *certPrepareCtl) getComponentCertMgr(component string) (*certutils.RootCertMgr, error) {
	if cpc.certPathMgr == nil {
		hwlog.RunLog.Error("pointer cpc.certPathMgr is nil")
		return nil, errors.New("pointer cpc.certPathMgr is nil")
	}

	rootCaFilePath := cpc.certPathMgr.GetRootCaCertPath()
	rootPrivFilePath := cpc.certPathMgr.GetRootCaKeyPath()
	componentKmcCfg := common.KmcCfg{
		SdpAlgID:       common.Aes256gcm,
		PrimaryKeyPath: cpc.certPathMgr.GetComponentMasterKmcPath(component),
		StandbyKeyPath: cpc.certPathMgr.GetComponentBackKmcPath(component),
		DoMainId:       common.DoMainId,
	}
	CertMgr := certutils.InitRootCertMgr(rootCaFilePath, rootPrivFilePath, util.CaCommonName, &componentKmcCfg)
	return CertMgr, nil
}
