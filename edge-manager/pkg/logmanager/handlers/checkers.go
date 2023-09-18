// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package handlers
package handlers

import (
	"math"

	"huawei.com/mindx/common/checker"

	"edge-manager/pkg/constants"
)

const (
	minNodes = 1
	maxNodes = 100

	genericStringRegexp = `^.{0,512}$`
)

func newCreateTaskReqChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		checker.GetUniqueListChecker(
			"EdgeNodes",
			checker.GetUintChecker("", 1, math.MaxInt64, true),
			minNodes, maxNodes, true),
		checker.GetStringChoiceChecker("Module", []string{"edgeNode"}, true),
	)
}

func newTaskErrorChecker() *checker.AndChecker {
	return checker.GetAndChecker(
		checker.GetRegChecker("Id", constants.SingleNodeTaskIdRegexpStr, true),
		checker.GetRegChecker("Reason", genericStringRegexp, false),
		checker.GetRegChecker("Message", genericStringRegexp, false),
	)
}
