// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package checker for test float checker
package checker

import (
	"math"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

type testData struct {
	testField float64
}

func TestFloatChecker(t *testing.T) {
	const testField = "testField"
	convey.Convey("test float checker, test field does not exist and is not required", t, func() {
		checker := GetFloatChecker(testField, 0, 10, false)
		data := struct{}{}
		ret := checker.Check(data)
		convey.So(ret.Result, convey.ShouldBeTrue)
	})

	convey.Convey("test float checker, test field does not exist and is required", t, func() {
		checker := GetFloatChecker(testField, 0, 10, true)
		data := struct{}{}
		ret := checker.Check(data)
		convey.So(ret.Result, convey.ShouldBeFalse)
	})

	convey.Convey("test float checker, the value is within the valid range", t, func() {
		checker := GetFloatChecker(testField, 0, 10, true)
		dataWithValue := testData{testField: 5.0}
		ret := checker.Check(dataWithValue)
		convey.So(ret.Result, convey.ShouldBeTrue)
	})

	convey.Convey("test float checker, the value is not within the valid range", t, func() {
		checker := GetFloatChecker(testField, 0, 10, true)
		dataWithValue := testData{testField: 11.0}
		ret := checker.Check(dataWithValue)
		convey.So(ret.Result, convey.ShouldBeFalse)
	})

	convey.Convey("test float checker, test field is empty and the value is NaN", t, func() {
		checker := GetFloatChecker("", 0, 10, true)
		ret := checker.Check(math.NaN())
		convey.So(ret.Result, convey.ShouldBeFalse)
	})

	convey.Convey("test float checker, test field is not empty and the value is NaN", t, func() {
		checker := GetFloatChecker(testField, 0, 10, true)
		dataWithValue := testData{testField: math.NaN()}
		ret := checker.Check(dataWithValue)
		convey.So(ret.Result, convey.ShouldBeFalse)
	})
}
