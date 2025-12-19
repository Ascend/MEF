// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
