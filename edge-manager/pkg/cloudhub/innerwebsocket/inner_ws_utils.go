// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package innerwebsocket to init config manager
package innerwebsocket

import (
	"encoding/json"
	"errors"
	"time"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

// ModifyMsgForAlarmManager modify msg for clear or report alarms to alarm manager by established inner websocket
func modifyMsgForAlarmManager(msg *model.Message) {
	sn := msg.GetPeerInfo().Sn

	var content string
	if err := msg.ParseContent(&content); err != nil {
		hwlog.RunLog.Errorf("add alarm parse content failed: %v", err)
		return
	}

	var alarmsReq requests.AlarmsReq
	if err := json.Unmarshal([]byte(content), &alarmsReq); err != nil {
		hwlog.RunLog.Errorf("unmarshal alarm req failed: %s", err.Error())
		return
	}

	ip, err := getIpBySn(sn)
	if err != nil {
		return
	}

	alarmsReq.Sn = sn
	alarmsReq.Ip = ip

	if err = msg.FillContent(alarmsReq, true); err != nil {
		hwlog.RunLog.Errorf("fill alarm req into content failed: %v", err)
		return
	}
	msg.Router.Destination = common.AlarmManagerName
}

func getIpBySn(sn string) (string, error) {
	const waitTime = 10 * time.Second

	router := common.Router{
		Source:      common.CloudHubName,
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
