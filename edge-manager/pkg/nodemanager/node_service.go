// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"edge-manager/pkg/common"
	"edge-manager/pkg/util"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"
)

// CreateNode Create Node
func CreateNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node")
	req, ok := input.(util.CreateEdgeNodeReq)
	if !ok {
		hwlog.RunLog.Error("create node conver request error")
		return common.RespMsg{Status: "", Msg: "conver request error", Data: nil}
	}
	total, err := GetTableCount(NodeInfo{})
	if err != nil {
		hwlog.RunLog.Error("get node table num failed")
		return common.RespMsg{Status: "", Msg: "get node table num failed", Data: nil}
	}
	if total >= MaxNode {
		hwlog.RunLog.Error("node number is enough, connot create")
		return common.RespMsg{Status: "", Msg: "node number is enough, connot create", Data: nil}
	}
	node := &NodeInfo{
		Description: req.Description,
		UniqueName:  req.UniqueName,
		NodeName:    req.NodeName,
		Status:      statusOffline,
		IsManaged:   true,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdateAt:    time.Now().Format(TimeFormat),
	}
	if err = NodeServiceInstance().CreateNode(node); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("node name is duplicate")
			return common.RespMsg{Status: "", Msg: "node name is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("node db create failed")
		return common.RespMsg{Status: "", Msg: "db create failed", Data: nil}
	}
	hwlog.RunLog.Info("node db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
