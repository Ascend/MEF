// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package edgemsgmanager to deal manage node msg
package edgemsgmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/checker/checker"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func getSftDownloadInfo(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to deal download info query")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	sfwName, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("get message content failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message content failed", Data: nil}
	}

	if checkResult := checker.GetStringChoiceChecker("", []string{common.EdgeCore, common.DevicePlugin},
		true).Check(sfwName); !checkResult.Result {
		hwlog.RunLog.Error("get software download info failed because invalid software name")
		return common.RespMsg{Status: common.ErrorParamInvalid, Msg: "get software download info failed", Data: nil}
	}

	sftDownloadInfo, err := innerGetSfwDownloadInfo(sfwName)
	if err != nil {
		hwlog.RunLog.Errorf("get [%s] software download info failed", sfwName)
		return common.RespMsg{Status: common.ErrorInnerGetData, Msg: "get software download info failed", Data: nil}
	}

	if err = sendMessageToEdge(message, string(sftDownloadInfo)); err != nil {
		hwlog.RunLog.Errorf("send message to edge failed, error: %v", err)
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "send msg to edge failed", Data: nil}
	}

	hwlog.RunLog.Infof("deal sft [%s] download info query success", sfwName)
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}

func innerGetSfwDownloadInfo(sfwName string) ([]byte, error) {
	router := common.Router{
		Source:      common.NodeMsgManagerName,
		Destination: common.SoftwareManagerName,
		Option:      common.Inner,
		Resource:    common.ResSfwDownloadInfo,
	}

	resp := common.SendSyncMessageByRestful(sfwName, &router)
	if resp.Status != common.Success {
		hwlog.RunLog.Errorf("get [%s] download info failed:%s", sfwName, resp.Msg)
		return []byte{}, fmt.Errorf("get [%s] download info failed", sfwName)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal internal response error %v", err)
		return nil, errors.New("marshal internal response error")
	}

	var softwareDownloadInfo SoftwareDownloadInfo
	softwareDownloadInfo.SoftwareName = sfwName

	if err = json.Unmarshal(data, &softwareDownloadInfo.DownloadInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal internal response error %v", err)
		return nil, errors.New("unmarshal internal response error")
	}

	data, err = json.Marshal(softwareDownloadInfo)
	if err != nil {
		hwlog.RunLog.Errorf("marshal sft download info error %v", err)
		return []byte{}, errors.New("marshal sft download info error")
	}

	return data, nil
}
