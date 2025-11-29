// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgeproxy forward msg to device-om, edgecore or edge-om
package edgeproxy

import (
	"context"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/constants"
)

// edgeOmProxy edge-om client proxy
type edgeOmProxy struct {
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

// NewEdgeOmProxy edge-om client proxy
func NewEdgeOmProxy(enable bool) model.Module {
	module := &edgeOmProxy{
		enable: enable,
	}
	module.ctx, module.cancel = context.WithCancel(context.Background())
	return module
}

// Name module name
func (m *edgeOmProxy) Name() string {
	return constants.ModEdgeOm
}

// Enable module enable
func (m *edgeOmProxy) Enable() bool {
	return m.enable
}

// Stop module stop
func (m *edgeOmProxy) Stop() bool {
	m.cancel()
	return true
}

// Start edge-om client proxy
func (m *edgeOmProxy) Start() {
	hwlog.RunLog.Info("edge-om client proxy start success")
	for {
		select {
		case <-m.ctx.Done():
			hwlog.RunLog.Info("----------------edge-om client proxy exit-------------------")
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(m.Name())
		if err != nil {
			hwlog.RunLog.Errorf("edge-om client proxy receive module message failed, error: %v", err)
			continue
		}
		hwlog.RunLog.Infof("%s receive msg router: %+v, route: %+v", "edge-om proxy", msg.Router, msg.KubeEdgeRouter)
		m.sendMsgToClient(msg)
	}
}

func (m *edgeOmProxy) sendMsgToClient(msg *model.Message) {
	if err := SendMsgToWs(msg, constants.ModEdgeOm); err != nil {
		hwlog.RunLog.Errorf("send %v message to edge-om error: %v", msg.Router.Resource, err)
	}
}
