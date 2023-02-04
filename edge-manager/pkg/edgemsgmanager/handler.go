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
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Errorf("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	var req EdgeUpgradeInfoReq
	var err error
	if err = common.ParamConvert(message.GetContent(), &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
		return common.RespMsg{Status: common.ErrorNewMsg, Msg: "create message failed", Data: nil}
	}

	msg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, common.ResEdgeUpgradeInfo)
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

// effectEdgeSoftware [method] effect edge software
func effectEdgeSoftware(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start effect edge software")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Errorf("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	var req EffectInfoReq
	var err error
	if err = common.ParamConvert(message.GetContent(), &req); err != nil {
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: err.Error(), Data: nil}
	}

	msg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("create message failed")
		return common.RespMsg{Status: common.ErrorNewMsg, Msg: "create message failed", Data: nil}
	}

	msg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptPost, common.ResEdgeEffectInfo)
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

func getNodeInfo(UniqueName string) (types.InnerGetNodeInfoByNameResp, error) {
	var nodeInfo types.InnerGetNodeInfoByNameResp
	router := common.Router{
		Source:      common.NodeMsgManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.Node,
	}
	req := types.InnerGetNodeInfoByNameReq{
		UniqueName: UniqueName,
	}
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status != common.Success {
		hwlog.RunLog.Errorf("get node info failed:%s", resp.Msg)
		return nodeInfo, errors.New(resp.Msg)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		hwlog.RunLog.Error("marshal internal response error")
		return nodeInfo, errors.New("marshal internal response error")
	}

	if err = json.Unmarshal(data, &nodeInfo); err != nil {
		hwlog.RunLog.Error("unmarshal internal response error")
		return nodeInfo, errors.New("unmarshal internal response error")
	}

	return nodeInfo, nil
}

func getNodeVersionInfo(nodeName string) (map[string]map[string]string, error) {
	nodeInfo, err := getNodeInfo(nodeName)
	if err != nil {
		hwlog.RunLog.Error("get node version failed")
		return map[string]map[string]string{}, errors.New("get node version failed")
	}

	return nodeInfo.SoftwareInfo, nil
}

func getNodeUpgradeProgressInfo(nodeName string) (string, error) {
	nodeInfo, err := getNodeInfo(nodeName)
	if err != nil {
		hwlog.RunLog.Error("get node upgrade progress failed")
		return "", errors.New("get node upgrade progress failed")
	}

	return nodeInfo.UpgradeProgress, nil
}

// queryEdgeSoftwareVersion [method] query edge software version
func queryEdgeSoftwareVersion(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query edge software version")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	uniqueName, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("query edge software version failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query edge software version " +
			"request convert error", Data: nil}
	}

	nodeVersionInfo, err := getNodeVersionInfo(uniqueName)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetNodeVersion, Msg: "", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: nodeVersionInfo}
}

// queryEdgeSoftwareUpgradeProgress [method] query edge software upgrade progress
func queryEdgeSoftwareUpgradeProgress(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start query edge software upgrade progress")
	message, ok := input.(*model.Message)
	if !ok {
		hwlog.RunLog.Error("get message failed")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "get message failed", Data: nil}
	}

	uniqueName, ok := message.GetContent().(string)
	if !ok {
		hwlog.RunLog.Error("query edge software upgrade progress failed: para type not valid")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "query edge software upgrade progress" +
			" convert error", Data: nil}
	}

	upgradeProgress, err := getNodeUpgradeProgressInfo(uniqueName)
	if err != nil {
		return common.RespMsg{Status: common.ErrorGetNodesUpgradeProgress, Msg: "", Data: nil}
	}

	return common.RespMsg{Status: common.Success, Msg: "", Data: upgradeProgress}
}
