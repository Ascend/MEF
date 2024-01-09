// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package alarmmanager to init config manager
package alarmmanager

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

type alarmManager struct {
	enable bool
	ctx    context.Context
}

// NewAlarmManager create config manager
func NewAlarmManager(enable bool) model.Module {
	return &alarmManager{
		enable: enable,
		ctx:    context.Background(),
	}
}

func (am *alarmManager) Name() string {
	return common.AlarmManagerName
}

func (am *alarmManager) Enable() bool {
	return am.enable
}

func (am *alarmManager) Start() {
	for {
		select {
		case _, ok := <-am.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		req, err := modulemgr.ReceiveMessage(am.Name())
		if err != nil {
			hwlog.RunLog.Errorf("%s receive request from restful service failed", am.Name())
			continue
		}

		go am.forwardMsgToAlarmManager(req)
	}
}

func (am *alarmManager) forwardMsgToAlarmManager(msg *model.Message) {
	sn := msg.GetPeerInfo().Sn

	var content string
	if err := msg.ParseContent(&content); err != nil {
		hwlog.RunLog.Errorf("add alarm parse content failed: %v", err)
		return
	}

	var alarmReq requests.AddAlarmReq
	err := json.Unmarshal([]byte(content), &alarmReq)
	if err != nil {
		hwlog.RunLog.Errorf("unmarshal alarm req failed: %s", err.Error())
		return
	}

	ip, err := am.getIpBySn(sn)
	if err != nil {
		return
	}

	alarmReq.Sn = sn
	alarmReq.Ip = ip

	if err = msg.FillContent(alarmReq, true); err != nil {
		hwlog.RunLog.Errorf("fill alarm req into content failed: %v", err)
		return
	}
	msg.SetNodeId(common.AlarmManagerClientName)
	msg.Router.Destination = common.InnerServerName
	if err = modulemgr.SendAsyncMessage(msg); err != nil {
		hwlog.RunLog.Errorf("send msg to inner server failed: %s", err.Error())
	}
}

func (am *alarmManager) getIpBySn(sn string) (string, error) {
	const waitTime = 10 * time.Second

	router := common.Router{
		Source:      common.AlarmManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Get,
		Resource:    common.GetIpBySn,
	}
	resp := common.SendSyncMessageByRestful(sn, &router, waitTime)
	if resp.Status != common.Success {
		hwlog.RunLog.Errorf("send msg to node manager failed: %s", resp.Msg)
		return "", errors.New("send msg to node manager failed")
	}

	ip, ok := resp.Data.(string)
	if !ok {
		hwlog.RunLog.Errorf("unsupported type of ip received")
		return "", errors.New("unsupported type of ip received")
	}

	return ip, nil
}
