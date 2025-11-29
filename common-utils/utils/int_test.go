//  Copyright(c) 2023. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package utils provides test of the util func about int
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	int1 = 1
	int2 = 2
	int3 = 3
)

func TestMaxInt(t *testing.T) {
	convey.Convey("test MaxInt", t, func() {
		res := MaxInt(int1, int2, int3)
		convey.So(res, convey.ShouldEqual, int3)
	})
}

func TestMinInt(t *testing.T) {
	convey.Convey("test MinInt", t, func() {
		res := MinInt(int1, int2, int3)
		convey.So(res, convey.ShouldEqual, int1)
	})
}
