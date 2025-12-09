// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package edgemsgmanager common func
package edgemsgmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"

	"edge-manager/pkg/types"

	"huawei.com/mindxedge/base/common"
)

func sendMessageToEdge(msg *model.Message, content string) error {
	respMsg, err := model.NewMessage()
	if err != nil {
		hwlog.RunLog.Errorf("edge msg manager new message failed, error: %v", err)
		return fmt.Errorf("edge msg manager new message failed, error: %v", err)
	}

	respMsg.SetNodeId(msg.GetNodeId())
	if err = respMsg.FillContent(content); err != nil {
		hwlog.RunLog.Errorf("fill content failed: %v", err)
		return errors.New("fill content failed")
	}
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
	if err = respMsg.FillContent(content); err != nil {
		return fmt.Errorf("fill content failed: %v", err)
	}
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
	resp := common.SendSyncMessageByRestful(req, &router, common.ResponseTimeout)
	if resp.Status != common.Success {
		hwlog.RunLog.Errorf("get node info failed: %s", resp.Msg)
		return nil, errors.New(resp.Msg)
	}

	data, err := json.Marshal(resp.Data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal internal response error: %v", err)
		return nil, errors.New("marshal internal response error")
	}

	if err = json.Unmarshal(data, &nodeSoftwareInfo); err != nil {
		hwlog.RunLog.Errorf("unmarshal internal response error: %v", err)
		return nil, errors.New("unmarshal internal response error")
	}

	return nodeSoftwareInfo.SoftwareInfo, nil
}
