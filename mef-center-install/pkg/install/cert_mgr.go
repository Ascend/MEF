// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package install this package is for handle mef center install
package install

import (
	"context"
	"errors"
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
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
	rootCertPath := cpc.certPathMgr.GetRootCaCertDirPath()
	if err := common.MakeSurePath(rootCertPath); err != nil {
		hwlog.RunLog.Errorf("create root certs path [%s] failed: %v", rootCertPath, err.Error())
		return errors.New("create root certs path failed")
	}

	rootKeyPath := cpc.certPathMgr.GetRootCaKeyDirPath()
	if err := common.MakeSurePath(rootKeyPath); err != nil {
		hwlog.RunLog.Errorf("create root key path [%s] failed: %v", rootKeyPath, err.Error())
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

	ctx, cancel := context.WithTimeout(context.Background(), common.DefCmdTimeoutSec*time.Second)
	defer cancel()
	ch := make(chan error)
	go cpc.doPrepareCerts(ch)

	select {
	case err := <-ch:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		hwlog.RunLog.Errorf("generate certs timeout!")
		return errors.New("generate certs timeout")
	}

	hwlog.RunLog.Info("prepare certs successful")
	return nil
}

func (cpc *certPrepareCtl) doPrepareCerts(ch chan<- error) {
	if ch == nil {
		hwlog.RunLog.Errorf("ch is nil")
		return
	}
	var (
		err         error
		rootCertMgr *certutils.RootCertMgr
	)

	if rootCertMgr, err = cpc.prepareCA(); err != nil {
		ch <- err
		return
	}

	for _, component := range cpc.components {
		componentMgr := util.GetComponentMgr(component)
		if err = componentMgr.PrepareComponentCert(rootCertMgr, cpc.certPathMgr); err != nil {
			hwlog.RunLog.Errorf("prepare %s component cert failed: %s", component, err.Error())
			ch <- errors.New("prepare single component cert failed")
			return
		}
	}

	if err := util.PrepareKubeConfigCert(cpc.certPathMgr); err != nil {
		hwlog.RunLog.Errorf("prepare kube config cert failed: %s", err.Error())
		ch <- errors.New("prepare kube config cert failed")
	}

	ch <- nil
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

	rootKmcCfg := kmc.GetKmcCfg(kmcKeyPath, kmcBackKeyPath)
	initCertMgr := certutils.InitRootCertMgr(rootCaFilePath, rootPrivFilePath,
		common.MefCertCommonNamePrefix, rootKmcCfg)
	if _, err := initCertMgr.NewRootCaWithBackup(); err != nil {
		hwlog.RunLog.Errorf("init root ca info failed: %v", err)
		return nil, errors.New("init root ca info failed")
	}
	return initCertMgr, nil
}
