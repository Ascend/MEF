// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
