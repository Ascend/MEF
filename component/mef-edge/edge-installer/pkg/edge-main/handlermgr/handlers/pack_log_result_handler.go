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

// Package handlers
package handlers

import (
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
)

type packLogResultHandler struct {
}

func (packLogResultHandler) Handle(msg *model.Message) error {
	var result string
	if err := msg.ParseContent(&result); err != nil {
		return fmt.Errorf("convert pack log result failed, %v", err)
	}

	hwlog.RunLog.Infof("get pack log result: %s", result)
	var err error
	func() {
		defer func() {
			if data := recover(); data != nil {
				err = fmt.Errorf("send_result_error(%v)", data)
			}
		}()

		select {
		case getDumpLogHandler().resultCh <- result:
		default:
			err = errors.New("failed to send pack result")
		}
	}()

	return err
}
