// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for testing effect installer
package tasks

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
)

func TestEffectInstaller(t *testing.T) {
	p := gomonkey.ApplyMethodReturn(config.VersionXmlMgr{}, "GetVersion", "5.0.RC1", nil).
		ApplyFuncReturn(config.EffectToOldestVersionSmooth, nil).
		ApplyMethodReturn(UpgradeInstaller(testDir, constants.EffectMode), "Run", nil)
	defer p.Reset()

	convey.Convey("effect success", t, func() {
		err := EffectInstaller(testDir).Run()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("get version failed", t, func() {
		p1 := gomonkey.ApplyMethodReturn(config.VersionXmlMgr{}, "GetVersion", "", test.ErrTest)
		defer p1.Reset()
		err := EffectInstaller(testDir).Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("get upgrade inner version failed, error: %v", test.ErrTest))
	})

	convey.Convey("smooth config file failed", t, func() {
		p1 := gomonkey.ApplyFuncReturn(config.EffectToOldestVersionSmooth, test.ErrTest)
		defer p1.Reset()
		err := EffectInstaller(testDir).Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("smooth config file for old version failed,"+
			" error: %v", test.ErrTest))
	})

	convey.Convey("effect MEFEdge failed", t, func() {
		p1 := gomonkey.ApplyMethodReturn(UpgradeInstaller(testDir, constants.EffectMode), "Run", test.ErrTest)
		defer p1.Reset()
		err := EffectInstaller(testDir).Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("effect %s failed: %v", constants.MEFEdgeName, test.ErrTest))
	})
}
