// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

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
