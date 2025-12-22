// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils offer the some utils for certificate handling
package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestIsNil(t *testing.T) {
	var a interface{}               // type = nil, data = nil
	var b interface{} = (*int)(nil) // type is *int , data = nil
	var c interface{} = "dd"
	convey.Convey("test IsNil func, type and data is both nil", t, func() {
		convey.So(a == nil, convey.ShouldEqual, true)
		convey.So(b == nil, convey.ShouldEqual, false)
		convey.So(c == nil, convey.ShouldEqual, false)
		convey.So(IsNil(a), convey.ShouldEqual, true)
		convey.So(IsNil(b), convey.ShouldEqual, true)
		convey.So(IsNil(c), convey.ShouldEqual, false)
	})
}
