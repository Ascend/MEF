// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package restfulservice to init restful service
package restfulservice

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"huawei.com/mindx/common/hwlog"

	"edge-manager/pkg/nodemanager"
	"huawei.com/mindxedge/base/common"
)

func createEdgeNode(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("create node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Create,
		Resource:    common.Node,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listNodeUnManaged(c *gin.Context) {
	input, err := pageUtil(c)
	if err != nil {
		hwlog.OpLog.Error("list node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.List,
		Resource:    common.NodeUnManaged,
	}
	resp := common.SendSyncMessageByRestful(input, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func addUnManagedNode(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("add unmanaged node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Add,
		Resource:    common.Node,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listNodeManaged(c *gin.Context) {
	input, err := pageUtil(c)
	if err != nil {
		hwlog.OpLog.Error("list node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.List,
		Resource:    common.Node,
	}
	resp := common.SendSyncMessageByRestful(input, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func createEdgeNodeGroup(c *gin.Context) {
	reqContent, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("create node group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Create,
		Resource:    common.NodeGroup,
	}
	resp := common.SendSyncMessageByRestful(string(reqContent), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getNodeDetail(c *gin.Context) {
	var req nodemanager.GetNodeDetailReq
	if err := c.BindUri(&req); err != nil {
		hwlog.OpLog.Error("get node detail: get input parameter failed")
		common.ConstructResp(c, common.ErrorParamInvalid, err.Error(), nil)
		return
	}
	res, err := json.Marshal(req)
	if err != nil {
		hwlog.OpLog.Error("get node detail: marshal parameter failed")
		common.ConstructResp(c, common.ErrorParamInvalid, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Get,
		Resource:    common.Node,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func modifyNode(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("modify node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Update,
		Resource:    common.Node,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func modifyNodeGroup(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("modify node group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Update,
		Resource:    common.NodeGroup,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func deleteNode(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("delete node: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Delete,
		Resource:    common.Node,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func addNodeToGroup(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("add node to group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Add,
		Resource:    common.NodeRelation,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func deleteNodeFromGroup(c *gin.Context) {
	var req nodemanager.DeleteNodeToGroupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		hwlog.OpLog.Error("delete node from group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Delete,
		Resource:    common.NodeRelation,
	}
	var batchDeleteReq nodemanager.BatchDeleteNodeRelationReq
	for _, nodeID := range req.NodeIDs {
		deleteReq := nodemanager.DeleteNodeRelationReq{
			GroupID: req.GroupID,
			NodeID:  nodeID,
		}
		batchDeleteReq = append(batchDeleteReq, deleteReq)
	}
	res, err := json.Marshal(batchDeleteReq)
	if err != nil {
		hwlog.OpLog.Error("delete node from group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func batchDeleteNodeRelation(c *gin.Context) {
	res, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("delete node from group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Delete,
		Resource:    common.NodeRelation,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getNodeStatistics(c *gin.Context) {
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Get,
		Resource:    common.NodeStatistics,
	}
	resp := common.SendSyncMessageByRestful("", &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getNodeGroupStatistics(c *gin.Context) {
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Get,
		Resource:    common.NodeGroupStatistics,
	}
	resp := common.SendSyncMessageByRestful("", &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func listEdgeNodeGroup(c *gin.Context) {
	input, err := pageUtil(c)
	if err != nil {
		hwlog.OpLog.Error("list node group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, "", nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.List,
		Resource:    common.NodeGroup,
	}
	resp := common.SendSyncMessageByRestful(input, &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func getEdgeNodeGroupDetail(c *gin.Context) {
	var req nodemanager.GetNodeGroupDetailReq
	if err := c.BindUri(&req); err != nil {
		hwlog.OpLog.Error("get node group detail: get input parameter failed")
		common.ConstructResp(c, common.ErrorParamInvalid, err.Error(), nil)
		return
	}
	res, err := json.Marshal(req)
	if err != nil {
		hwlog.OpLog.Error("get node group detail: marshal parameter failed")
		common.ConstructResp(c, common.ErrorParamInvalid, err.Error(), nil)
		return
	}
	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Get,
		Resource:    common.NodeGroup,
	}
	resp := common.SendSyncMessageByRestful(string(res), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}

func batchDeleteNodeGroup(c *gin.Context) {
	reqContent, err := c.GetRawData()
	if err != nil {
		hwlog.OpLog.Error("batch delete node group: get input parameter failed")
		common.ConstructResp(c, common.ErrorParseBody, err.Error(), nil)
		return
	}

	router := common.Router{
		Source:      common.RestfulServiceName,
		Destination: common.NodeManagerName,
		Option:      common.Delete,
		Resource:    common.NodeGroup,
	}
	resp := common.SendSyncMessageByRestful(string(reqContent), &router)
	common.ConstructResp(c, resp.Status, resp.Msg, resp.Data)
}
