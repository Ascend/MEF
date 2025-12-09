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
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"edge-manager/pkg/constants"
)

// TestCreateTaskChecker tests createTaskHandler's arguments checker
func TestCreateTaskChecker(t *testing.T) {
	convey.Convey("test createTaskChecker", t, func() {
		var longEdgeNodeIdSlice []uint64
		for i := 0; i <= maxNodes; i++ {
			longEdgeNodeIdSlice = append(longEdgeNodeIdSlice, uint64(i+1))
		}
		testcases := []struct {
			input  CreateTaskReq
			result bool
		}{
			{input: CreateTaskReq{Module: "", EdgeNodes: []uint64{1}}},
			{input: CreateTaskReq{Module: "badModule", EdgeNodes: []uint64{1}}},
			{input: CreateTaskReq{Module: "edgeNode", EdgeNodes: []uint64{0}}},
			{input: CreateTaskReq{Module: "edgeNode", EdgeNodes: []uint64{math.MaxUint32 + 1}}},
			{input: CreateTaskReq{Module: "edgeNode", EdgeNodes: []uint64{1, 1}}},
			{input: CreateTaskReq{Module: "edgeNode"}},
			{input: CreateTaskReq{Module: "edgeNode", EdgeNodes: longEdgeNodeIdSlice}},
			{input: CreateTaskReq{Module: "edgeNode", EdgeNodes: []uint64{1, 2}}, result: true},
		}

		for _, testcase := range testcases {
			checkResult := newCreateTaskReqChecker().Check(testcase.input)
			assertion := convey.ShouldBeFalse
			if testcase.result {
				assertion = convey.ShouldBeTrue
			}
			convey.So(checkResult.Result, assertion)
		}
	})
}

// TestReportErrorChecker tests reportErrorHandler's arguments checker
func TestReportErrorChecker(t *testing.T) {
	convey.Convey("test reportErrorChecker", t, func() {
		const longStrLen = 513
		longStrBuffer := make([]byte, longStrLen)
		for i := 0; i < len(longStrBuffer); i++ {
			longStrBuffer[i] = 'a'
		}
		longStr := string(longStrBuffer)
		testcases := []struct {
			input  TaskErrorInfo
			result bool
		}{
			{input: TaskErrorInfo{}},
			{input: TaskErrorInfo{Id: constants.DumpSingleNodeLogTaskName + "/"}},
			{input: TaskErrorInfo{Id: constants.DumpSingleNodeLogTaskName + "\\"}},
			{input: TaskErrorInfo{Id: constants.DumpSingleNodeLogTaskName + ".Aa1", Message: longStr}},
			{input: TaskErrorInfo{Id: constants.DumpSingleNodeLogTaskName +
				".13808ef0-b624-4f62-8adc-6e1430ba3cd2", Message: longStr}},
			{input: TaskErrorInfo{Id: constants.DumpSingleNodeLogTaskName + ".Aa1"}, result: true},
			{input: TaskErrorInfo{Id: constants.DumpSingleNodeLogTaskName +
				".13808ef0-b624-4f62-8adc-6e1430ba3cd2"}, result: true},
		}

		for _, testcase := range testcases {
			checkResult := newTaskErrorChecker().Check(testcase.input)
			assertion := convey.ShouldBeFalse
			if testcase.result {
				assertion = convey.ShouldBeTrue
			}
			convey.So(checkResult.Result, assertion)
		}
	})
}
