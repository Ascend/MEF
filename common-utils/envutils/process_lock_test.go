// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package envutils
package envutils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestFlock(t *testing.T) {
	convey.Convey("test flock", t, func() {
		const lockName = "my_lock"
		err := GetFlock(lockName).Lock("reason one")
		convey.So(err, convey.ShouldBeNil)
		err = GetFlock(lockName).Lock("reason two")
		convey.So(err, convey.ShouldNotBeNil)
		GetFlock(lockName).Unlock()
	})
}
