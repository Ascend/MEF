// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestStringLengthChecker(t *testing.T) {
	const (
		minLength = 2
		maxLength = 5
	)
	checker := GetStringLengthChecker("", minLength, maxLength, true)
	convey.Convey("test normal case", t, func() {
		str := "aaa"
		checkResult := checker.Check(str)
		convey.So(checkResult.Result, convey.ShouldBeTrue)
	})
	convey.Convey("test short str", t, func() {
		str := "a"
		checkResult := checker.Check(str)
		convey.So(checkResult.Result, convey.ShouldBeFalse)
	})
	convey.Convey("test long str", t, func() {
		str := "aaaaaa"
		checkResult := checker.Check(str)
		convey.So(checkResult.Result, convey.ShouldBeFalse)
	})
}
