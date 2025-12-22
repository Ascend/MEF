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
	"encoding/json"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"
)

func TestMergePatch(t *testing.T) {
	dateA := []byte(`{"name":"TestMergePatch","date":28}`)
	dateB := []byte(`{"name":"TestMergePatch","date":27}`)

	convey.Convey("TestMergePatch", t, func() {
		_, err := MergePatch(dateA, dateB)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func MergePatch failed, patch type is not allowed", t, func() {
		patchByte, err := json.Marshal("patch string")
		convey.So(err, convey.ShouldBeNil)

		_, err = MergePatch(dateA, patchByte)
		convey.So(err, convey.ShouldResemble, errBadPatchType)
	})

	convey.Convey("test MergePatch failed, unmarshal failed", t, func() {
		var p1 = gomonkey.ApplyFuncSeq(json.Unmarshal, []gomonkey.OutputCell{
			{Values: gomonkey.Params{test.ErrTest}},

			{Values: gomonkey.Params{nil}},
			{Values: gomonkey.Params{test.ErrTest}},
		})
		defer p1.Reset()
		_, err := MergePatch(dateA, dateB)
		convey.So(err, convey.ShouldResemble, test.ErrTest)

		_, err = MergePatch(dateA, dateB)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("test MergePatch failed, marshal failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(json.Marshal, nil, test.ErrTest)
		defer p1.Reset()
		_, err := MergePatch(dateA, dateB)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
