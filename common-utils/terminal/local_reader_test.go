//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package terminal provide a safe reader for password
package terminal

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestLocalReaderRead(t *testing.T) {
	convey.Convey("test localReader", t, func() {
		buf := make([]byte, 0)
		r, err := localReader(0).Read(buf)
		convey.So(r, convey.ShouldEqual, 0)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("test localReader, buf length 1", t, func() {
		buf := make([]byte, 1)
		fmt.Println(len(buf))
		r, err := localReader(0).Read(buf)
		convey.So(r, convey.ShouldEqual, 0)
		convey.So(err, convey.ShouldEqual, nil)
	})
}
