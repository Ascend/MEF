// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package innerclient this file for innerclient module register
package innerclient

import (
	"context"
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-om/omjob/handlers/jobs"
)

// innerClient edge client module
type innerClient struct {
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

// NewEdgeClient new edge client
func NewEdgeClient(ctx context.Context, enable bool) model.Module {
	module := &innerClient{
		ctx:    ctx,
		enable: enable,
	}
	return module
}

// Name module name
func (m *innerClient) Name() string {
	return constants.InnerClient
}

// Enable module enable
func (m *innerClient) Enable() bool {
	return m.enable
}

// Stop module stop
func (m *innerClient) Stop() bool {
	m.cancel()
	return true
}

// Start module start running
func (m *innerClient) Start() {
	hwlog.RunLog.Info("-------------------inner-client start--------------------------")

	var i int
	var err error
	for i = 0; i < constants.DefaultTryCount; i++ {
		if err = InitClient(); err == nil {
			hwlog.RunLog.Info("init inner-client success")
			break
		}
		hwlog.RunLog.Errorf("init inner-client failed, error: %v", err)
		time.Sleep(constants.StartWsWaitTime)
	}

	if i == constants.DefaultTryCount {
		hwlog.RunLog.Error("init inner-client failed, has reached the maximum number of the connection attempts")
		return
	}
	jobs.StartReportJob(m.ctx)
	for {
		select {
		case <-m.ctx.Done():
			hwlog.RunLog.Info("-------------------inner-client exit--------------------------")
			return
		default:
		}
		receivedMsg, err := modulemgr.ReceiveMessage(m.Name())
		if err != nil {
			hwlog.RunLog.Errorf("inner-client get receive module message failed, error: %v", err)
			continue
		}
		hwlog.RunLog.Infof("inner client receive msg option:[%s] resource:[%s] send to inner server",
			receivedMsg.GetOption(), receivedMsg.GetResource())

		if err = m.sendMsgToSever(receivedMsg); err != nil {
			hwlog.RunLog.Errorf("inner-client send message [header: %+v, router: %+v] to inner-server failed, error: %v",
				receivedMsg.Header, receivedMsg.Router, err)
		}
	}
}

func (m *innerClient) sendMsgToSever(msg *model.Message) error {
	sender, err := GetCltSender()
	if err != nil {
		hwlog.RunLog.Errorf("inner-client get client sender failed, error: %v", err)
		return errors.New("inner-client get client sender failed")
	}
	if err = sender.Send(msg); err != nil {
		hwlog.RunLog.Errorf("inner-client sender send message to inner-server failed, error: %v", err)
		return errors.New("inner-client sender send message to inner-server failed")
	}

	return nil
}
