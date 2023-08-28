// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
