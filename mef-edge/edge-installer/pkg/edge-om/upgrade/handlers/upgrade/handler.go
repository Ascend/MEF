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

// Package upgrade this file for upgrade software handler
package upgrade

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-om/upgrade/reporter"
	"edge-installer/pkg/installer/preupgrade/flows"
)

// Handler download software handler
type Handler struct {
	effectInfo util.SoftwareUpdateInfo
}

// Handle configHandler handle entry
func (ch *Handler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle upgrade software")
	if err := ch.processSoftware(msg); err != nil {
		return err
	}
	hwlog.RunLog.Info("handle upgrade software success")
	return nil
}

func (ch *Handler) processSoftware(req *model.Message) error {
	ch.effectInfo = util.SoftwareUpdateInfo{}
	if err := req.ParseContent(&ch.effectInfo); err != nil {
		hwlog.RunLog.Error("convert request param failed")
		return err
	}

	installDir, err := path.GetInstallDir()
	if err != nil {
		hwlog.RunLog.Errorf("get install dir failed, error: %v", err)
		return err
	}

	if ch.effectInfo.SoftwareName != constants.MEFEdgeName {
		hwlog.RunLog.Errorf("unknown software %s", ch.effectInfo.SoftwareName)
		return fmt.Errorf("unknown software %s", ch.effectInfo.SoftwareName)
	}

	flow := flows.OnlineUpgradeInstaller(installDir)
	defer func() {
		go reporter.ReportSoftwareVersion(1)
	}()
	if err = flow.RunTasks(); err != nil {
		hwlog.RunLog.Errorf("upgrade software[%s] failed,error:%v", ch.effectInfo.SoftwareName, err)
		return err
	}
	return nil
}
