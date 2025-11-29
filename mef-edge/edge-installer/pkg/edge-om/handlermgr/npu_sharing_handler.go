// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package handlermgr this file for setting npu sharing switch (open or close)
package handlermgr

import (
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

var openNpuFailTip = config.ProgressTip{
	Topic:      "npu_sharing",
	Percentage: "0%",
	Result:     constants.ResultFailed,
	Reason:     "Open NPU sharing failed",
}

var closeNpuFailTip = config.ProgressTip{
	Topic:      "npu_sharing",
	Percentage: "0%",
	Result:     constants.ResultFailed,
	Reason:     "Close NPU sharing failed",
}

var npuSuccessTip = config.ProgressTip{
	Topic:      "npu_sharing",
	Percentage: "100%",
	Result:     constants.ResultSuccess,
	Reason:     "",
}

type npuSharingHandler struct {
	topic string
}

func (n *npuSharingHandler) runLog(open bool) {
	operation := constants.Close
	if open {
		operation = constants.Open
	}
	hwlog.RunLog.Infof("%s npu_sharing start", operation)
}

func (n *npuSharingHandler) runResultLog(open bool, success bool) {
	operation := constants.Close
	if open {
		operation = constants.Open
	}
	result := constants.Success
	if !success {
		result = constants.Failed
	}
	hwlog.RunLog.Infof("%s npu_sharing %s", operation, result)
}

func (n *npuSharingHandler) Handle(msg *model.Message) error {
	var mapContent map[string]interface{}
	if err := msg.ParseContent(&mapContent); err != nil {
		hwlog.RunLog.Errorf("parse content failed: %v", err)
		return errors.New("parse content failed")
	}
	objWrapper := util.NewWrapper(mapContent)
	npuSharingVal, err := objWrapper.GetBool("npu_sharing_enabled")
	if err != nil {
		hwlog.RunLog.Errorf("npu sharing handler get bool val error: %s", err.Error())
		return err
	}
	n.runLog(npuSharingVal)
	err = config.GetCapabilityMgr().Switch("npu_sharing", npuSharingVal)
	if err == nil {
		n.runResultLog(npuSharingVal, true)
		n.sendResponse(npuSuccessTip)
		return nil
	}
	n.runResultLog(npuSharingVal, false)
	hwlog.RunLog.Errorf("switch npu sharing failed, %s", err.Error())
	if npuSharingVal {
		n.sendResponse(openNpuFailTip)
		return nil
	}
	n.sendResponse(closeNpuFailTip)
	return nil
}

func (n *npuSharingHandler) sendResponse(tip config.ProgressTip) {
	newResponse, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("get new response failed: %v", err)
		return
	}
	newResponse.Router.Option = constants.OptUpdate
	newResponse.Router.Resource = constants.ResConfigResult
	newResponse.Header.ID = newResponse.Header.Id
	newResponse.SetKubeEdgeRouter(constants.HardwareModule, constants.GroupHub,
		constants.OptUpdate, constants.ResConfigResult)
	if err = newResponse.FillContent(tip, true); err != nil {
		hwlog.RunLog.Errorf("fill progress tip into content failed: %v", err)
		return
	}
	err = sendHandlerReplyMsg(newResponse)
	if err != nil {
		hwlog.RunLog.Errorf("send npu sharing handler response failed: %v", err)
	}
}
