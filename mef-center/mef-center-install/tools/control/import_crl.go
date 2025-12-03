// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package main

import (
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/mef-center-install/pkg/control"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

type importCrlController struct {
	operate      string
	installParam *util.InstallParamJsonTemplate
	crlPath      string
	crlName      string
}

const (
	importCrlPathFlag = "crl_path"
	importPeerFlag    = "peer"
)

func (icc *importCrlController) bindFlag() bool {
	flag.StringVar(&(icc.crlPath), importCrlPathFlag, "", "path that saves crl to import")
	flag.StringVar(&(icc.crlName), importPeerFlag, "",
		"name of crl to import, currently only supports north")
	utils.MarkFlagRequired(importCrlPathFlag)
	utils.MarkFlagRequired(importPeerFlag)
	return true
}

func (icc *importCrlController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	icc.installParam = installParam
}

func (icc *importCrlController) doControl() error {
	if icc.crlName != common.NorthernCertName {
		hwlog.RunLog.Errorf("current version only support [%s] crl name ", common.NorthernCertName)
		return fmt.Errorf("crl name is in valid, [%s] is only value supported", common.NorthernCertName)
	}

	pathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init path mgr failed: %v", err)
		return errors.New("init path mgr failed")
	}
	uid, gid, err := util.GetMefId()
	if err != nil {
		hwlog.RunLog.Errorf("get MEF uid/gid failed: %s", err.Error())
		return errors.New("get MEF uid/gid failed")
	}

	savePath := pathMgr.ConfigPathMgr.GetNorthernCrlPath()
	caPath := pathMgr.ConfigPathMgr.GetNorthernCertPath()
	exchangeFlow := control.NewImportCrlFlow(icc.crlPath, savePath, caPath, uid, gid)
	if err = exchangeFlow.DoImportCrl(); err != nil {
		hwlog.RunLog.Errorf("execute import crl flow failed: %s", err.Error())
		return err
	}

	return nil
}

func (icc *importCrlController) printExecutingLog(ip, user string) {
	hwlog.RunLog.Info("-------------------start to import crl-------------------")
	hwlog.OpLog.Infof("[%s@%s] start to import crl", user, ip)
	fmt.Println("start to import crl")
}

func (icc *importCrlController) printSuccessLog(ip, user string) {
	hwlog.RunLog.Info("-------------------import crl successful-------------------")
	hwlog.OpLog.Infof("[%s@%s] import crl successful", user, ip)
	fmt.Println("import crl successful")
}

func (icc *importCrlController) printFailedLog(ip, user string) {
	hwlog.RunLog.Error("-------------------import crl failed-------------------")
	hwlog.OpLog.Errorf("[%s@%s] import crl failed", user, ip)
	fmt.Println("import crl failed")
}

func (icc *importCrlController) getName() string {
	return icc.operate
}
