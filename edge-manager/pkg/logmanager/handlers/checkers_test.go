// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
			{input: CreateTaskReq{Module: "edgeNode", EdgeNodes: []uint64{math.MaxInt64 + 1}}},
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
