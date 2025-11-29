// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeproxy forward msg from device-om, edgecore or edge-om
package edgeproxy

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

// deviceOmProxy device-om client proxy
type deviceOmProxy struct {
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

// NewDeviceOmProxy device-om client proxy
func NewDeviceOmProxy(enable bool) model.Module {
	module := &deviceOmProxy{
		enable: enable,
	}
	module.ctx, module.cancel = context.WithCancel(context.Background())
	return module
}

// Name module name
func (m *deviceOmProxy) Name() string {
	return constants.ModDeviceOm
}

// Enable module enable
func (m *deviceOmProxy) Enable() bool {
	return m.enable
}

// Stop module stop
func (m *deviceOmProxy) Stop() bool {
	m.cancel()
	return true
}

// Start device-om client proxy
func (m *deviceOmProxy) Start() {
	hwlog.RunLog.Info("device-om client proxy start success")
	for {
		select {
		case <-m.ctx.Done():
			hwlog.RunLog.Info("-------------------edge-proxy exit-------------------")
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(m.Name())
		if err != nil {
			hwlog.RunLog.Errorf("edge-proxy receive module message failed, error: %v", err)
			continue
		}
		hwlog.RunLog.Infof("[routeToFd], route: %+v, {ID: %s, parentID: %s}", msg.KubeEdgeRouter,
			msg.Header.Id, msg.Header.ParentId)
		m.sendMsgToClient(msg)
	}
}

func (m *deviceOmProxy) sendMsgToClient(msg *model.Message) {
	if err := SendMsgToWs(msg, constants.ModDeviceOm); err != nil {
		hwlog.RunLog.Errorf("SendMsgToWs error: %v", err)
		return
	}
}
