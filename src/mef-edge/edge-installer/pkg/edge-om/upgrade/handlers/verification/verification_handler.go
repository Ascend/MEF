// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package verification unpack and verify downloaded software
package verification

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/edge-om/upgrade/reporter"
	"edge-installer/pkg/installer/preupgrade/flows"
)

// Handler [struct] verify and unpack downloaded software
type Handler struct {
}

// Handle [method] handle msg
func (vh *Handler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle verify and unpack message form edge-main")
	resp, err := msg.NewResponse()
	if err != nil {
		return err
	}
	if err = resp.FillContent("OK"); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return errors.New("fill content failed")
	}
	processErr := vh.processSoftware()
	if processErr != nil {
		hwlog.RunLog.Errorf("verify and unpack downloaded software failed, %v", processErr)
		if fillErr := resp.FillContent(processErr.Error()); fillErr != nil {
			hwlog.RunLog.Errorf("fill process err into content failed: %v", fillErr)
		}
	}
	resp.SetRouter(constants.ModEdgeOm, constants.InnerClient,
		constants.OptResponse, constants.InnerSoftwareVerification)
	if err := modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("response verification message failed: %v", err)
		return err
	}
	return processErr
}

func (vh *Handler) processSoftware() error {
	hwlog.RunLog.Info("start to verify and unpack downloaded software")
	installDir, err := path.GetInstallDir()
	if err != nil {
		return fmt.Errorf("get install dir failed, %v", err)
	}
	flow := flows.NewVerificationInstaller(installDir)
	defer func() {
		go reporter.ReportSoftwareVersion(1)
	}()
	if err = flow.RunTasks(); err != nil {
		return err
	}
	hwlog.RunLog.Info("successfully verify and unpack downloaded software")
	return nil
}
