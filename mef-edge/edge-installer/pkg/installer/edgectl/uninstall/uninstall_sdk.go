// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

package uninstall

import (
	"fmt"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/domainconfig"
)

func (pu processUninstallTask) Run() error {
	var setFunc = []func() error{
		pu.clearImageCert,
		pu.clearHostsFiles,
		pu.unsetImmutable,
		pu.removeService,
		pu.removeExternalFiles,
		pu.removeContainer,
		pu.removeInstallDir,
	}
	for _, function := range setFunc {
		if err := function(); err != nil {
			return err
		}
	}
	return nil
}

func (pu processUninstallTask) clearHostsFiles() error {
	if err := common.InitEdgeOmResource(); err != nil {
		fmt.Println("remove domain config in /etc/hosts failed, " +
			"please remove image registry domain mapping manually")
		hwlog.RunLog.Errorf("clear hosts files failed, init edge om resource error: %v", err)
		return nil
	}
	cfg, err := config.GetDomainCfg()
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("domain config does not exist, do not need to clear")
		return nil
	}
	var errFlag bool
	for _, cfg := range cfg.Configs {
		if err = domainconfig.DeleteDomainCfgInFile(cfg.Domain, cfg.IP); err != nil {
			errFlag = true
		}
	}

	if errFlag {
		fmt.Println("remove domain config in /etc/hosts failed, " +
			"please remove image registry domain mapping manually")
		hwlog.RunLog.Warn("remove domain config in /etc/hosts failed, " +
			"please remove image registry domain mapping manually")
		return nil
	}

	hwlog.RunLog.Info("remove domain config in /etc/hosts success")
	return nil
}

func (pu processUninstallTask) clearImageCert() error {
	if err := common.InitEdgeOmResource(); err != nil {
		fmt.Println("remove image cert in /etc/docker/certs.d failed, please remove image cert manually")
		hwlog.RunLog.Errorf("clear image cert failed, init edge om resource error: %v", err)
		return nil
	}
	cfg, err := config.GetImageCfg()
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("image config is not exist, no need to clear")
		return nil
	}
	if err = util.DeleteImageCertFile(cfg.ImageAddress); err != nil {
		fmt.Println("remove image cert in /etc/docker/certs.d failed, please remove image cert manually")
		hwlog.RunLog.Warn("remove image cert in /etc/docker/certs.d failed, please remove image cert manually")
		return nil
	}
	hwlog.RunLog.Info("remove image cert in /etc/docker/certs.d success")
	return nil
}
