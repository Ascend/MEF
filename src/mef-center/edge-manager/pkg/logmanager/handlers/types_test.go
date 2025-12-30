// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

// TestNullableTime tests NullableTime
func TestNullableTime(t *testing.T) {
	convey.Convey("test nullable time", t, func() {
		now := time.Now()
		nowStr, err := json.Marshal(now)
		convey.So(err, convey.ShouldBeNil)
		testcases := []struct {
			input    NullableTime
			expected string
		}{
			{input: NullableTime(time.Time{}), expected: "null"},
			{input: NullableTime(now), expected: string(nowStr)},
		}

		for _, testcase := range testcases {
			output, err := json.Marshal(testcase.input)
			convey.So(err, convey.ShouldBeNil)
			convey.So(string(output), convey.ShouldEqual, testcase.expected)
		}
	})
}
