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

// edgeOmProxy edgecore client proxy
type edgeCoreProxy struct {
	ctx    context.Context
	cancel context.CancelFunc
	enable bool
}

// NewEdgeCoreProxy edgecore client proxy
func NewEdgeCoreProxy(enable bool) model.Module {
	module := &edgeCoreProxy{
		enable: enable,
	}
	module.ctx, module.cancel = context.WithCancel(context.Background())
	return module
}

// Name module name
func (m *edgeCoreProxy) Name() string {
	return constants.ModEdgeCore
}

// Enable module enable
func (m *edgeCoreProxy) Enable() bool {
	return m.enable
}

// Stop module stop
func (m *edgeCoreProxy) Stop() bool {
	m.cancel()
	return true
}

// Start edgecore client proxy
func (m *edgeCoreProxy) Start() {
	hwlog.RunLog.Info("edgecore client proxy start success")
	for {
		select {
		case <-m.ctx.Done():
			hwlog.RunLog.Info("----------------edgecore client proxy exit-------------------")
			return
		default:
		}
		msg, err := modulemgr.ReceiveMessage(m.Name())
		if err != nil {
			hwlog.RunLog.Errorf("edgecore client proxy receive module message failed, error: %v", err)
			continue
		}
		hwlog.RunLog.Infof("%s receive msg router: %+v, route: %+v", "edgecore proxy", msg.Router, msg.KubeEdgeRouter)
		m.sendMsgToClient(msg)
	}
}

func (m *edgeCoreProxy) sendMsgToClient(msg *model.Message) {
	if err := SendMsgToWs(msg, constants.ModEdgeCore); err != nil {
		hwlog.RunLog.Errorf("send message to edgecore error: %v", err)
	}
}
