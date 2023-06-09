// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package edgemsgmanager handler
package edgemsgmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/types"
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

	if err = modulemgr.SendMessage(respMsg); err != nil {
		hwlog.RunLog.Errorf("edge msg manager send message failed, error: %v", err)
		return fmt.Errorf("edge msg manager send message failed, error: %v", err)
	}

	return nil
}

func sendRespToEdge(msg *model.Message, content string) error {
	respMsg, err := msg.NewResponse()
	if err != nil {
		return err
	}

	respMsg.SetNodeId(msg.GetNodeId())
	respMsg.FillContent(content)
	respMsg.SetRouter(common.NodeMsgManagerName, common.CloudHubName, common.OptResp, msg.GetResource())

	return modulemgr.SendAsyncMessage(respMsg)
}

func getNodeSoftwareInfo(serialNumber string) ([]types.SoftwareInfo, error) {
	var nodeSoftwareInfo types.InnerSoftwareInfoResp
	router := common.Router{
		Source:      common.NodeMsgManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.NodeSoftwareInfo,
	}
	req := types.InnerGetSfwInfoBySNReq{
		SerialNumber: serialNumber,
	}
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status != common.Success {
		hwlog.RunLog.Errorf("get node info failed:%s", resp.Msg)
		return nil, errors.New(resp.Msg)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal internal response error %v", err)
		return nil, errors.New("marshal internal response error")
	}

	if err = json.Unmarshal(data, &nodeSoftwareInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal internal response error %v", err)
		return nil, errors.New("unmarshal internal response error")
	}

	return nodeSoftwareInfo.SoftwareInfo, nil
}
