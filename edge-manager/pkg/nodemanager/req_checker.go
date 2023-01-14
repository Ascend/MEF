// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"huawei.com/mindxedge/base/common/checker/checker"
)

const (
	fieldDescription   = "Description"
	fieldNodeName      = "NodeName"
	fieldUniqueName    = "UniqueName"
	fieldGroupName     = "GroupName"
	fieldNodeGroupName = "NodeGroupName"
	fieldID            = "ID"
	fieldNodeID        = "NodeID"
	fieldGroupID       = "GroupID"
	fieldNodeIDs       = "NodeIDs"
	fieldGroupIDs      = "GroupIDs"
)

func newCreateEdgeNodeChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		descriptionChecker(fieldDescription),
		nodeNameChecker(fieldNodeName),
		uniqueNameChecker(fieldUniqueName),
		optionalIDListChecker(fieldGroupIDs, idChecker("")),
	)
}

func newGetNodeDetailChecker() *checker.IntChecker {
	return idChecker(fieldID)
}

func newGetGroupDetailChecker() *checker.IntChecker {
	return idChecker(fieldID)
}

func newModifyNodeChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		idChecker(fieldNodeID),
		nodeNameChecker(fieldNodeName),
		descriptionChecker(fieldDescription),
	)
}

func newBatchDeleteNodeChecker() *checker.UniqueListChecker {
	return idListChecker("", idChecker(""))
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

func newAddUnManagedNodeChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		idChecker(fieldNodeID),
		nodeNameChecker(fieldNodeName),
		optionalIDListChecker(fieldGroupIDs, idChecker("")),
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
