// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

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
