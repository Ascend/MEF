// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package checker this file for base path check method
package checker

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestIsPathValid(t *testing.T) {
	convey.Convey("TestIsPathValid", t, func() {
		convey.ShouldBeTrue(IsPathValid("/test"))
	})
}
