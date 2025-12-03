// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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
	"math"

	"huawei.com/mindx/common/checker"
)

const (
	regexpNodeName         = `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`
	regUniqueName          = `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`
	regexpNodeSerialNumber = `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`
	regexpGroupName        = `^[a-zA-Z]([_a-zA-Z0-9]{0,30}[a-zA-Z0-9])?$`
	regexpDescription      = `^[\S ]{0,512}$`
	maxListSize            = 1024
)

func nodeNameChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpNodeName, true)
}

func groupNameChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpGroupName, true)
}

func uniqueNameChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regUniqueName, true)
}

func nodeSerialNumberChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpNodeSerialNumber, true)
}

func descriptionChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpDescription, false)
}

func idChecker(fieldName string) *checker.UintChecker {
	return checker.GetUintChecker(fieldName, 1, math.MaxUint32, true)
}

func idListChecker(fieldName string, elemChecker *checker.UintChecker) *checker.UniqueListChecker {
	return checker.GetUniqueListChecker(
		fieldName,
		elemChecker,
		1,
		maxListSize,
		true,
	)
}

func optionalIDListChecker(fieldName string, maxLen int, elemChecker *checker.UintChecker) *checker.UniqueListChecker {
	return checker.GetUniqueListChecker(
		fieldName,
		elemChecker,
		0,
		maxLen,
		false,
	)
}

func uniqueListChecker(fieldName string, elemChecker *checker.AndChecker) *checker.UniqueListChecker {
	return checker.GetUniqueListChecker(
		fieldName,
		elemChecker,
		1,
		maxListSize,
		true,
	)
}
