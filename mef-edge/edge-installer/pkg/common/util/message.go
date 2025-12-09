// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package util for
package util

import (
	"fmt"

	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

// InnerMsgParams is the struct contains all necessary params for inner msg
type InnerMsgParams struct {
	Source                string
	Destination           string
	Operation             string
	Resource              string
	Content               interface{}
	TransferStructIntoStr bool
}

// NewInnerMsgWithFullParas util method for create message
func NewInnerMsgWithFullParas(params InnerMsgParams) (*model.Message, error) {
	msg, err := model.NewMessage()
	if err != nil {
		return nil, fmt.Errorf("new message error, %v", err)
	}
	msg.SetRouter(params.Source, params.Destination, params.Operation, params.Resource)
	if params.Content == nil {
		return msg, nil
	}

	if err = msg.FillContent(params.Content, params.TransferStructIntoStr); err != nil {
		return nil, err
	}
	return msg, nil
}

// SendSyncMsg send sync msg
func SendSyncMsg(params InnerMsgParams) (string, error) {
	msg, err := NewInnerMsgWithFullParas(params)
	if err != nil {
		return "", fmt.Errorf("new model message failed, error: %v", err)
	}

	resp, err := modulemgr.SendSyncMessage(msg, constants.WsSycMsgWaitTime)
	if err != nil {
		return "", fmt.Errorf("send sync message failed, error: %v", err)
	}

	var data string
	if err = resp.ParseContent(&data); err != nil {
		return "", fmt.Errorf("fill data into content failed: %v", err)
	}

	return data, nil
}

// SendInnerMsgResponse send response for inner message
func SendInnerMsgResponse(message *model.Message, content interface{}, transferStructIntoStr ...bool) error {
	resp, err := message.NewResponse()
	if err != nil {
		return fmt.Errorf("create response message failed, %v", err)
	}
	if err := resp.FillContent(content, transferStructIntoStr...); err != nil {
		return fmt.Errorf("fill response message failed, %v", err)
	}
	if err := modulemgr.SendMessage(resp); err != nil {
		return fmt.Errorf("send response message failed, %v", err)
	}
	return nil
}
