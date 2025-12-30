// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
