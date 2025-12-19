// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package websocket init websocket service for communicating with edge-manager
package websocket

import (
	"context"
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

const (
	initTryTime         = 6
	initStartWsWaitTime = 5 * time.Second
	maxStartWsWaitTime  = 60 * time.Second
)

// alarmWsClient edge hub module
type alarmWsClient struct {
	ctx    context.Context
	enable bool
}

// NewAlarmWsClient new edge hub
func NewAlarmWsClient(enable bool, ctx context.Context) model.Module {
	return &alarmWsClient{
		enable: enable,
		ctx:    ctx,
	}
}

// Name module name
func (m *alarmWsClient) Name() string {
	return common.AlarmManagerWsMoudle
}

// Enable module enable
func (m *alarmWsClient) Enable() bool {
	return m.enable
}

// Start module start running
func (m *alarmWsClient) Start() {
	hwlog.RunLog.Info("-------------------alarm websocket client start--------------------------")

	failedCount := 0
	waitTime := initStartWsWaitTime
	const addTime = 5 * time.Second
	for {
		if waitTime < maxStartWsWaitTime {
			waitTime = waitTime + addTime
		}
		err := initClient()
		if err == nil {
			hwlog.RunLog.Info("init alarm-manager ws client success")
			break
		}
		failedCount++
		hwlog.RunLog.Errorf("init alarm-manager ws client failed: %v", err)
		if failedCount == initTryTime {
			return
		}
		time.Sleep(waitTime)
	}

	for {
		select {
		case <-m.ctx.Done():
			hwlog.RunLog.Error("-------------------alarm websocket client exit--------------------------")
			return
		default:
		}
		receivedMsg, err := modulemgr.ReceiveMessage(m.Name())
		if err != nil {
			hwlog.RunLog.Errorf("alarm-manager receive module message failed, error: %v", err)
			continue
		}
		hwlog.RunLog.Infof("[routeToCenter], route: %+v, {ID: %s, parentID: %s}", receivedMsg.Router,
			receivedMsg.Header.Id, receivedMsg.Header.ParentId)

		if err = m.sendMsgToSever(receivedMsg); err != nil {
			hwlog.RunLog.Errorf("alarm-manager send message [header: %+v, router: %+v] to mef-center failed, error: %v",
				receivedMsg.Header, receivedMsg.Router, err)
		}
	}
}

func (m *alarmWsClient) sendMsgToSever(msg *model.Message) error {
	msg.SetNodeId(common.AlarmManagerWsMoudle)
	if err := proxy.Send(msg); err != nil {
		hwlog.RunLog.Errorf("alarm-manager sender send message to edge-manager failed, error: %v", err)
		return errors.New("alarm-manager sender send message to edge-manager failed")
	}
	return nil
}
