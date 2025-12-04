// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
