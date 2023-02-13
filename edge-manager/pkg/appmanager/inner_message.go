// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

package appmanager

import (
	"encoding/json"
	"errors"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"edge-manager/pkg/types"
	"huawei.com/mindxedge/base/common"
)

func getAppInstanceCountByNodeGroup(input interface{}) common.RespMsg {
	hwlog.RunLog.Info("start to get appInstance count")
	req, ok := input.([]uint64)
	if !ok {
		hwlog.RunLog.Error("failed to convert param")
		return common.RespMsg{Status: common.ErrorTypeAssert, Msg: "failed to convert param"}
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

func checkNodeGroupResources(groupID uint64, daemonSet *appv1.DaemonSet) error {
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
	resp := common.SendSyncMessageByRestful(req, &router)
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
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status != common.Success {
		return errors.New(resp.Msg)
	}
	return nil
}

func getAppResReqs(daemonSet *appv1.DaemonSet) corev1.ResourceList {
	var appResReqs = make(map[corev1.ResourceName]resource.Quantity)
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
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status != common.Success {
		return nil, errors.New(resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, errors.New("marshal internal response error")
	}
	var nodeGroupInfosResp types.InnerGetNodeGroupInfosResp
	if err = json.Unmarshal(data, &nodeGroupInfosResp); err != nil {
		return nil, errors.New("unmarshal internal response error")
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
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status != common.Success {
		return 0, "", errors.New(resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return 0, "", errors.New("marshal internal response error")
	}
	var nodeInfo types.InnerGetNodeInfoByNameResp
	if err = json.Unmarshal(data, &nodeInfo); err != nil {
		return 0, "", errors.New("unmarshal internal response error")
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
	resp := common.SendSyncMessageByRestful(req, &router)
	if resp.Status != common.Success {
		return nodeStatusUnknown, fmt.Errorf("get info from other module error, %v", resp.Msg)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return nodeStatusUnknown, errors.New("marshal internal response error")
	}
	var node types.InnerGetNodeStatusResp
	if err = json.Unmarshal(data, &node); err != nil {
		return nodeStatusUnknown, errors.New("unmarshal internal response error")
	}
	return node.NodeStatus, nil
}
