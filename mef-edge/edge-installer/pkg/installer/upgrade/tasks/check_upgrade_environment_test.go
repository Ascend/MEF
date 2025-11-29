// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package tasks for testing check upgrade environment task
package tasks

import (
	"errors"
	"fmt"
	"os/exec"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/util"
)

var checkUpgradeEnvironment = CheckUpgradeEnvironmentTask{SoftwarePathMgr: pathMgr.SoftwarePathMgr}

func TestCheckUpgradeEnvironmentTask(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(fileutils.EvalSymlinks, "", nil).
		ApplyMethodReturn(config.VersionXmlMgr{}, "GetInnerVersion", "1.0", nil).
		ApplyMethodReturn(config.VersionXmlMgr{}, "GetSftPkgName", "MEFEdge", nil).
		ApplyFuncReturn(exec.LookPath, "", nil)
	defer p.Reset()

	convey.Convey("check environment success", t, checkEnvSuccess)
	convey.Convey("check environment failed, check version failed", t, checkVersionFailed)
	convey.Convey("check environment failed, check sft pkg consistency failed", t, checkSftPkgConsistencyFailed)
	convey.Convey("check environment failed, check disk space failed", t, checkDiskSpaceFailed)
}

func checkEnvSuccess() {
	var p1 = gomonkey.ApplyFuncReturn(envutils.CheckDiskSpace, nil)
	defer p1.Reset()
	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldBeNil)
}

func checkVersionFailed() {
	convey.Convey("check environment failed, get real file path failed", getRealFilePathFailed)
	convey.Convey("check environment failed, get old inner version failed", getOldInnerVersionFailed)
	convey.Convey("check environment failed, get new inner version failed", getNewInnerVersionFailed)
	convey.Convey("check environment failed, compare version failed", compareVersionFailed)
	convey.Convey("check environment failed, upgrade version is invalid", upgradeVersionIsInvalid)
}

func checkSftPkgConsistencyFailed() {
	convey.Convey("check environment failed, get old package name failed", getOldPackageNameFailed)
	convey.Convey("check environment failed, get new package name failed", getNewPackageNameFailed)
	convey.Convey("check environment failed, package names are inconsistent", packageNameInconsistent)
}

func getRealFilePathFailed() {
	p1 := gomonkey.ApplyFuncSeq(fileutils.EvalSymlinks, []gomonkey.OutputCell{
		{Values: gomonkey.Params{"", test.ErrTest}},

		{Values: gomonkey.Params{"", nil}},
		{Values: gomonkey.Params{"", test.ErrTest}},
	})
	defer p1.Reset()

	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("get real file path failed"))

	err = checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("get real file path failed"))
}

func getOldInnerVersionFailed() {
	p1 := gomonkey.ApplyMethodReturn(config.VersionXmlMgr{}, "GetInnerVersion", "", test.ErrTest)
	defer p1.Reset()
	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("get old inner version failed"))
}

func getNewInnerVersionFailed() {
	p1 := gomonkey.ApplyMethodSeq(config.VersionXmlMgr{}, "GetInnerVersion", []gomonkey.OutputCell{
		{Values: gomonkey.Params{"1.0", nil}},
		{Values: gomonkey.Params{"1.0", test.ErrTest}},
	})
	defer p1.Reset()
	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("get new inner version failed"))
}

func compareVersionFailed() {
	p1 := gomonkey.ApplyMethodSeq(config.VersionXmlMgr{}, "GetInnerVersion", []gomonkey.OutputCell{
		{Values: gomonkey.Params{"1.0", nil}, Times: 2},
	}).
		ApplyFuncReturn(util.IsValidVersion, false, test.ErrTest)
	defer p1.Reset()
	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func upgradeVersionIsInvalid() {
	oldVersion, newVersion := "1.0", "3.0"
	p1 := gomonkey.ApplyMethodSeq(config.VersionXmlMgr{}, "GetInnerVersion", []gomonkey.OutputCell{
		{Values: gomonkey.Params{oldVersion, nil}},
		{Values: gomonkey.Params{newVersion, nil}},
	})
	defer p1.Reset()
	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("upgrade version [%s] is not the previous version or the "+
		"next version of current version [%s]", newVersion, oldVersion))
}

func getOldPackageNameFailed() {
	p1 := gomonkey.ApplyMethodReturn(config.VersionXmlMgr{}, "GetSftPkgName", "", test.ErrTest)
	defer p1.Reset()
	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("get old package name failed"))
}

func getNewPackageNameFailed() {
	p1 := gomonkey.ApplyMethodSeq(config.VersionXmlMgr{}, "GetSftPkgName", []gomonkey.OutputCell{
		{Values: gomonkey.Params{"MEFEdge", nil}},
		{Values: gomonkey.Params{"MEFEdge", test.ErrTest}},
	})
	defer p1.Reset()
	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("get new inner package name failed"))
}

func packageNameInconsistent() {
	p1 := gomonkey.ApplyMethodSeq(config.VersionXmlMgr{}, "GetSftPkgName", []gomonkey.OutputCell{
		{Values: gomonkey.Params{"MEFEdge", nil}},
		{Values: gomonkey.Params{"MEFEdgeSDK", nil}},
	})
	defer p1.Reset()
	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("package names are inconsistent"))
}

func checkDiskSpaceFailed() {
	var p1 = gomonkey.ApplyFuncReturn(envutils.CheckDiskSpace, test.ErrTest)
	defer p1.Reset()

	err := checkUpgradeEnvironment.Run()
	convey.So(err, convey.ShouldResemble, errors.New("check disk space failed"))
}
