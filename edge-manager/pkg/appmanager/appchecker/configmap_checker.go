// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package appchecker checker for configmap
package appchecker

import (
	"math"

	"huawei.com/mindx/common/checker"
)

const (
	regexpCmName       = "^[a-z][a-zA-Z0-9-]{2,62}[a-zA-Z0-9]$"
	regexpDescription  = `^[\S ]{0,255}$`
	regexpCmContentKey = "^[a-zA-Z-]([a-zA-Z0-9-_]){0,62}$"

	minCmId           = 1
	maxCmId           = math.MaxInt64
	minCmContentCount = 0
	maxCmContentCount = 64
	minCmListSize     = 1
	maxCmListSize     = 1024
)

// NewCreateCmChecker get create cm checker
func NewCreateCmChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		cmNameChecker("ConfigmapName"),
		descriptionChecker("Description"),
		cmContentChecker("ConfigmapContent"),
	)
}

func cmNameChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpCmName, true)
}

func descriptionChecker(fieldName string) *checker.RegChecker {
	return checker.GetRegChecker(fieldName, regexpDescription, false)
}

func cmContentChecker(fieldName string) *checker.UniqueListChecker {
	return checker.GetUniqueListChecker(
		fieldName,
		checker.GetRegChecker("Name", regexpCmContentKey, true),
		minCmContentCount,
		maxCmContentCount,
		false,
	)
}

// NewDeleteCmChecker get delete cm checker
func NewDeleteCmChecker() *checker.UniqueListChecker {
	return idListChecker("ConfigmapIDs", idChecker(""))
}

func idListChecker(fieldName string, elemChecker *checker.UintChecker) *checker.UniqueListChecker {
	return checker.GetUniqueListChecker(
		fieldName,
		elemChecker,
		minCmListSize,
		maxCmListSize,
		true,
	)
}

func idChecker(fieldName string) *checker.UintChecker {
	return checker.GetUintChecker(fieldName, minCmId, maxCmId, true)
}

// NewQueryCmChecker get query cm checker
func NewQueryCmChecker() *checker.UintChecker {
	return idChecker("")
}

// NewUpdateCmChecker get update cm checker
func NewUpdateCmChecker() *checker.AndChecker {
	return NewCreateCmChecker()
}
