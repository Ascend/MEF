// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package commands
package commands

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/kmcupdate"
)

func TestUpdateKmcCmd(t *testing.T) {
	convey.Convey("test update kmc cmd methods", t, updateKmcCmdMethods)
	convey.Convey("test update kmc cmd successful", t, updateKmcCmdSuccess)
	convey.Convey("test update kmc cmd failed", t, executeUpdateKmcFailed)
}

func updateKmcCmdMethods() {
	convey.So(UpdateKmcCmd().Name(), convey.ShouldEqual, common.UpdateKmc)
	convey.So(UpdateKmcCmd().Description(), convey.ShouldEqual, common.UpdateKmcDesc)
	convey.So(UpdateKmcCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(UpdateKmcCmd().LockFlag(), convey.ShouldBeFalse)
}

func updateKmcCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(&kmcupdate.UpdateKmcFlow{}, "RunFlow", nil)
	defer p.Reset()
	err := UpdateKmcCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	UpdateKmcCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeUpdateKmcFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := UpdateKmcCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("run update kmc flow failed", func() {
		p := gomonkey.ApplyMethodReturn(&kmcupdate.UpdateKmcFlow{}, "RunFlow", test.ErrTest)
		defer p.Reset()
		err := UpdateKmcCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	UpdateKmcCmd().PrintOpLogFail(userRoot, ipLocalhost)
}
