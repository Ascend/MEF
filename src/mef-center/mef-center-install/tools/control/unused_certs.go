// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package main for
package main

import (
	"errors"
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/mef-center-install/pkg/control"
	"huawei.com/mindxedge/base/mef-center-install/pkg/util"
)

const (
	certNameFlag = "name"
)

type unusedCertsController struct {
	operate      string
	caName       string
	installParam *util.InstallParamJsonTemplate
}

func (ucc *unusedCertsController) doControl() error {
	installPathMgr, err := util.InitInstallDirPathMgr()
	if err != nil {
		hwlog.RunLog.Errorf("init install path mgr failed: %v", err)
		return errors.New("init install path mgr failed")
	}
	controlMgr := control.InitUnusedCertsMgr(installPathMgr, ucc.operate, ucc.caName)
	if err := controlMgr.DoOperate(); err != nil {
		hwlog.RunLog.Errorf("%s unused certificates of [%s] ca failed,err:%s", ucc.getVerb(), ucc.caName, err.Error())
		return err
	}
	return nil
}

func (ucc *unusedCertsController) setInstallParam(installParam *util.InstallParamJsonTemplate) {
	ucc.installParam = installParam
}

func (ucc *unusedCertsController) bindFlag() bool {
	flag.StringVar(&ucc.caName, certNameFlag, "", "third party ca name")
	utils.MarkFlagRequired(certNameFlag)
	return true
}

func (ucc *unusedCertsController) printExecutingLog(ip, user string) {
	fmt.Printf("start to %s unused certificates of [%s] ca\n", ucc.getVerb(), ucc.caName)
	hwlog.RunLog.Infof("-------------------start to %s unused certificates of [%s] ca-------------------",
		ucc.getVerb(), ucc.caName)
	hwlog.OpLog.Infof("[%s@%s] start to %s unused certificates of [%s] ca", user, ip, ucc.getVerb(), ucc.caName)
}

func (ucc *unusedCertsController) printFailedLog(ip, user string) {
	fmt.Printf("%s unused certificates of [%s] ca failed\n", ucc.getVerb(), ucc.caName)
	hwlog.RunLog.Errorf("-------------------%s unused certificates of [%s] ca failed-------------------",
		ucc.getVerb(), ucc.caName)
	hwlog.OpLog.Errorf("[%s@%s] %s unused certificates of [%s] ca failed", user, ip, ucc.getVerb(), ucc.caName)
}

func (ucc *unusedCertsController) printSuccessLog(ip, user string) {
	fmt.Printf("%s unused certificates of [%s] ca successfully\n", ucc.getVerb(), ucc.caName)
	hwlog.RunLog.Infof("-------------------%s unused certificates of [%s] ca successfully-------------------",
		ucc.getVerb(), ucc.caName)
	hwlog.OpLog.Infof("[%s@%s] %s unused certificates of [%s] ca successfully", user, ip, ucc.getVerb(), ucc.caName)
}

func (ucc *unusedCertsController) getName() string {
	return ucc.operate
}

func (ucc *unusedCertsController) getVerb() string {
	return map[string]string{
		util.GetUnusedCertOperateFlag: "get",
		util.RestoreCertOperateFlag:   "restore",
		util.DeleteCertOperateFlag:    "delete",
	}[ucc.operate]
}
