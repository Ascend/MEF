// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager module edgemsgmanager init
package edgemsgmanager

import (
	"context"
	"net/http"
	"path/filepath"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/types"
)

var nodesProgress map[string]types.ProgressInfo

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
	nodesProgress = make(map[string]types.ProgressInfo, 0)
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
	msg := methodSelect(req)
	if msg == nil {
		hwlog.RunLog.Errorf("%s get method by option and resource failed", nm.Name())
		return
	}

	if !req.GetIsSync() {
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
	common.Combine(http.MethodPost, filepath.Join(edgeSoftwareRootPath, "/download")):         downloadSoftware,
	common.Combine(http.MethodPost, filepath.Join(edgeSoftwareRootPath, "/upgrade")):          upgradeEdgeSoftware,
	common.Combine(http.MethodGet, filepath.Join(edgeSoftwareRootPath, "/version-info")):      queryEdgeSoftwareVersion,
	common.Combine(http.MethodGet, filepath.Join(edgeSoftwareRootPath, "/download-progress")): queryEdgeDownloadProgress,

	common.Combine(common.OptGet, common.ResConfig):              GetConfigInfo,
	common.Combine(common.OptGet, common.ResDownLoadCert):        GetCertInfo,
	common.Combine(common.OptReport, common.ResDownloadProgress): UpdateEdgeDownloadProgress,
}
