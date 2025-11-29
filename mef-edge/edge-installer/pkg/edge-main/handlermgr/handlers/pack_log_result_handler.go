// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
