// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package flows for testing upgrade base
package flows

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

var base = upgradeBase{
	edgeDir:     testDir,
	extractPath: testDir,
}

func TestUpgradeBase(t *testing.T) {
	convey.Convey("test unlock upgrade flag", t, testUnlockUpgradeFlag)
	convey.Convey("test clear unpack path", t, testClearUnpackPath)
}

func testUnlockUpgradeFlag() {
	mockLockInstance := util.FlagLockInstance(constants.FlagPath, constants.ProcessFlag, constants.Upgrade)
	p := gomonkey.ApplyFuncReturn(util.FlagLockInstance, mockLockInstance)
	defer p.Reset()

	convey.Convey("unlock upgrade success", func() {
		p1 := gomonkey.ApplyMethodReturn(mockLockInstance, "Unlock", nil)
		defer p1.Reset()
		base.unlockUpgradeFlag()
	})

	convey.Convey("unlock upgrade failed", func() {
		p1 := gomonkey.ApplyMethodReturn(mockLockInstance, "Unlock", test.ErrTest)
		defer p1.Reset()
		base.unlockUpgradeFlag()
	})
}

func testClearUnpackPath() {
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true)
	defer p.Reset()

	convey.Convey("clear unpack path success", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, nil)
		defer p1.Reset()
		base.clearUnpackPath()
	})

	convey.Convey("clear unpack path failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p1.Reset()
		base.clearUnpackPath()
	})

	convey.Convey("unpack package path does not exist", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
		defer p1.Reset()
		base.clearUnpackPath()
	})
}
