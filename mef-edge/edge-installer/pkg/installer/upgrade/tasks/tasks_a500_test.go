// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_A500

// Package tasks for testing methods that are performed only on the a500 device
package tasks

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
)

func TestPrepareCfgBackupDir(t *testing.T) {
	defer clearEnv(testDir)
	convey.Convey("prepare config backup dir success", t, func() {
		err := setWorkPathTask.prepareCfgBackupDir()
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("clean config backup dir failed", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p.Reset()

		err := setWorkPathTask.prepareCfgBackupDir()
		expectErr := fmt.Errorf("clean config backup dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("create config backup dir dir failed", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p.Reset()

		err := setWorkPathTask.prepareCfgBackupDir()
		expectErr := fmt.Errorf("create config backup dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func TestRefreshDefaultCfgDir(t *testing.T) {
	convey.Convey("remove old config backup dir failed", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p.Reset()

		err := postEffectProcess.refreshDefaultCfgDir()
		expectErr := fmt.Errorf("remove old config backup dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("rename config backup temp directory failed", t, func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RenameFile, test.ErrTest)
		defer p.Reset()

		err := postEffectProcess.refreshDefaultCfgDir()
		expectErr := fmt.Errorf("rename config backup temp directory to [%s] failed, error: %v",
			postEffectProcess.ConfigPathMgr.GetConfigBackupDir(), test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
