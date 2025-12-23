// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package handlers

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/handlermgr/modeltask"
)

type deleteModelFileHandler struct {
}

// Handle The model files need to be deleted when the pods_data is deleted.
func (u *deleteModelFileHandler) Handle(msg *model.Message) error {
	if !modeltask.GetModelMgr().LockGlobal() {
		return lockResError
	}
	defer modeltask.GetModelMgr().UnLockGlobal()
	modeltask.GetModelMgr().CancelTasks()
	content := types.OperateModelFileContent{
		Operate:     constants.OptDelete,
		OperateInfo: nil,
	}
	result, err := util.SendSyncMsg(util.InnerMsgParams{
		Source:                constants.ModHandlerMgr,
		Destination:           constants.ModEdgeOm,
		Operation:             constants.OptRaw,
		Resource:              constants.ActionModelFiles,
		Content:               content,
		TransferStructIntoStr: true,
	})
	if err != nil {
		hwlog.RunLog.Errorf("send pods_data message to edge om failed, error: %v", err)
		modeltask.SendFailResponse(constants.ResourceTypePodsData, "send pods_data message to edge om failed")
		return errors.New("send pods_data message to edge om failed")
	}
	if result == constants.Failed {
		hwlog.RunLog.Error("delete all model file failed by edge om")
		modeltask.SendFailResponse(constants.ResourceTypePodsData, "delete all model file failed by edge om")
		return errors.New("delete all model file failed by edge om")
	}
	modeltask.GetModelMgr().Clear()
	hwlog.RunLog.Info("delete all model file success")
	modeltask.SendConfigResult(constants.ResourceTypePodsData)
	return nil
}
