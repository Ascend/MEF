// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package tasks for testing check online upgrade environment
package tasks

import (
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

var (
	testOnlineCheckDir = "/tmp/test_online_check_env"
	testDownloadPath   = filepath.Join(testOnlineCheckDir, "MEFEdgeDownload")
)

func setupOnlineCheck() error {
	if err := fileutils.CreateDir(testOnlineCheckDir, constants.Mode755); err != nil {
		return err
	}
	if err := fileutils.CreateDir(testDownloadPath, constants.Mode700); err != nil {
		return err
	}
	suffixes := []string{constants.TarGzExt, constants.CrlExt, constants.CmsExt}
	for _, suffix := range suffixes {
		testFile := filepath.Join(testDownloadPath, "test"+suffix)
		if err := fileutils.CreateFile(testFile, constants.Mode400); err != nil {
			return err
		}
	}
	return nil
}

func teardownOnlineCheck() {
	if err := fileutils.DeleteAllFileWithConfusion(testOnlineCheckDir); err != nil {
		hwlog.RunLog.Errorf("clear test online check env dir failed, error: %v", err)
	}
}

func TestCheckOnlineEdgeInstallerEnv(t *testing.T) {
	if err := setupOnlineCheck(); err != nil {
		panic(err)
	}
	defer teardownOnlineCheck()

	p := gomonkey.ApplyPrivateMethod(CheckEnvironmentBase{}, "checkDiskSpace",
		func(CheckEnvironmentBase) error { return nil }).
		ApplyPrivateMethod(CheckOfflineEdgeInstallerEnv{}, "checkPackageValid",
			func(CheckOfflineEdgeInstallerEnv) error { return nil }).
		ApplyPrivateMethod(CheckOfflineEdgeInstallerEnv{}, "unpackUgpTarPackage",
			func(CheckOfflineEdgeInstallerEnv) error { return nil })
	defer p.Reset()

	convey.Convey("check online upgrade environment success", t, checkOnlineEdgeInstallerEnvSuccess)
	convey.Convey("check online upgrade environment failed, change file owner failed", t, changeFileOwnerFailed)
}

func checkOnlineEdgeInstallerEnvSuccess() {
	err := NewPrepareOnlineInstallEnv(testDownloadPath, testOnlineCheckDir, testOnlineCheckDir).Run()
	convey.So(err, convey.ShouldBeNil)
}

func changeFileOwnerFailed() {
	p1 := gomonkey.ApplyFuncReturn(fileutils.SetPathOwnerGroup, test.ErrTest)
	defer p1.Reset()

	err := NewPrepareOnlineInstallEnv(testDownloadPath, testOnlineCheckDir, testOnlineCheckDir).Run()
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}
