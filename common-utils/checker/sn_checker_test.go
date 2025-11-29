// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

package checker

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSnChecker(t *testing.T) {
	checker := GetSnChecker("", true)
	convey.Convey("test normal case", t, func() {
		sn := "2102314NMV10P7100008"
		checkResult := checker.Check(sn)
		convey.So(checkResult.Result, convey.ShouldBeTrue)
	})
	convey.Convey("test null sn", t, func() {
		sn := ""
		checkResult := checker.Check(sn)
		convey.So(checkResult.Result, convey.ShouldBeFalse)
	})
	convey.Convey("test sn has invalid character", t, func() {
		sn := "!123"
		checkResult := checker.Check(sn)
		convey.So(checkResult.Result, convey.ShouldBeFalse)
	})
}
