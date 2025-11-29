// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlers
package handlers

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"k8s.io/api/core/v1"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/edge-main/common"
	"edge-installer/pkg/edge-main/msgconv/statusmanager"
)

type podRestartHandler struct{}

// Handle podRestartHandler handle entry
func (pr *podRestartHandler) Handle(msg *model.Message) error {
	opResult := constants.Failed
	// defer anonymous function is used to pass the variable that changes dynamically, do not delete it
	defer func() {
		pr.PrintOpLog(msg, opResult)
	}()

	var podName string
	if err := msg.ParseContent(&podName); err != nil {
		pr.sendResponse(msg, "FAILED: parse data failed: "+err.Error())
		hwlog.RunLog.Errorf("parse pod name failed: %v", err)
		return errors.New("parse pod name failed")
	}

	msg.Router.Resource = msg.GetResource() + podName
	if err := CheckPodRestartPolicy(msg.Router.Resource); err != nil {
		pr.sendResponse(msg, "FAILED: check pod restart policy failed, "+err.Error())
		hwlog.RunLog.Errorf("check pod restart policy failed, error: %v", err)
		return errors.New("check pod restart policy failed")
	}

	result, err := util.SendSyncMsg(util.InnerMsgParams{
		Source:      constants.ModHandlerMgr,
		Destination: constants.ModEdgeOm,
		Operation:   constants.OptRestart,
		Resource:    constants.ActionPod,
		Content:     podName,
	})
	if err != nil {
		pr.sendResponse(msg, "FAILED: send restart pod message to edge om failed")
		hwlog.RunLog.Errorf("send restart pod message to edge om failed, error: %v", err)
		return errors.New("send restart pod message to edge om failed")
	}
	if result == constants.Failed {
		pr.sendResponse(msg, "FAILED: restart pod failed")
		hwlog.RunLog.Error("restart pod failed by edge om")
		return errors.New("restart pod failed by edge om")
	}

	opResult = constants.Success
	pr.sendResponse(msg, "OK")
	hwlog.RunLog.Info("restart pod success")
	return nil
}

func (pr *podRestartHandler) sendResponse(msg *model.Message, respMsg string) {
	newResp, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("get a new response message failed, error: %v", err)
		return
	}
	newResp.Header.IsSync = false
	newResp.Router.Destination = constants.ModDeviceOm
	newResp.Router.Option = constants.OptResponse
	newResp.SetKubeEdgeRouter(msg.GetSource(), msg.KubeEdgeRouter.Group, constants.OptResponse, msg.GetResource())
	newResponse, err := common.MsgOutProcess(newResp)
	if err != nil {
		hwlog.RunLog.Errorf("message out process failed, error: %v", err)
		return
	}
	if err = newResponse.FillContent(respMsg); err != nil {
		hwlog.RunLog.Errorf("fill resp content failed: %v", err)
		return
	}
	if err = modulemgr.SendAsyncMessage(newResponse); err != nil {
		hwlog.RunLog.Errorf("send pod restart handler response failed, error: %v", err)
		return
	}
}

func (pr *podRestartHandler) PrintOpLog(msg *model.Message, opResult string) {
	fdIp, err := common.GetFdIp()
	if err != nil {
		hwlog.RunLog.Warnf("get fd ip failed, error: %v", err)
	}
	hwlog.OpLog.Infof("[%s@%s] %s %s %s, the message(id:%s) is forwarded from [%s:%s]", constants.DeviceOmModule,
		constants.LocalIp, msg.GetOption(), msg.GetResource(), opResult, msg.Header.ID, constants.FD, fdIp)
}

// CheckPodRestartPolicy check pod restart policy before restart pod
func CheckPodRestartPolicy(podResource string) error {
	content, err := statusmanager.GetPodStatusMgr().Get(podResource)
	if err != nil {
		hwlog.RunLog.Errorf("get pod status failed, error: %v", err)
		return errors.New("get pod status failed")
	}

	var pod v1.Pod
	if err = json.Unmarshal([]byte(content), &pod); err != nil {
		hwlog.RunLog.Errorf("unmarshal pod status failed, error: %v", err)
		return errors.New("unmarshal pod status failed")
	}

	if pod.Spec.RestartPolicy == v1.RestartPolicyNever {
		hwlog.RunLog.Error("the restart policy of pod is Never, cannot restart the pod")
		return errors.New("the restart policy of pod is Never, cannot restart the pod")
	}
	return nil
}
