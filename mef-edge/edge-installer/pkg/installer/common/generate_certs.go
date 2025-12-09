// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common this file for generating edge certs task
package common

import (
	"context"
	"errors"
	"fmt"
	"net"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

type componentInfo struct {
	name string
	uid  uint32
	gid  uint32
}

// GenerateCertsTask the task for generate certs
type GenerateCertsTask struct {
	configPathMgr *pathmgr.ConfigPathMgr
	components    []componentInfo
}

// NewGenerateCertsTask create generate certs task
func NewGenerateCertsTask(installRootDir string) (*GenerateCertsTask, error) {
	edgeUserId, edgeGroupId, err := util.GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get edge user and group id failed, error: %v", err)
		return nil, fmt.Errorf("get edge user and group id failed")
	}

	components := []componentInfo{
		{name: constants.EdgeOm, uid: constants.RootUserUid, gid: constants.RootUserGid},
		{name: constants.EdgeMain, uid: edgeUserId, gid: edgeGroupId},
		{name: constants.EdgeCore, uid: constants.RootUserUid, gid: constants.RootUserGid},
	}
	return &GenerateCertsTask{
		configPathMgr: pathmgr.NewConfigPathMgr(installRootDir),
		components:    components,
	}, nil
}

// Run generate certs task
func (gc *GenerateCertsTask) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), constants.GenerateCertWaitTime)
	defer cancel()
	ch := make(chan error)
	go gc.prepareEdgeCerts(ch)
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		hwlog.RunLog.Warn("generate inner certs timeout!")
		return nil
	}
}

// MakeSureEdgeCerts make sure edge certs are generated
func (gc *GenerateCertsTask) MakeSureEdgeCerts() error {
	var checkCerts []string
	for _, component := range gc.components {
		checkCerts = append(checkCerts, gc.configPathMgr.GetCompInnerRootCertPath(component.name))
		checkCerts = append(checkCerts, gc.configPathMgr.GetCompInnerSvrCertPath(component.name))
	}

	allCertsValid := true
	for _, cert := range checkCerts {
		if !fileutils.IsExist(cert) || !gc.checkCertDate(cert) {
			allCertsValid = false
			break
		}
	}
	if allCertsValid {
		return nil
	}

	fmt.Println("generating certificates...")
	hwlog.RunLog.Warn("edge inner certs not exist or with wrong status, generate them now")
	ch := make(chan error, 1)
	gc.prepareEdgeCerts(ch)
	return <-ch
}

func (gc *GenerateCertsTask) checkCertDate(certPath string) bool {
	certBytes, err := fileutils.LoadFile(certPath)
	if err != nil {
		hwlog.RunLog.Warnf("load content of cert %s failed: %v", certPath, err)
		return false
	}

	if certBytes == nil {
		hwlog.RunLog.Warnf("content of cert %s is nil", certPath)
		return false
	}

	caCert, err := x509.LoadCertsFromPEM(certBytes)
	if err != nil {
		hwlog.RunLog.Warnf("load cert %s failed: %v", certPath, err)
		return false
	}

	const overdueDays = 90
	overdueData, err := x509.GetValidityPeriod(caCert, false)
	if err != nil {
		hwlog.RunLog.Warnf("get cert %s overdue days failed: %v", certPath, err)
		return false
	}
	if overdueData <= overdueDays {
		hwlog.RunLog.Warn("the certificate is overdue or close to overdue")
		return false
	}
	return true
}

func (gc *GenerateCertsTask) prepareEdgeCerts(ch chan<- error) {
	if ch == nil {
		hwlog.RunLog.Error("prepare edge certs input param ch is nil")
		return
	}

	rootCertMgr, err := gc.prepareRootCaCert()
	if err != nil {
		ch <- err
		return
	}

	defer gc.clearRootCaDir()
	for _, component := range gc.components {
		if err = gc.prepareComponentCert(rootCertMgr, component.name); err != nil {
			hwlog.RunLog.Errorf("prepare %s component cert failed, error: %v", component.name, err)
			ch <- fmt.Errorf("prepare %s component cert failed", component.name)
			return
		}

		if err = fileutils.CopyFile(gc.configPathMgr.GetTempRootCertPath(),
			gc.configPathMgr.GetCompInnerRootCertPath(component.name)); err != nil {
			hwlog.RunLog.Errorf("copy root ca to %s failed, error: %v", component.name, err)
			ch <- fmt.Errorf("copy root ca to %s failed", component.name)
			return
		}

		if err = gc.setCertPerm(component.name); err != nil {
			hwlog.RunLog.Errorf("set %s permission failed, error: %v", component.name, err)
			ch <- fmt.Errorf("set %s permission failed", component.name)
			return
		}

		if err = gc.setCertOwner(component); err != nil {
			hwlog.RunLog.Errorf("set %s owner failed, error: %v", component.name, err)
			ch <- fmt.Errorf("set %s owner failed", component.name)
			return
		}
		hwlog.RunLog.Infof("prepare component [%s] cert success", component.name)
	}
	hwlog.RunLog.Info("prepare edge inner certs success")
	ch <- nil
}

