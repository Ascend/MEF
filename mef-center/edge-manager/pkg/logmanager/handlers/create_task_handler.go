// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/handler"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/tasks"
	"edge-manager/pkg/types"
)

// NewCreateTaskHandler get createTaskHandler
func NewCreateTaskHandler() handler.HandleBase {
	return &createTaskHandler{}
}

type createTaskHandler struct {
}

func (h *createTaskHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle task creating")
	req, err := h.parseAndCheckArgs(msg)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse or check arguments, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}, msg)
	}

	edgeNodeSNs, edgeNodeIps, err := getNodeSnAndIpByID(req.EdgeNodes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get serial number of edge node, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorLogDumpNodeInfoError, Msg: err.Error()}, msg)
	}

	taskId, err := tasks.SubmitLogDumpTask(edgeNodeSNs, edgeNodeIps, req.EdgeNodes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to create master task, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorLogDumpBusiness, Msg: err.Error()}, msg)
	}
	return sendRestfulResponse(common.RespMsg{Data: CreateTaskResp{TaskId: taskId}, Status: common.Success}, msg)
}

func (h *createTaskHandler) parseAndCheckArgs(msg *model.Message) (CreateTaskReq, error) {
	var req CreateTaskReq
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return CreateTaskReq{}, errors.New("parse content failed")
	}
	if checkResult := newCreateTaskReqChecker().Check(req); !checkResult.Result {
		return CreateTaskReq{}, fmt.Errorf("check request failed, %s", checkResult.Reason)
	}
	return req, nil
}

func getNodeSnAndIpByID(nodeIds []uint64) ([]string, []string, error) {
	router := common.Router{
		Source:      constants.LogManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    constants.NodeSnAndIp,
	}
	resp := common.SendSyncMessageByRestful(
		types.InnerGetNodeInfosReq{NodeIds: nodeIds}, &router, common.ResponseTimeout)
	if resp.Status != common.Success {
		return nil, nil, errors.New(resp.Msg)
	}
	var nodeInfos types.InnerGetNodeInfosResp
	if err := utils.ObjectConvert(resp.Data, &nodeInfos); err != nil {
		return nil, nil, errors.New("convert internal response error")
	}
	notFoundIdSet := utils.NewSet()
	for _, nodeId := range nodeIds {
		notFoundIdSet.Add(strconv.FormatUint(nodeId, common.BaseHex))
	}
	var (
		serialNumbers []string
		ips           []string
	)
	for _, info := range nodeInfos.NodeInfos {
		serialNumbers = append(serialNumbers, info.SerialNumber)
		ips = append(ips, info.Ip)
		notFoundIdSet.Delete(strconv.FormatUint(info.NodeID, common.BaseHex))
	}
	notFoundIdList := notFoundIdSet.List()
	if len(notFoundIdList) > 0 {
		return nil, nil, fmt.Errorf("node (%s) not found", strings.Join(notFoundIdList, ","))
	}
	return serialNumbers, ips, nil
}
