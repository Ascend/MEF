// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
			checker.GetUintChecker("", 1, math.MaxUint32, true),
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
