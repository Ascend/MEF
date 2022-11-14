// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/common"
	"edge-manager/pkg/kubeclient"
	"edge-manager/pkg/util"
)

// CreateNode Create Node
func createNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start create node")
	var req util.CreateEdgeNodeReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	total, err := GetTableCount(NodeInfo{})
	if err != nil {
		hwlog.RunLog.Error("get node table num failed")
		return common.RespMsg{Status: "", Msg: "get node table num failed", Data: nil}
	}
	if total >= maxNode {
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
	if err = NodeServiceInstance().createNode(node); err != nil {
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

// ListNode get node list
func listNode(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start list node managed")
	var req util.ListReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}
	nodes, err := NodeServiceInstance().listNodesByName(req.PageNum, req.PageSize, req.Name)
	if err == nil {
		hwlog.RunLog.Info("list node success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: nodes}
	}
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("dont have any managed node")
		return common.RespMsg{Status: common.Success, Msg: "dont have any managed node", Data: nil}
	}
	hwlog.RunLog.Error("list node failed")
	return common.RespMsg{Status: "", Msg: "list node failed", Data: nil}
}

// ListNode get node list
func listNodeUnManaged(input interface{}) common.RespMsg {
	if err := autoAddUnmanagedNode(); err != nil {
		hwlog.RunLog.Error("auto add unmanaged node filed")
		return common.RespMsg{Status: "", Msg: "auto add unmanaged node filed", Data: nil}
	}
	hwlog.RunLog.Info("start list node unmanaged")
	var req util.ListReq
	if err := common.ParamConvert(input, &req); err != nil {
		return common.RespMsg{Status: "", Msg: err.Error(), Data: nil}
	}

	nodes, err := NodeServiceInstance().listUnManagedNodesByName(req.PageNum, req.PageSize, req.Name)
	if err == nil {
		hwlog.RunLog.Info("list node unmanaged success")
		return common.RespMsg{Status: common.Success, Msg: "", Data: nodes}
	}
	if err == gorm.ErrRecordNotFound {
		hwlog.RunLog.Info("dont have any unmanaged node")
		return common.RespMsg{Status: common.Success, Msg: "dont have any unmanaged node", Data: nil}
	}
	hwlog.RunLog.Error("list unmanaged node failed")
	return common.RespMsg{Status: "", Msg: "list unmanaged node failed", Data: nil}
}

func autoAddUnmanagedNode() error {
	realNodes, err := kubeclient.GetKubeClient().ListNode()
	if err != nil {
		return err
	}
	nodeDb, err := GetTableCount(NodeInfo{})
	if err != nil {
		hwlog.RunLog.Error("get node table num failed")
		return err
	}
	if len(realNodes.Items) == nodeDb {
		return nil
	}
	for _, node := range realNodes.Items {
		_, err := NodeServiceInstance().getNodeByUniqueName(node.Name)
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("get node by name(%s) failed", node.Name)
		}
		nodeInfo := &NodeInfo{
			NodeName:   node.Name,
			UniqueName: node.Name,
			Status:     statusOffline,
			IsManaged:  false,
			CreatedAt:  time.Now().Format(TimeFormat),
			UpdateAt:   time.Now().Format(TimeFormat),
		}
		if err := NodeServiceInstance().createNode(nodeInfo); err != nil {
			return err
		}
		hwlog.RunLog.Debugf("auto create unmanaged node %s", node.Name)
	}
	return nil
}
