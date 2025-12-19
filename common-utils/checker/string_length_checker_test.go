// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