func (gc *GenerateCertsTask) clearRootCaDir() {
	if err := fileutils.DeleteAllFileWithConfusion(gc.configPathMgr.GetTempRootCertDir()); err != nil {
		hwlog.RunLog.Warnf("remove root cert and key files failed, error: %v", err)
	} else {
		hwlog.RunLog.Infof("Destroyed %s key file", gc.configPathMgr.GetTempRootCerKeyPath())
	}

	if err := fileutils.DeleteAllFileWithConfusion(gc.configPathMgr.GetCompKmcDir(constants.EdgeInstaller)); err != nil {
		hwlog.RunLog.Warnf("remove kmc dir for root cert failed, error: %v", err)
	} else {
		hwlog.RunLog.Infof("Destroyed kmc key files in %s", gc.configPathMgr.GetCompKmcDir(constants.EdgeInstaller))
	}
}

func (gc *GenerateCertsTask) prepareRootCaCert() (*certutils.RootCertMgr, error) {
	kmcCfgPath := gc.configPathMgr.GetCompKmcConfigPath(constants.EdgeOm)
	if err := kmc.InitKmcCfg(kmcCfgPath); err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, will use default kmc config", err)
	}

	kmcRootCfg, err := util.GetKmcConfig(gc.configPathMgr.GetCompKmcDir(constants.EdgeInstaller))
	if err != nil {
		hwlog.RunLog.Errorf("get kmc config for root cert failed, error: %v", err)
		return nil, errors.New("get kmc config for root cert failed")
	}

	rootCertMgr := certutils.InitRootCertMgr(gc.configPathMgr.GetTempRootCertPath(),
		gc.configPathMgr.GetTempRootCerKeyPath(), constants.MefCertCommonNamePrefix, kmcRootCfg)
	if _, err = rootCertMgr.NewRootCa(); err != nil {
		hwlog.RunLog.Errorf("new root ca failed, error: %v", err)
		return nil, errors.New("new root ca failed")
	}
	return rootCertMgr, nil
}

func (gc *GenerateCertsTask) prepareComponentCert(rootCertMgr *certutils.RootCertMgr, component string) error {
	if rootCertMgr == nil {
		hwlog.RunLog.Error("pointer rootCertMgr is nil")
		return errors.New("pointer rootCertMgr is nil")
	}

	kmcCfgPath := gc.configPathMgr.GetCompKmcConfigPath(component)
	if err := kmc.InitKmcCfg(kmcCfgPath); err != nil {
		hwlog.RunLog.Warnf("init kmc config from json failed: %v, will use default kmc config", err)
	}

	kmcCfg, err := util.GetKmcConfig(gc.configPathMgr.GetCompKmcDir(component))
	if err != nil {
		hwlog.RunLog.Errorf("get kmc config for %s cert failed, error: %v", component, err)
		return errors.New("get kmc config for cert failed")
	}

	selfSignCert := certutils.SelfSignCert{
		RootCertMgr:      rootCertMgr,
		SvcCertPath:      gc.configPathMgr.GetCompInnerSvrCertPath(component),
		SvcKeyPath:       gc.configPathMgr.GetCompInnerSvrKeyPath(component),
		CommonNamePrefix: constants.MefCertCommonNamePrefix,
		KmcCfg:           kmcCfg,
		San:              certutils.CertSan{IpAddr: []net.IP{net.ParseIP(constants.LocalIp)}},
	}
	if err = selfSignCert.CreateSignCert(); err != nil {
		hwlog.RunLog.Errorf("create component [%s] cert failed, error: %v", component, err)
		return errors.New("create component cert failed")
	}
	return nil
}

func (gc *GenerateCertsTask) setCertPerm(component string) error {
	certDir := gc.configPathMgr.GetCompInnerCertsDir(component)
	kmcDir := gc.configPathMgr.GetCompKmcDir(component)
	if err := fileutils.SetPathPermission(certDir, constants.Mode400, true, false); err != nil {
		hwlog.RunLog.Errorf("set cert files mode in dir [%s] failed, error: %v", certDir, err)
		return fmt.Errorf("set cert files mode in dir [%s] failed", certDir)
	}

	if err := fileutils.SetPathPermission(certDir, constants.Mode700, false, true); err != nil {
		hwlog.RunLog.Errorf("set cert dirs in dir [%s] mode failed, error: %v", certDir, err)
		return fmt.Errorf("set cert dirs in dir [%s] mode failed", certDir)
	}

	if err := fileutils.SetPathPermission(kmcDir, constants.Mode600, true, false); err != nil {
		hwlog.RunLog.Errorf("set kmc files mode in dir [%s] failed, error: %v", kmcDir, err)
		return fmt.Errorf("set kmc files mode in dir [%s] failed", kmcDir)
	}

	if err := fileutils.SetPathPermission(kmcDir, constants.Mode700, false, true); err != nil {
		hwlog.RunLog.Errorf("set kmc dir [%s] mode failed, error: %v", kmcDir, err)
		return fmt.Errorf("set kmc dir [%s] mode failed", kmcDir)
	}
	return nil
}

func (gc *GenerateCertsTask) setCertOwner(component componentInfo) error {
	compConfigDir := gc.configPathMgr.GetCompConfigDir(component.name)
	param := fileutils.SetOwnerParam{
		Path:       compConfigDir,
		Uid:        component.uid,
		Gid:        component.gid,
		Recursive:  true,
		IgnoreFile: false,
	}
	if err := fileutils.SetPathOwnerGroup(param); err != nil {
		hwlog.RunLog.Errorf("set dir [%s] owner and group failed, error: %v", compConfigDir, err)
		return errors.New("set dir owner and group failed")
	}
	return nil
}
