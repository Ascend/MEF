// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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

type structWithPointer struct {
	val *string
}

func TestUniqueListChecker(t *testing.T) {
	const maxLen = 100
	checker := GetUniqueListChecker("", GetAndChecker(), 1, maxLen, true)

	convey.Convey("list with duplicate elements", t, func() {
		var (
			str1 = "a"
			str2 = "a"
		)
		val1 := structWithPointer{val: &str1}
		val2 := structWithPointer{val: &str2}
		convey.So(val1.val, convey.ShouldNotEqual, val2.val)
		convey.So(*val1.val, convey.ShouldEqual, *val2.val)
		list := []interface{}{val1, val2}
		result := checker.Check(list)
		convey.So(result.Result, convey.ShouldBeFalse)
		convey.So(result.Reason, convey.ShouldContainSubstring, "unique")
	})

	convey.Convey("list without duplicate elements", t, func() {
		var (
			str1 = "a"
			str2 = "b"
		)
		val1 := structWithPointer{val: &str1}
		val2 := structWithPointer{val: &str2}
		list := []interface{}{val1, val2}
		result := checker.Check(list)
		convey.So(result.Result, convey.ShouldBeTrue)
	})

	convey.Convey("list contains null pointer", t, func() {
		var ptr1, ptr2 *structWithPointer
		list := []interface{}{ptr1, ptr2}
		result := checker.Check(list)
		convey.So(result.Result, convey.ShouldBeFalse)
		convey.So(result.Reason, convey.ShouldContainSubstring, "unique")
	})

	convey.Convey("element contains null pointer", t, func() {
		var val1, val2 structWithPointer
		list := []interface{}{val1, val2}
		result := checker.Check(list)
		convey.So(result.Result, convey.ShouldBeFalse)
		convey.So(result.Reason, convey.ShouldContainSubstring, "unique")
	})
}
