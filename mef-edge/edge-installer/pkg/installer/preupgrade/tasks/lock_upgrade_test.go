// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing lock upgrade
package tasks

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

func TestLockUpgrade(t *testing.T) {
	convey.Convey("lock upgrade success", t, func() {
		p := gomonkey.ApplyMethodReturn(util.FlagLockInstance(constants.FlagPath,
			constants.ProcessFlag, constants.Upgrade), "Lock", nil)
		defer p.Reset()
		err := LockUpgrade().Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("lock upgrade failed", t, func() {
		p := gomonkey.ApplyMethodReturn(util.FlagLockInstance(constants.FlagPath,
			constants.ProcessFlag, constants.Upgrade), "Lock", test.ErrTest)
		defer p.Reset()
		err := LockUpgrade().Run()
		convey.So(err, convey.ShouldResemble, errors.New("lock upgrade failed,there may be"+
			" another process processing the upgrade"))
	})
}
