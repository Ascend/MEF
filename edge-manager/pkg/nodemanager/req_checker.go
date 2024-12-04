// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"huawei.com/mindx/common/checker"

	"huawei.com/mindxedge/base/common"
)

const (
	fieldDescription   = "Description"
	fieldNodeName      = "NodeName"
	fieldUniqueName    = "UniqueName"
	fieldSerialNumber  = "SerialNumber"
	fieldGroupName     = "GroupName"
	fieldNodeGroupName = "NodeGroupName"
	fieldIP            = "IP"
	fieldNodeID        = "NodeID"
	fieldGroupID       = "GroupID"
	fieldNodeIDs       = "NodeIDs"
	fieldGroupIDs      = "GroupIDs"
)

func newGetNodeDetailIdChecker() *checker.UintChecker {
	return idChecker("")
}

func newModifyNodeChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		idChecker(fieldNodeID),
		nodeNameChecker(fieldNodeName),
		descriptionChecker(fieldDescription),
	)
}

func newBatchDeleteNodeChecker() *checker.UniqueListChecker {
	return idListChecker(fieldNodeIDs, idChecker(""))
}

func newBatchDeleteNodeRelationChecker() *checker.UniqueListChecker {
	return uniqueListChecker(
		"",
		checker.GetAndChecker(
			idChecker(fieldGroupID),
			idChecker(fieldNodeID),
		),
	)
}

func newAddNodeRelationChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		idChecker(fieldGroupID),
		idListChecker(fieldNodeIDs, idChecker("")),
	)
}

func newNodeInfoChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		nodeNameChecker(fieldNodeName),
		uniqueNameChecker(fieldUniqueName),
		nodeSerialNumberChecker(fieldSerialNumber),
		checker.GetIpV4Checker(fieldIP, true))
}

func newDeleteNodeFromGroupChecker() *checker.AndChecker {
	return newAddNodeRelationChecker()
}

func newAddUnManagedNodeChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		idChecker(fieldNodeID),
		nodeNameChecker(fieldNodeName),
		optionalIDListChecker(fieldGroupIDs, common.MaxGroupPerNode, idChecker("")),
		descriptionChecker(fieldDescription),
	)
}

func newCreateGroupChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		groupNameChecker(fieldNodeGroupName),
		descriptionChecker(fieldDescription),
	)
}

func newModifyGroupChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		idChecker(fieldGroupID),
		groupNameChecker(fieldGroupName),
		descriptionChecker(fieldDescription),
	)
}

func newBatchDeleteGroupChecker() *checker.UniqueListChecker {
	return idListChecker(fieldGroupIDs, idChecker(""))
}
