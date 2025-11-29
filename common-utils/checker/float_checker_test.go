// Copyright (c) 2025. Huawei Technologies Co., Ltd. All rights reserved.

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
