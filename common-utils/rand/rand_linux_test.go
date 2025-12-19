// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package rand implement the security rand
package rand

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

const (
	illegalSize = 1 << 25
)

func TestInnerRead(t *testing.T) {
	convey.Convey("test random read func", t, func() {
		reader := &randomReader{}
		convey.Convey("read size too large, err returned", func() {
			bs := make([]byte, illegalSize, illegalSize)
			r, err := reader.Read(bs)
			convey.So(err.Error(), convey.ShouldEqual, "byte size is too large")
			convey.So(r, convey.ShouldEqual, 0)
		})
		convey.Convey("windows,err returned", func() {
			mock := gomonkey.ApplyGlobalVar(&supportOs, "windows")
			defer mock.Reset()
			bs := make([]byte, 1, 1)
			r, err := reader.Read(bs)
			convey.So(err.Error(), convey.ShouldEqual, "not supported")
			convey.So(r, convey.ShouldEqual, 0)
		})
		convey.Convey("open dev failed,err returned", func() {
			mock := gomonkey.ApplyFuncReturn(os.Open, nil, errors.New("mock error"))
			defer mock.Reset()
			bs := make([]byte, 1, 1)
			r, err := reader.Read(bs)
			convey.So(err.Error(), convey.ShouldEqual, "mock error")
			convey.So(r, convey.ShouldEqual, 0)
		})
		convey.Convey("normal situation,no err returned", func() {
			//  the length of byte is one, to prevent block when generate random
			bs := make([]byte, 1, 1)
			r, err := reader.Read(bs)
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(r, convey.ShouldEqual, 1)
		})
	})
}
