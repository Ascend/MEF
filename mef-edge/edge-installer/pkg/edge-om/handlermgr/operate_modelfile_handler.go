// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlermgr for deal every handler
package handlermgr

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/checker"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/types"
)

type operateModelFileHandler struct{}

// Handle entry
func (o *operateModelFileHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start operate model file")

	var operateContent types.OperateModelFileContent
	if err := msg.ParseContent(&operateContent); err != nil {
		o.sendResponse(msg, constants.Failed)
		hwlog.RunLog.Errorf("parse operate content failed, error: %v", err)
		return errors.New("parse operate content failed")
	}

	operateContentChecker := checker.GetAndChecker(checker.GetStringChoiceChecker("Operate",
		[]string{constants.OptCheck, constants.OptUpdate, constants.OptSync, constants.OptDelete}, true))
	if checkResult := operateContentChecker.Check(operateContent); !checkResult.Result {
		o.sendResponse(msg, constants.Failed)
		hwlog.RunLog.Errorf("check operate content failed, error: %s", checkResult.Reason)
		return errors.New("check operate content failed")
	}

	operateModelFile := NewOperateModelFile(operateContent)
	if operateModelFile.operateContent.Operate == "sync" {
		toDelList := operateModelFile.syncFiles()
		toDelListStr, err := json.Marshal(toDelList)
		if err != nil {
			hwlog.RunLog.Errorf("cannot marshal sync del list: %v", err)
			return errors.New("marshal sync list fail")
		}
		o.sendResponse(msg, string(toDelListStr))
		return nil
	}
	if err := operateModelFile.OperateModelFile(); err != nil {
		o.sendResponse(msg, constants.Failed)
		hwlog.RunLog.Errorf("operate model file failed, error: %v", err)
		return errors.New("operate model file failed")
	}

	o.sendResponse(msg, constants.Success)
	hwlog.RunLog.Info("operate model file success")
	return nil
}

func (o *operateModelFileHandler) sendResponse(msg *model.Message, respMsg string) {
	newResp, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("get new response message failed, error: %v", err)
		return
	}
	if err = newResp.FillContent(respMsg); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return
	}
	if err = sendHandlerReplyMsg(newResp); err != nil {
		hwlog.RunLog.Errorf("send operate model file handler response failed, error: %v", err)
	}
}
