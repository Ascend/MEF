//  Copyright(C) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package terminal provide a safe reader for password
package terminal

import (
	"syscall"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestErrnoConvert(t *testing.T) {
	convey.Convey("test errnoConvert EAGAIN, ", t, func() {
		err := errnoConvert(syscall.EAGAIN)
		convey.So(err, convey.ShouldEqual, errorEAGAINE)
	})

	convey.Convey("test errnoConvert EINVAL, ", t, func() {
		err := errnoConvert(syscall.EINVAL)
		convey.So(err, convey.ShouldEqual, errorEINVAL)
	})

	convey.Convey("test errnoConvert ENOENT, ", t, func() {
		err := errnoConvert(syscall.ENOENT)
		convey.So(err, convey.ShouldEqual, errorENOENT)
	})
	convey.Convey("test errnoConvert 0, ", t, func() {
		err := errnoConvert(0)
		convey.So(err, convey.ShouldEqual, nil)
	})
}
