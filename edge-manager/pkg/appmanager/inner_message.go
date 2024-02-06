// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package appmanager

import (
	"encoding/json"
	"errors"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr/model"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"huawei.com/mindxedge/base/common"

	"edge-manager/pkg/types"
)

func getAppInstanceCountByNodeGroup(msg *model.Message) common.RespMsg {
	hwlog.RunLog.Info("start to get appInstance count")
	var req []uint64
	if err := msg.ParseContent(&req); err != nil {
		hwlog.RunLog.Errorf("failed to parse param: %v", err)
		return common.RespMsg{Status: common.ErrorParamConvert, Msg: "parse content failed"}
	}
	appInstanceCount := make(map[uint64]int64)
	for _, groupId := range req {
		count, err := AppRepositoryInstance().countDeployedAppByGroupID(groupId)
		if err != nil {
			hwlog.RunLog.Error("failed to count appInstance by node group")
			return common.RespMsg{Status: common.ErrorGetAppInstanceCountByNodeGroup, Msg: ""}
		}
		appInstanceCount[groupId] = count
	}
	hwlog.RunLog.Info("get appInstance count success")
	return common.RespMsg{Status: common.Success, Data: appInstanceCount}
}

func checkNodeGroupResource(groupID uint64, daemonSet *appv1.DaemonSet) error {
	router := common.Router{
		Source:      common.AppManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.CheckResource,
	}
	req := types.InnerCheckNodeResReq{
		NodeGroupID:  groupID,
		ResourceReqs: getAppResReqs(daemonSet),
	}
	resp := common.SendSyncMessageByRestful(req, &router, common.ResponseTimeout)
	if resp.Status != common.Success {
		return errors.New(resp.Msg)
	}
	return nil
}

func updateAllocatedNodeRes(daemonSet *appv1.DaemonSet, nodeGroupID uint64, isUndeploy bool) error {
	router := common.Router{
		Source:      common.AppManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.UpdateResource,
	}
	req := types.InnerUpdateNodeResReq{
		NodeGroupID:  nodeGroupID,
		ResourceReqs: getAppResReqs(daemonSet),
		IsUndeploy:   isUndeploy,
	}
	resp := common.SendSyncMessageByRestful(req, &router, common.ResponseTimeout)
	if resp.Status != common.Success {
		return errors.New(resp.Msg)
	}
	return nil
}

func getAppResReqs(daemonSet *appv1.DaemonSet) corev1.ResourceList {
	appResReqs := make(map[corev1.ResourceName]resource.Quantity)
	if daemonSet == nil {
		return appResReqs
	}
	for _, container := range daemonSet.Spec.Template.Spec.Containers {
		for resName, quantity := range container.Resources.Limits {
			totalResReq := appResReqs[resName]
			totalResReq.Add(quantity)
			appResReqs[resName] = totalResReq
		}
	}
	return appResReqs
}

func getNodeGroupInfos(nodeGroupIds []uint64) ([]types.NodeGroupInfo, error) {
	router := common.Router{
		Source:      common.AppManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.NodeGroup,
	}
	req := types.InnerGetNodeGroupInfosReq{
		NodeGroupIds: nodeGroupIds,
	}
	resp := common.SendSyncMessageByRestful(req, &router, common.ResponseTimeout)
	var nodeGroupInfosResp types.InnerGetNodeGroupInfosResp
	if err := parseDataFromResp(resp, &nodeGroupInfosResp); err != nil {
		return nil, err
	}
	return nodeGroupInfosResp.NodeGroupInfos, nil
}

func getNodeInfoByUniqueName(eventPod *corev1.Pod) (uint64, string, error) {
	if eventPod.Spec.NodeName == "" {
		hwlog.RunLog.Warn("app instance node name is empty, pod is in pending phase")
		return 0, "", nil
	}
	router := common.Router{
		Source:      common.AppManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.Node,
	}
	req := types.InnerGetNodeInfoByNameReq{
		UniqueName: eventPod.Spec.NodeName,
	}
	resp := common.SendSyncMessageByRestful(req, &router, common.ResponseTimeout)
	var nodeInfo types.InnerGetNodeInfoByNameResp
	if err := parseDataFromResp(resp, &nodeInfo); err != nil {
		return 0, "", err
	}
	return nodeInfo.NodeID, nodeInfo.NodeName, nil
}

func getNodeStatus(nodeUniqueName string) (string, error) {
	if nodeUniqueName == "" {
		hwlog.RunLog.Warn("app instance node name is empty, pod is in pending phase")
		return nodeStatusUnknown, nil
	}
	router := common.Router{
		Source:      common.AppManagerName,
		Destination: common.NodeManagerName,
		Option:      common.Inner,
		Resource:    common.NodeStatus,
	}
	req := types.InnerGetNodeStatusReq{
		UniqueName: nodeUniqueName,
	}
	resp := common.SendSyncMessageByRestful(req, &router, common.ResponseTimeout)
	var node types.InnerGetNodeStatusResp
	if err := parseDataFromResp(resp, &node); err != nil {
		return "", err
	}
	return node.NodeStatus, nil
}

func parseDataFromResp(resp common.RespMsg, v interface{}) error {
	if resp.Status != common.Success {
		return errors.New(resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return errors.New("marshal internal response error")
	}
	if err = json.Unmarshal(data, v); err != nil {
		return errors.New("unmarshal internal response error")
	}
	return nil
}
