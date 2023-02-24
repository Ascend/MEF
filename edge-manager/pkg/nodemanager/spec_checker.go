// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node database table
package nodemanager

import (
	"errors"

	"huawei.com/mindxedge/base/common"
)

type specificationChecker struct {
	nodeService NodeService
}

func (checker specificationChecker) checkAddNodeToGroup(nodeIds, groupIds []uint64) error {
	for _, groupId := range groupIds {
		total, err := checker.nodeService.countNodeByGroup(groupId)
		if err != nil {
			return errors.New("get node in group table nodePerGroup failed")
		}
		nodePerGroup := total + int64(len(nodeIds))
		if nodePerGroup > common.MaxNodePerGroup {
			return errors.New("node in group number is enough, cannot join")
		}
	}
	for _, nodeId := range nodeIds {
		total, err := checker.nodeService.countGroupsByNode(nodeId)
		if err != nil {
			return errors.New("get group in node table groupPerNode failed")
		}
		groupPerNode := total + int64(len(groupIds))
		if groupPerNode > common.MaxGroupPerNode {
			return errors.New("group number of node is enough, cannot join")
		}
	}
	return nil
}

func (checker specificationChecker) checkAddNodes(addCount int) error {
	total, err := GetTableCount(NodeInfo{})
	if err != nil {
		return errors.New("get node table num failed")
	}
	if int64(total)+int64(addCount) > int64(common.MaxNode) {
		return errors.New("node number is enough, cannot create")
	}
	return nil
}

func (checker specificationChecker) checkAddGroups(addCount int) error {
	total, err := GetTableCount(NodeGroup{})
	if err != nil {
		return errors.New("get group table num failed")
	}
	if int64(total)+int64(addCount) > int64(common.MaxNodeGroup) {
		return errors.New("group number is enough, cannot create")
	}
	return nil
}
