//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package rand implement the security rand
package rand

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestRead(t *testing.T) {
	convey.Convey("package function test,normal situation", t, func() {
		//  the length of byte is one, to prevent block when generate random
		bs := make([]byte, 1, 1)
		l, err := Read(bs)
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(l, convey.ShouldEqual, 1)
	})
}
