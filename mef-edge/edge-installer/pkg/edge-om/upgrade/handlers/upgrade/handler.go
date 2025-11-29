// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
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
