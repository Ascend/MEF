// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
	req, err := h.parseAndCheckArgs(msg.Content)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse or check arguments, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}, msg)
	}

	edgeNodes, err := getNodeSerialNumbersByID(req.EdgeNodes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get serial number of edge node, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorLogDumpNodeInfoError, Msg: err.Error()}, msg)
	}

	taskId, err := tasks.SubmitLogDumpTask(edgeNodes)
	if err != nil {
		hwlog.RunLog.Errorf("failed to create master task, %v", err)
		return sendRestfulResponse(common.RespMsg{Status: common.ErrorLogDumpBusiness, Msg: err.Error()}, msg)
	}
	return sendRestfulResponse(common.RespMsg{Data: CreateTaskResp{TaskId: taskId}, Status: common.Success}, msg)
}

func (h *createTaskHandler) parseAndCheckArgs(content interface{}) (CreateTaskReq, error) {
	var req CreateTaskReq
	if err := common.ParamConvert(content, &req); err != nil {
		return CreateTaskReq{}, err
	}
	if checkResult := newCreateTaskReqChecker().Check(req); !checkResult.Result {
		return CreateTaskReq{}, fmt.Errorf("check request failed, %s", checkResult.Reason)
	}
	return req, nil
}

func getNodeSerialNumbersByID(nodeIds []uint64) ([]string, error) {
	router := common.Router{
		Source:      constants.LogManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    constants.NodeSerialNumber,
	}
	resp := common.SendSyncMessageByRestful(
		types.InnerGetNodeInfosReq{NodeIds: nodeIds}, &router, common.ResponseTimeout)
	if resp.Status != common.Success {
		return nil, errors.New(resp.Msg)
	}
	var nodeInfos types.InnerGetNodeInfosResp
	if err := utils.ObjectConvert(resp.Data, &nodeInfos); err != nil {
		return nil, errors.New("convert internal response error")
	}
	notFoundIdSet := utils.NewSet()
	for _, nodeId := range nodeIds {
		notFoundIdSet.Add(strconv.FormatUint(nodeId, common.BaseHex))
	}
	var serialNumbers []string
	for _, info := range nodeInfos.NodeInfos {
		serialNumbers = append(serialNumbers, info.SerialNumber)
		notFoundIdSet.Delete(strconv.FormatUint(info.NodeID, common.BaseHex))
	}
	notFoundIdList := notFoundIdSet.List()
	if len(notFoundIdList) > 0 {
		return nil, fmt.Errorf("node (%s) not found", strings.Join(notFoundIdList, ","))
	}
	return serialNumbers, nil
}
