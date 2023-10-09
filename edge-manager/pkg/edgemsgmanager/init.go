// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager module edgemsgmanager init
package edgemsgmanager

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/cache"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"huawei.com/mindxedge/base/common"
)

const (
	maxEntries   = 2048
	neverOverdue = -1
)

var nodesProgress = cache.New(maxEntries)

// NodeMsgDealer [struct] to deal node msg
type NodeMsgDealer struct {
	ctx    context.Context
	enable bool
}

// Name returns the name of NodeMsgDealer module
func (nm *NodeMsgDealer) Name() string {
	return common.NodeMsgManagerName
}

// Enable indicates whether this module is enabled
func (nm *NodeMsgDealer) Enable() bool {
	return nm.enable
}

// NewNodeMsgManager new NodeMsgDealer
func NewNodeMsgManager(enable bool) *NodeMsgDealer {
	nodeMsgDealer := &NodeMsgDealer{
		enable: enable,
		ctx:    context.Background(),
	}
	return nodeMsgDealer
}

// Start receives and sends message
func (nm *NodeMsgDealer) Start() {
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

		msg, err := modulemgr.ReceiveMessage(nm.Name())
		if err != nil {
			hwlog.RunLog.Errorf("module [%s] receive message from channel failed, error: %v", nm.Name(), err)
			continue
		}
		hwlog.RunLog.Infof("receive msg header: %+v, router: %+v", msg.Header, msg.Router)

		go nm.dispatch(msg)
	}
}

func (nm *NodeMsgDealer) dispatch(req *model.Message) {
	var err error
	msg := methodSelect(req)
	if msg == nil {
		msg, err = methodWithOpLogSelect(handlerWithOpLogFuncMap, req)
		if err != nil {
			hwlog.RunLog.Error(err)
			return
		}
	}

	if !req.GetIsSync() || req.GetSource() == common.WebsocketName {
		hwlog.RunLog.Infof(
			"ignore response for async/websocket message, header: %+v, router: %+v", req.Header, req.Router)
		hwlog.RunLog.Infof(
			"handle async/websocket message successfully, status: %+v, message: %+v", msg.Status, msg.Msg)
		return
	}

	resp, err := req.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("%s new response failed", nm.Name())
		return
	}

	resp.FillContent(msg)
	if err = modulemgr.SendMessage(resp); err != nil {
		hwlog.RunLog.Errorf("%s send response failed: %v", nm.Name(), err)
	}
}

type handlerFunc func(req interface{}) common.RespMsg

func methodSelect(req *model.Message) *common.RespMsg {
	var res common.RespMsg
	method, ok := handlerFuncMap[common.Combine(req.GetOption(), req.GetResource())]
	if !ok {
		return nil
	}
	res = method(req)
	return &res
}

func methodWithOpLogSelect(funcMap map[string]handlerFunc, req *model.Message) (*common.RespMsg, error) {
	sn := req.GetNodeId()
	ip := req.GetIp()
	var res common.RespMsg
	method, ok := funcMap[common.Combine(req.GetOption(), req.GetResource())]
	if !ok {
		return nil, fmt.Errorf("handler func is not exist, option: %s, resource: %s", req.GetOption(),
			req.GetResource())
	}
	hwlog.RunLog.Infof("%v [%v:%v] %v %v start", time.Now().Format(time.RFC3339), ip, sn, req.GetOption(),
		req.GetResource())
	res = method(req)
	if res.Status == common.Success {
		hwlog.RunLog.Infof("%v [%v:%v] %v %v success", time.Now().Format(time.RFC3339),
			ip, sn, req.GetOption(), req.GetResource())
	} else {
		hwlog.RunLog.Errorf("%v [%v:%v] %v %v failed", time.Now().Format(time.RFC3339),
			ip, sn, req.GetOption(), req.GetResource())
	}
	return &res, nil
}

var (
	edgeSoftwareRootPath = "/edgemanager/v1/software/edge"
)

var handlerFuncMap = map[string]handlerFunc{
	common.Combine(http.MethodPost, filepath.Join(edgeSoftwareRootPath, "/download")):         downloadSoftware,
	common.Combine(http.MethodPost, filepath.Join(edgeSoftwareRootPath, "/upgrade")):          upgradeEdgeSoftware,
	common.Combine(http.MethodGet, filepath.Join(edgeSoftwareRootPath, "/version-info")):      queryEdgeSoftwareVersion,
	common.Combine(http.MethodGet, filepath.Join(edgeSoftwareRootPath, "/download-progress")): queryEdgeDownloadProgress,

	common.Combine(common.OptGet, common.ResConfig):       GetConfigInfo,
	common.Combine(common.OptGet, common.ResDownLoadCert): GetCertInfo,
}

var handlerWithOpLogFuncMap = map[string]handlerFunc{
	common.Combine(common.OptReport, common.ResDownloadProgress): UpdateEdgeDownloadProgress,
}
