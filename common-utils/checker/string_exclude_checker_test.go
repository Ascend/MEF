// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestStringExcludeChecker(t *testing.T) {
	checker := GetStringExcludeChecker("", []string{".."}, true)
	convey.Convey("test string exclude checker test failed", t, func() {
		var (
			testFileName = "/usr/local/../test"
		)
		ret := checker.Check(testFileName)
		convey.So(ret.Result, convey.ShouldEqual, false)
	})

	convey.Convey("test string exclude checker test ok", t, func() {
		var (
			testFileName = "./usr/local/test"
		)
		ret := checker.Check(testFileName)
		convey.So(ret.Result, convey.ShouldEqual, true)
	})
}
