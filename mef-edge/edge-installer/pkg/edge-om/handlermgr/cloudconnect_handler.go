// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

package handlermgr

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-om/common/certsrequest"
	"edge-installer/pkg/edge-om/common/cloudconnect"
)

type cloudConnectHandler struct{}

// Handle cloudConnectHandler cache connection status from edge to center
func (ch *cloudConnectHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle cloud connect msg")
	var connectResult bool
	if err := msg.ParseContent(&connectResult); err != nil {
		hwlog.RunLog.Errorf("get connect result failed: %v", err)
		return errors.New("get connect result failed")
	}
	hwlog.RunLog.Infof("edge connect cloud state is: %t", connectResult)
	cloudconnect.SetCloudConnectStatus(connectResult)
	if connectResult {
		go certsrequest.RequestCertsFromCenter()
	}
	resp, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("create response message failed, %v", err)
		return errors.New("create response message failed")
	}

	if err := resp.FillContent(constants.OK); err != nil {
		hwlog.RunLog.Errorf("fill response message content failed, %v", err)
		return errors.New("fill response message content failed")
	}
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("response message failed, %v", err)
		return errors.New("response message failed")
	}
	return nil
}
