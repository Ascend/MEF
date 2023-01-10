// Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

// Package nodemanager to init node service
package nodemanager

import (
	"math"

	"huawei.com/mindxedge/base/common/checker/checker"
)

const (
	regexpNodeName    = `^[a-zA-Z][-_a-zA-Z0-9]{0,62}[a-zA-Z0-9]$`
	regexpGroupName   = `^[a-zA-Z]([_a-zA-Z0-9]{0,30}[a-zA-Z0-9])?$`
	regexpDescription = `^[\S ]{0,512}$`
	maxListSize       = 1024
)

func nodeNameChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpNodeName, true)
}

func groupNameChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpGroupName, true)
}

func uniqueNameChecker(fieldName string) *checker.ExistChecker {
	return checker.GetExistChecker(fieldName)
}

func descriptionChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpDescription, false)
}

func idChecker(fieldName string) *checker.IntChecker {
	return checker.GetIntChecker(fieldName, 1, math.MaxInt64, true)
}

func idListChecker(fieldName string, elemChecker *checker.IntChecker) *checker.UniqueListChecker {
	return checker.GetUniqueListChecker(
		fieldName,
		elemChecker,
		1,
		maxListSize,
		true,
	)
}

func optionalIDListChecker(fieldName string, elemChecker *checker.IntChecker) *checker.UniqueListChecker {
	return checker.GetUniqueListChecker(
		fieldName,
		elemChecker,
		0,
		maxListSize,
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
