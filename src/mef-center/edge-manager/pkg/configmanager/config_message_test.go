// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package configmanager for inner message test
package configmanager

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetAllNodeInfo(t *testing.T) {
	convey.Convey("get all node info functional test", t, func() {
		convey.Convey("get all node info success", func() {
			_, err := getAllNodeInfo()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
