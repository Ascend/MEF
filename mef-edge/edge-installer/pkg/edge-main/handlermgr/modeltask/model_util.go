// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package modeltask

import (
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

// SendOkResponse send the ok resp to FD
func SendOkResponse(message *model.Message) {
	msg, err := createMsg(constants.ControllerModule, constants.HardwareModule,
		constants.OptReport, message.GetResource())
	if err != nil {
		hwlog.RunLog.Errorf("create response failed: %v", err)
		return
	}
	if err = msg.FillContent("OK"); err != nil {
		hwlog.RunLog.Errorf("fill ok content failed: %v", err)
		return
	}
	msg.Header.ParentID = message.Header.ID
	sendToFd(msg)
}

// SendFailResponse [method] send failed info to FD
func SendFailResponse(topic, reason string) {
	failTip := config.ProgressTip{
		Topic:      topic,
		Percentage: "0%",
		Result:     constants.ResultFailed,
		Reason:     reason,
	}
	msg, err := createMsg(constants.HardwareModule, constants.GroupHub, constants.OptUpdate, constants.ResConfigResult)
	if err != nil {
		hwlog.RunLog.Errorf("create response failed: %v", err)
		return
	}
	if err = msg.FillContent(failTip, true); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return
	}
	sendToFd(msg)
}

// SendConfigResult send the config result to FD
func SendConfigResult(topic string) {
	successTip := config.ProgressTip{
		Topic:      topic,
		Percentage: "100%",
		Result:     constants.Success,
		Reason:     "",
	}
	msg, err := createMsg(constants.HardwareModule, constants.GroupHub, constants.OptUpdate, constants.ResConfigResult)
	if err != nil {
		hwlog.RunLog.Errorf("create success response failed: %v", err)
		return
	}
	if err = msg.FillContent(successTip, true); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return
	}
	sendToFd(msg)
}

func createMsg(source, group, operation, resource string) (*model.Message, error) {
	newResponse, err := model.NewMessage()
	if err != nil {

		return nil, err
	}
	newResponse.Header.ID = newResponse.Header.Id
	newResponse.SetKubeEdgeRouter(source, group, operation, resource)
	newResponse.SetRouter(constants.ModHandlerMgr, constants.ModDeviceOm, operation, newResponse.GetResource())
	return newResponse, nil
}

func sendToFd(msg *model.Message) {
	err := modulemgr.SendAsyncMessage(msg)
	if err != nil {
		hwlog.RunLog.Errorf("send response failed: %v", err)
		return
	}
}

// SendReport send the info of model files to FD
func SendReport(reportData interface{}) {
	newResponse, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("get new response failed: %v", err)
		return
	}
	newResponse.Header.ID = newResponse.Header.Id
	newResponse.SetKubeEdgeRouter(constants.HardwareModule, constants.GroupHub,
		constants.OptUpdate, constants.ActionModelFileInfo)
	if err = newResponse.FillContent(reportData, true); err != nil {
		hwlog.RunLog.Errorf("fill report into content failed: %v", err)
		return
	}
	newResponse.SetRouter(constants.ModHandlerMgr, constants.ModDeviceOm,
		newResponse.GetOption(), newResponse.GetResource())
	err = modulemgr.SendAsyncMessage(newResponse)
	if err != nil {
		hwlog.RunLog.Errorf("send model file report failed: %v", err)
		return
	}
}

// BuildModelStatus [method] creat model status
func BuildModelStatus(statusClass StatusIntf, task *ModelFileTask) StatusIntf {
	if statusClass == nil {
		hwlog.RunLog.Errorf("create status error, status interface nil")
		return buildFailStatus(task, "create error")
	}

	baseStatus := &BaseTaskStatus{}
	baseStatus.Task = task
	statusClass.setBaseStatus(baseStatus)
	return statusClass
}

func buildFailStatus(task *ModelFileTask, reason string) *FailStatus {
	failStatus := &FailStatus{
		BaseTaskStatus: &BaseTaskStatus{
			Task: task,
		},
		reason: reason,
	}
	return failStatus
}
