// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeinstaller for set handler
package edgeinstaller

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

type accountHandler struct{}

// Handle accountHandler handle entry
func (ah *accountHandler) Handle(message *model.Message) error {
	hwlog.RunLog.Info("----------set edge account start----------")
	hwlog.RunLog.Info("edge-installer received message success for setting edge account from restful module")

	respToRestful := setEdgeAccount(message.GetContent())
	resp, err := message.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed, error: %v", common.EdgeInstallerName, err)
		return fmt.Errorf("%s new response failed, error: %v", common.EdgeInstallerName, err)
	}
	resp.FillContent(respToRestful)
	if err = modulemanager.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send message failed, error: %v", common.EdgeInstallerName, err)
		return fmt.Errorf("%s send message failed, error: %v", common.EdgeInstallerName, err)
	}

	hwlog.RunLog.Info("edge-installer send message to restful success for setting edge account")
	hwlog.RunLog.Info("----------set edge account end----------")
	return nil
}
