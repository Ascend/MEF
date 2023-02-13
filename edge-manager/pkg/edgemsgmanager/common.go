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

func getNodeSoftwareInfo(serialNumber string) (map[string]map[string]string, error) {
	var nodeSoftwareInfo types.InnerSoftwareInfoResp
	router := common.Router{
		Source:      common.NodeMsgManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.Node,
	}
	req := types.InnerGetSoftwareInfoBySerialNumberReq{
		SerialNumber: serialNumber,
	}
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status != common.Success {
		hwlog.RunLog.Errorf("get node info failed:%s", resp.Msg)
		return nil, errors.New(resp.Msg)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		hwlog.RunLog.Error("marshal internal response error")
		return nil, errors.New("marshal internal response error")
	}

	if err = json.Unmarshal(data, &nodeSoftwareInfo); err != nil {
		hwlog.RunLog.Error("unmarshal internal response error")
		return nil, errors.New("unmarshal internal response error")
	}

	return nodeSoftwareInfo.SoftwareInfo, nil
}
