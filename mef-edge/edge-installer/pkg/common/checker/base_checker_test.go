// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker this file for base check method
package checker

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	value = 3
	min   = 1
	max   = 5
)

func TestRegexStringChecker(t *testing.T) {
	convey.Convey("TestRegexStringChecker", t, func() {
		convey.So(RegexStringChecker("3", "^[0-9]$"), convey.ShouldBeTrue)
	})
	convey.Convey("TestIntChecker", t, func() {
		convey.So(IntChecker(value, min, max), convey.ShouldBeTrue)
	})
}
