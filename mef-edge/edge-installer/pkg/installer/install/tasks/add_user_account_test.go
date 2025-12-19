// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for testing add user account task
package tasks

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/test"
)

func TestAddUserAccountTask(t *testing.T) {
	addUserAccount := AddUserAccountTask{}
	convey.Convey("add user account should be success", t, func() {
		var p = gomonkey.ApplyMethodReturn(&envutils.UserMgr{}, "AddUserAccount", nil)
		defer p.Reset()
		err := addUserAccount.Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("add user account should be failed", t, func() {
		var p = gomonkey.ApplyMethodReturn(&envutils.UserMgr{}, "AddUserAccount", test.ErrTest)
		defer p.Reset()
		err := addUserAccount.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
