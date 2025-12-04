// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for testing check offline upgrade environment
package tasks

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/util"
	"edge-installer/pkg/common/veripkgutils"
	"huawei.com/mindx/mef/common/cmsverify"
)

var (
	testOfflineCheckDir = "/tmp/test_offline_check_env"
	testExtractPath     = filepath.Join(testOfflineCheckDir, "unpack")
	testInstallPath     = filepath.Join(testOfflineCheckDir, "MEFEdge")
	testTarPath         = filepath.Join(testOfflineCheckDir, "test.tar.gz")
	testCmsPath         = filepath.Join(testOfflineCheckDir, "test.tar.gz.cms")
	testCrlPath         = filepath.Join(testOfflineCheckDir, "test.tar.gz.crl")
	offlineChecker      = NewCheckOfflineEdgeInstallerEnv(testTarPath, testCmsPath, testCrlPath,
		testExtractPath, testInstallPath)
)

func TestCheckOfflineEdgeInstallerEnv(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, nil).
		ApplyFuncReturn(util.InSamePartition, true, nil).
		ApplyFuncReturn(envutils.CheckDiskSpace, nil).
		ApplyFuncReturn(util.CheckNecessaryCommands, nil).
		ApplyFuncReturn(veripkgutils.PrepareVerifyCrl, true, "", nil).
		ApplyFuncReturn(cmsverify.VerifyPackage, nil).
		ApplyFuncReturn(veripkgutils.UpdateLocalCrl, nil).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
		ApplyFuncReturn(fileutils.ExtraTarGzFile, nil)
	defer p.Reset()

	convey.Convey("check offline upgrade env success", t, checkOfflineEdgeInstallerEnvSuccess)
	convey.Convey("check offline upgrade env failed, clean env failed", t, cleanEnvFailed)
	convey.Convey("check offline upgrade env failed, check disk space failed", t, checkDiskSpaceFailed)
	convey.Convey("check offline upgrade env failed, disk space is not enough", t, diskSpaceNotEnough)
	convey.Convey("check offline upgrade env failed, check necessary commands failed", t, checkNecessaryCommandsFailed)
	convey.Convey("check offline upgrade env failed, check package valid failed", t, checkPackageValidFailed)
	convey.Convey("check offline upgrade env failed, unpack tar package failed", t, unpackUgpTarPackageFailed)
}

func checkOfflineEdgeInstallerEnvSuccess() {
	err := offlineChecker.Run()
	convey.So(err, convey.ShouldBeNil)
}

func cleanEnvFailed() {
	p1 := gomonkey.ApplyFuncSeq(fileutils.DeleteAllFileWithConfusion, []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},
	})
	defer p1.Reset()

	err := offlineChecker.Run()
	convey.So(err, convey.ShouldResemble, test.ErrTest)

	err = offlineChecker.Run()
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func checkDiskSpaceFailed() {
	convey.Convey("create dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p1.Reset()
		err := offlineChecker.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("check is same partition failed", func() {
		p1 := gomonkey.ApplyFuncReturn(util.InSamePartition, false, test.ErrTest)
		defer p1.Reset()
		err := offlineChecker.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func diskSpaceNotEnough() {
	p1 := gomonkey.ApplyFuncSeq(envutils.CheckDiskSpace, []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 2},
		{Values: gomonkey.Params{test.ErrTest}},
	})
	defer p1.Reset()
	const caseNum = 3
	for i := 0; i < caseNum; i++ {
		err := offlineChecker.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	}
}

func checkNecessaryCommandsFailed() {
	p1 := gomonkey.ApplyFuncReturn(util.CheckNecessaryCommands, test.ErrTest)
	defer p1.Reset()
	err := offlineChecker.Run()
	convey.So(err, convey.ShouldResemble, errors.New("check necessary commands failed"))
}

func checkPackageValidFailed() {
	convey.Convey("prepare crl for verifying package failed", func() {
		p1 := gomonkey.ApplyFuncReturn(veripkgutils.PrepareVerifyCrl, false, "", test.ErrTest)
		defer p1.Reset()
		err := offlineChecker.Run()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("verify package failed", func() {
		p1 := gomonkey.ApplyFuncReturn(cmsverify.VerifyPackage, test.ErrTest)
		defer p1.Reset()
		err := offlineChecker.Run()
		convey.So(err, convey.ShouldResemble, errors.New("verify package failed"))
	})

	convey.Convey("update crl file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(veripkgutils.UpdateLocalCrl, test.ErrTest)
		defer p1.Reset()
		err := offlineChecker.Run()
		convey.So(err, convey.ShouldResemble, errors.New("update crl file failed"))
	})
}

func unpackUgpTarPackageFailed() {
	convey.Convey("check extract path failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p1.Reset()
		err := offlineChecker.Run()
		convey.So(err, convey.ShouldResemble, errors.New("check extractPath failed"))
	})

	convey.Convey("extract tar package file failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.ExtraTarGzFile, test.ErrTest)
		defer p1.Reset()
		err := offlineChecker.Run()
		convey.So(err, convey.ShouldResemble, errors.New("extract tar package file failed"))
	})
}
