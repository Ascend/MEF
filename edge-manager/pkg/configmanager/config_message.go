// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package configmanager to message config manager
package configmanager

import (
	"encoding/json"
	"errors"

	"edge-manager/pkg/nodemanager"
	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
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
	resp := common.SendSyncMessageByRestful(req, &router)
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
