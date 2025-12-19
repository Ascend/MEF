// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package nodemanager for node init test
package nodemanager

import (
	"net/http"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/modulemgr/model"
)

func TestDispatchMsg(t *testing.T) {
	convey.Convey("selectMethod functional test", t, func() {
		convey.Convey("innerGetNodeInfoByUniqueName success", func() {
			input := model.Message{}
			_, err := selectMethod(&input)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("innerGetNodeInfoByUniqueName param error", func() {
			input := model.Message{}
			input.SetRouter("", "", http.MethodGet, nodeUrlRootPath)
			_, err := selectMethod(&input)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
