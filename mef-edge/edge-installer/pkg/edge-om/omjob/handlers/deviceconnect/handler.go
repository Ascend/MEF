// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package deviceconnect this file for omjob handler
package deviceconnect

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/edge-om/omjob/handlers/jobs"
)

// Handler device connect handler
type Handler struct{}

// Handle configHandler handle entry
func (ch *Handler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle device connect msg")
	var connectResult bool
	if err := msg.ParseContent(&connectResult); err != nil {
		hwlog.RunLog.Errorf("get result failed: %v", err)
		return errors.New("get result failed")
	}

	hwlog.RunLog.Infof("edge connect device state is :%t", connectResult)
	if connectResult {
		jobs.ReportCapability()
	}
	return nil
}
