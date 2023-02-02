// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager handler
package edgemsgmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/modulemanager"
	"huawei.com/mindxedge/base/modulemanager/model"
)

func sendMessageToEdge(msg *model.Message, content string) error {
	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("edge msg manager new message failed, error: %v", err)
		return fmt.Errorf("edge msg manager new message failed, error: %v", err)
	}

	respMsg.SetNodeId(msg.GetNodeId())
	respMsg.FillContent(content)
	respMsg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, msg.GetResource())

	if err = modulemanager.SendMessage(respMsg); err != nil {
		hwlog.RunLog.Errorf("edge msg manager send message failed, error: %v", err)
		return fmt.Errorf("edge msg manager send message failed, error: %v", err)
	}

	return nil
}

func sendResponse(msg *model.Message, resp string) error {
	newResponse, err := msg.NewResponse()
	if err != nil {
		hwlog.RunLog.Errorf("edge msg manager new response failed, error: %v", err)
		return fmt.Errorf("edge msg manager new response failed, error: %v", err)
	}

	newResponse.FillContent(resp)
	newResponse.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, msg.GetResource())

	if err = modulemanager.SendAsyncMessage(newResponse); err != nil {
		hwlog.RunLog.Errorf("edge msg manager send sync message failed, error: %v", err)
		return fmt.Errorf("edge msg manager send sync message failed, error: %v", err)
	}

	return nil
}

// UpgradeEdgeSoftware [method] upgrade edge software
func UpgradeEdgeSoftware(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start update edge software")
	var req EdgeUpgradeInfoReq
	var err error
	if err = common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
		return common.RespMsg{Status: common.ErrorNewMsg, Msg: "create message failed", Data: nil}
	}

	msg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, msg.GetResource())
	msg.FillContent(input)
	var batchResp types.BatchResp
	for _, sn := range req.SNs {
		msg.SetNodeId(sn)

		err = modulemanager.SendMessage(msg)
		if err != nil {
			batchResp.FailedIDs = append(batchResp.FailedIDs, sn)
		} else {
			batchResp.SuccessIDs = append(batchResp.SuccessIDs, sn)
		}
	}

	if len(batchResp.FailedIDs) != 0 {
		hwlog.RunLog.Info("deal edge software upgrade info failed")
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "", Data: batchResp}
	} else {
		hwlog.RunLog.Info("deal edge software upgrade info success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: batchResp}
	}
}

// EffectEdgeSoftware [method] effect edge software
func EffectEdgeSoftware(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start update edge software")
	var req EffectInfoReq
	var err error
	if err = common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
		return common.RespMsg{Status: common.ErrorNewMsg, Msg: "create message failed", Data: nil}
	}

	msg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, msg.GetResource())
	msg.FillContent(input)
	var batchResp types.BatchResp
	for _, sn := range req.SNs {
		msg.SetNodeId(sn)

		err = modulemanager.SendMessage(msg)
		if err != nil {
			batchResp.FailedIDs = append(batchResp.FailedIDs, sn)
		} else {
			batchResp.SuccessIDs = append(batchResp.SuccessIDs, sn)
		}
	}

	if len(batchResp.FailedIDs) != 0 {
		hwlog.RunLog.Info("deal edge software effect info failed")
		return common.RespMsg{Status: common.ErrorSendMsgToNode, Msg: "", Data: batchResp}
	} else {
		hwlog.RunLog.Info("deal edge software effect info success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: batchResp}
	}
}

func getNodesVersionInfo(nodeNames []string) (map[string]map[string]string, error) {
	if len(nodeNames) == 0 {
		hwlog.RunLog.Warn("node names is nil, get version info failed")
		return map[string]map[string]string{}, nil
	}
	router := common.Router{
		Source:      common.NodeMsgManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.Node,
	}
	req := types.InnerGetNodesInfoByNameReq{
		UniqueNames: nodeNames,
	}
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status != common.Success {
		return map[string]map[string]string{}, errors.New(resp.Msg)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		return map[string]map[string]string{}, errors.New("marshal internal response error")
	}

	var nodeVersionInfos []types.InnerGetNodeInfoByNameResp
	if err = json.Unmarshal(data, &nodeVersionInfos); err != nil {
		return map[string]map[string]string{}, errors.New("unmarshal internal response error")
	}

	var res = make(map[string]map[string]string)
	for _, nodeVersion := range nodeVersionInfos {
		res[nodeVersion.UniqueName] = nodeVersion.SoftwareInfos
	}
	return res, nil
}

// QueryEdgeSoftwareVersion [method] query edge software version
func QueryEdgeSoftwareVersion(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query edge software version")

	var req SoftwareVersionInfoReq
	var err error
	if err = common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	nodeVersionInfo, err := getNodesVersionInfo(req.SNs)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetNodesVersion, Msg: "", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: nodeVersionInfo}
}
