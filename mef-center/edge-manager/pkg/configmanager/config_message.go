// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package configmanager to message config manager
package configmanager

import (
	"encoding/json"
	"errors"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/types"
)

func getAllNodeInfo() ([]nodemanager.NodeInfo, error) {
	router := common.Router{
		Source:      common.AppManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.NodeList,
	}
	req := types.InnerGetNodeInfoResReq{
		ModuleName: common.ConfigManagerName,
	}
	resp := common.SendSyncMessageByRestful(req, &router, common.ResponseTimeout)
	if resp.Status != common.Success {
		return nil, errors.New(resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, errors.New("marshal internal response error")
	}
	var nodes []nodemanager.NodeInfo
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nodes, errors.New("unmarshal internal response error")
	}
	return nodes, nil
}
