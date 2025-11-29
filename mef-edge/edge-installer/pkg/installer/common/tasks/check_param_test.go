// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing check param task
package tasks

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/installer/common"
)

var (
	testDir           = "/tmp/test_check_param"
	checkInstallParam = CheckParamTask{InstallRootDir: testDir, InstallationPkgDir: testDir}
)

func TestCheckParamTask(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(common.CheckDir, nil)
	defer p.Reset()
	convey.Convey("check Param should be success", t, checkParamSuccess)
	convey.Convey("check Param should be failed, check origin dir failed", t, checkOriginInstallPkgDirFailed)
	convey.Convey("check Param should be failed, check install root dir failed", t, checkInstallRootDirFailed)
}

func checkParamSuccess() {
	p1 := gomonkey.ApplyFuncReturn(common.CheckInTmpfs, nil)
	defer p1.Reset()

	err := checkInstallParam.Run()
	convey.So(err, convey.ShouldBeNil)
}

func checkOriginInstallPkgDirFailed() {
	p1 := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
	defer p1.Reset()

	err := checkInstallParam.Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("check install package dir [%s] failed, error: %v",
		testDir, test.ErrTest))
}

func checkInstallRootDirFailed() {
	convey.Convey("check dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(common.CheckDir, test.ErrTest)
		defer p1.Reset()
		err := checkInstallParam.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("check in tmpfs failed", func() {
		p1 := gomonkey.ApplyFuncReturn(common.CheckInTmpfs, test.ErrTest)
		defer p1.Reset()
		err := checkInstallParam.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}
