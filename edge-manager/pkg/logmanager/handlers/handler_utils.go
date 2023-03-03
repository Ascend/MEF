// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"errors"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func sendResponse(msg common.RespMsg, req *model.Message) error {
	var originateErr error
	if msg.Status != common.Success {
		errMessage := msg.Msg
		if errMessage == "" {
			var ok bool
			errMessage, ok = common.ErrorMap[msg.Status]
			if !ok {
				errMessage = ""
			}
		}
		originateErr = errors.New(errMessage)
	}
	resp, err := req.NewResponse()
	if err != nil {
		if originateErr == nil {
			originateErr = err
		}
		hwlog.RunLog.Error("failed to create message")
		return originateErr
	}
	resp.FillContent(msg)
	if err := modulemanager.SendMessage(resp); err != nil {
		if originateErr == nil {
			originateErr = err
		}
		hwlog.RunLog.Error("failed to send message")
		return originateErr
	}
	return originateErr
}
