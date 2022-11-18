// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node database table
package nodemanager

import (
	"edge-manager/pkg/common"
	"edge-manager/pkg/util"
	"strings"
	"time"

	"huawei.com/mindx/common/hwlog"
)

// CreateGroup Create Node Group
func createGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node group")
	var req util.CreateNodeGroupReq
	if err := common.ParamConvert(input, &req); err != nil {
		hwlog.RunLog.Error("create node group conver request error")
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	total, err := GetTableCount(NodeGroup{})
	if err != nil {
		hwlog.RunLog.Error("get node group table num failed")
		return common.RespMsg{Status: "", Msg: "get node group table num failed", Data: nil}
	}
	if total >= maxNodeGroup {
		hwlog.RunLog.Error("node group number is enough, connot create")
		return common.RespMsg{Status: "", Msg: "node group number is enough, connot create", Data: nil}
	}
	group := &NodeGroup{
		Description: req.Description,
		GroupName:   req.NodeGroupName,
		CreatedAt:   time.Now().Format(TimeFormat),
		UpdateAt:    time.Now().Format(TimeFormat),
	}
	if err = NodeServiceInstance().createNodeGroup(group); err != nil {
		if strings.Contains(err.Error(), common.ErrDbUniqueFailed) {
			hwlog.RunLog.Error("node group is duplicate")
			return common.RespMsg{Status: "", Msg: "node group is duplicate", Data: nil}
		}
		hwlog.RunLog.Error("node group db create failed")
		return common.RespMsg{Status: "", Msg: "db group create failed", Data: nil}
	}
	hwlog.RunLog.Info("node group db create success")
	return common.RespMsg{Status: common.Success, Msg: "", Data: nil}
}
