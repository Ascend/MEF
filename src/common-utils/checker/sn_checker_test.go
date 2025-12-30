// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package checker

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestSnChecker(t *testing.T) {
	checker := GetSnChecker("", true)
	convey.Convey("test normal case", t, func() {
		sn := "2102314NMV10P7100008"
		checkResult := checker.Check(sn)
		convey.So(checkResult.Result, convey.ShouldBeTrue)
	})
	convey.Convey("test null sn", t, func() {
		sn := ""
		checkResult := checker.Check(sn)
		convey.So(checkResult.Result, convey.ShouldBeFalse)
	})
	convey.Convey("test sn has invalid character", t, func() {
		sn := "!123"
		checkResult := checker.Check(sn)
		convey.So(checkResult.Result, convey.ShouldBeFalse)
	})
}
