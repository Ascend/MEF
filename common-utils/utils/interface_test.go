//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

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
