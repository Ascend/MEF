// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils this file for password handler
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	truePasswd   = []byte("aA0!\"#$%&'()*+,-. /:;<=>?@[\\]^_`{|}~")
	falsePasswd1 = []byte("userName")
	falsePasswd2 = []byte("12345678")
	falsePasswd3 = []byte("1234567")
	falsePasswd4 = []byte("emaNresu.")
	falsePasswd5 = []byte("不支持特殊字符测试test")
)

// TestCommonCheckForPassWord test common check for passWord
func TestCommonCheckForPassWord(t *testing.T) {
	convey.Convey("correct password", t, func() {
		err := ValidatePassWord("userName", truePasswd)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("username == password", t, func() {
		err := ValidatePassWord("userName", falsePasswd1)
		convey.So(err.Error(), convey.ShouldEqual, "password cannot equals username")
	})
	convey.Convey("complex not meet the requirement", t, func() {
		err := ValidatePassWord("userName", falsePasswd2)
		convey.So(err.Error(), convey.ShouldEqual, "password complex not meet the requirement")
	})
	convey.Convey("password too short", t, func() {
		err := ValidatePassWord("userName", falsePasswd3)
		convey.So(err.Error(), convey.ShouldEqual, "password not meet requirement")
	})
	convey.Convey("username equal reverse password", t, func() {
		err := ValidatePassWord(".userName", falsePasswd4)
		convey.So(err.Error(), convey.ShouldEqual, "password cannot equal reversed username")
	})
	convey.Convey("test special ", t, func() {
		err := ValidatePassWord("userName", falsePasswd5)
		convey.So(err.Error(), convey.ShouldEqual, "password not meet requirement")
	})
}

// TestClearStringMemory test ClearStringMemory
func TestClearStringMemory(t *testing.T) {
	convey.Convey("test clear string password", t, func() {
		testCleanStr := []byte{97, 98, 99}
		s := string(testCleanStr)
		ClearStringMemory(s)
		convey.So(s, convey.ShouldNotEqual, "abc")
	})
}
