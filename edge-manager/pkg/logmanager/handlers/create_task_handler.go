// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/logmanager/constants"
	"edge-manager/pkg/logmanager/modules"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker/checker"
	"huawei.com/mindxedge/base/common/handlerbase"
	"huawei.com/mindxedge/base/common/logmgmt/logcollect"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"

	"edge-manager/pkg/types"
)

const sendMessageTimeout = 5 * time.Second

// GetCreateTaskHandler get createTaskHandler
func GetCreateTaskHandler(
	taskMgr modules.TaskMgr, ip string, port int) handlerbase.HandleBase {
	return &createTaskHandler{progressMgr: taskMgr, ip: ip, port: port}
}

type createTaskHandler struct {
	progressMgr modules.TaskMgr
	ip          string
	port        int
}

func (h *createTaskHandler) Handle(msg *model.Message) error {
	hwlog.RunLog.Info("start to handle task creation")
	req, err := h.parse(msg.Content)
	if err != nil {
		hwlog.RunLog.Errorf("failed to handle task creation, %v", err)
		return sendResponse(common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error()}, msg)
	}
	if err := h.check(req); err != nil {
		hwlog.RunLog.Errorf("failed to handle task creation, %v", err)
		return sendResponse(common.RespMsg{Status: common.ErrorParamInvalid, Msg: err.Error()}, msg)
	}
	var resp types.BatchResp
	for _, node := range req.EdgeNodes {
		uploadConfig := h.prepareUpload(node)
		if err := h.progressMgr.AddTask(node, filepath.Base(uploadConfig.MethodAndUrl.Url)); err != nil {
			hwlog.RunLog.Warnf("failed to add task: %v", err)
			resp.FailedIDs = append(resp.FailedIDs, node)
			continue
		}
		if err := h.sendReqToEdge(node, uploadConfig); err != nil {
			sendErr := h.progressMgr.NotifyProgress(logcollect.TaskProgress{
				Status:  common.ErrorLogCollectEdgeBusiness,
				Message: "failed to send message to edge",
			}, node)
			hwlog.RunLog.Warnf("failed to send message to edge: %v, %v", err, sendErr)
			resp.FailedIDs = append(resp.FailedIDs, node)
			continue
		}
		resp.SuccessIDs = append(resp.SuccessIDs, node)
	}
	if len(resp.FailedIDs) > 0 {
		hwlog.RunLog.Error("failed to handle task creation")
		return sendResponse(
			common.RespMsg{Status: common.ErrorLogCollectEdgeBusiness, Msg: "handle adding task failed", Data: resp},
			msg)
	}
	hwlog.RunLog.Info("handle task creation successful")
	return sendResponse(common.RespMsg{Status: common.Success}, msg)
}

func (h *createTaskHandler) prepareUpload(nodeSn string) logcollect.UploadConfig {
	fileName := logcollect.GetLogPackFileName(logcollect.ModuleEdge, nodeSn)
	url := fmt.Sprintf("https://%s:%d%s/%s", h.ip, &h, constants.UploadUrlPathPrefix, fileName)
	return logcollect.UploadConfig{
		MethodAndUrl: logcollect.MethodAndUrl{
			Method: http.MethodPost,
			Url:    url,
		},
	}
}

func (h *createTaskHandler) sendReqToEdge(edgeNode string, config logcollect.UploadConfig) error {
	msg, err := model.NewMessage()
	if err != nil {
		return err
	}
	msg.SetRouter(common.LogManagerName, common.CloudHubName, common.OptPost, common.ResLogEdge)
	msg.SetNodeId(edgeNode)
	msg.Content = config
	_, err = modulemanager.SendSyncMessage(msg, sendMessageTimeout)
	return err
}

func (h *createTaskHandler) parse(content interface{}) (logcollect.CreateTaskReq, error) {
	var req logcollect.CreateTaskReq
	return req, common.ParamConvert(content, &req)
}

func (h *createTaskHandler) check(req logcollect.CreateTaskReq) error {
	checkResult := getBatchQueryChecker().Check(req)
	if !checkResult.Result {
		return errors.New(checkResult.Reason)
	}
	var emptyServer logcollect.UploadConfig
	if req.HttpsServer == emptyServer {
		return nil
	}
	if req.HttpsServer.MethodAndUrl.Method != http.MethodPost {
		return errors.New("method not allowed")
	}
	checkResult = checker.GetRegChecker("Url", `^https://`, true).Check(req.HttpsServer.MethodAndUrl)
	if !checkResult.Result {
		return errors.New("schema is not supported")
	}
	return nil
}

func (h *createTaskHandler) Parse(*model.Message) error {
	return nil
}

func (h *createTaskHandler) Check(*model.Message) error {
	return nil
}

func (h *createTaskHandler) PrintOpLogOk() {
}

func (h *createTaskHandler) PrintOpLogFail() {
}
