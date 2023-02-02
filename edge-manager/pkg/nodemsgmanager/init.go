// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package nodemsgmanager module nodemsgmanager init
package nodemsgmanager

import (
	"context"
	"net/http"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

// NodeMsgManager edge-installer struct
type NodeMsgManager struct {
	ctx    context.Context
	enable bool
}

// Name returns the name of edge installer module
func (nm *NodeMsgManager) Name() string {
	return common.NodeMsgManagerName
}

// Enable indicates whether this module is enabled
func (nm *NodeMsgManager) Enable() bool {
	if !nm.enable {
		return !nm.enable
	}
	return nm.enable
}

// NewNodeMsgManager new NodeMsgManager
func NewNodeMsgManager(enable bool) *NodeMsgManager {
	i := &NodeMsgManager{
		enable: enable,
		ctx:    context.Background(),
	}
	return i
}

// Start receives and sends message
func (nm *NodeMsgManager) Start() {
	for {
		select {
		case _, ok := <-nm.ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("has listened stop signal")
			return
		default:
		}

		req, err := modulemanager.ReceiveMessage(nm.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message from channel failed, error: %v", nm.Name(), err)
			continue
		}

		dispatch(req)
	}
}

func dispatch(req *model.Message) {
	msg := methodSelect(req)
	if msg == nil {
		hwlog.RunLog.Errorf("%s get method by option and resource failed", common.NodeMsgManagerName)
		return
	}

	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", common.NodeMsgManagerName)
		return
	}

	resp.FillContent(msg)
	if err = modulemanager.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed", common.NodeMsgManagerName)
	}
}

type handlerFunc func(req *model.Message) common.RespMsg

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, ok := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !ok {
		hwlog.RunLog.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
		return nil
	}
	res = method(req)
	return &res
}

var (
	edgeSoftwareRootPath = "/edgemanager/v1/software/edge"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, filepath.Join(edgeSoftwareRootPath, "/upgrade")): UpgradeEdgeSoftware,
	common.Combine(http.MethodPost, filepath.Join(edgeSoftwareRootPath, "/effect")):  EffectEdgeSoftware,
	common.Combine(common.OptPost, "/edgecore/config"):                               GetConfigInfo,
}
