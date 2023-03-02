// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package common

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestStringExcludeChecker(t *testing.T) {
	convey.Convey("test string exclude checker test failed", t, func() {
		var (
			testFileName = "/usr/local/../test"
		)
		ret := FileNameCheck(testFileName)
		convey.So(ret.Result, convey.ShouldEqual, false)
	})

	convey.Convey("test string exclude checker test ok", t, func() {
		var (
			testFileName = "/usr/local/test"
		)
		ret := FileNameCheck(testFileName)
		convey.So(ret.Result, convey.ShouldEqual, true)
	})

	convey.Convey("test string exclude checker test failed", t, func() {
		var (
			testFileName = "/usr/local$$/test"
		)
		ret := FileNameCheck(testFileName)
		convey.So(ret.Result, convey.ShouldEqual, false)
	})
}
