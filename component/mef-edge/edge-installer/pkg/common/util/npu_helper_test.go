// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package util

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestIsWholeNpu(t *testing.T) {
	convey.Convey("test func IsWholeNpu", t, func() {
		convey.So(IsWholeNpu("huawei.com/Ascend123"), convey.ShouldBeTrue)
		convey.So(IsWholeNpu("err"), convey.ShouldBeFalse)
	})
}

func TestFindMostQualifiedNpu(t *testing.T) {
	convey.Convey("Given a nil resObj", t, func() {
		result, ok := FindMostQualifiedNpu(nil)
		convey.So(result, convey.ShouldEqual, "")
		convey.So(ok, convey.ShouldBeFalse)
	})

	convey.Convey("Given an empty resObj", t, func() {
		resObj := make(map[string]interface{})
		result, ok := FindMostQualifiedNpu(resObj)
		convey.So(result, convey.ShouldEqual, "")
		convey.So(ok, convey.ShouldBeFalse)
	})

	convey.Convey("Given a resObj with multiple keys", t, func() {
		resObj := map[string]interface{}{
			"huawei.com/davinci-mini": "1111",
			"key2":                    "huawei.com/davinci-mini",
		}
		result, ok := FindMostQualifiedNpu(resObj)
		convey.So(result, convey.ShouldEqual, "huawei.com/davinci-mini")
		convey.So(ok, convey.ShouldBeTrue)
	})

	convey.Convey("test func FindMostQualifiedNpu failed, object type error", t, func() {
		result, ok := FindMostQualifiedNpu("error obj")
		convey.So(result, convey.ShouldEqual, "")
		convey.So(ok, convey.ShouldBeFalse)
	})
}
