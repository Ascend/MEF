// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers provides handlers to process business logic of log collection
package handlers

import (
	"huawei.com/mindx/common/checker"

	"huawei.com/mindxedge/base/common/logmgmt/logcollect"
)

const (
	minList = 1
	maxList = 16

	regexpNodeSn = `^[a-zA-Z0-9]([-_a-zA-Z0-9]{0,62}[a-zA-Z0-9])?$`
)

var (
	validModuleList = []string{logcollect.ModuleEdge}
)

func getBatchQueryChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		checker.GetUniqueListChecker(
			"EdgeNodes", checker.GetRegChecker("", regexpNodeSn, true), minList, maxList, true),
		checker.GetStringChoiceChecker("Module", validModuleList, true),
	)
}
