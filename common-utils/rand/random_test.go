// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
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
	"testing"

	"github.com/smartystreets/goconvey/convey"
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
