// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
