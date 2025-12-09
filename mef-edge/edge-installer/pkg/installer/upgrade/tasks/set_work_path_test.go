// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tasks for testing set work path task
package tasks

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
)

func TestSetWorkPathTask(t *testing.T) {
	convey.Convey("set work path success", t, setWorkPathSuccess)
	convey.Convey("set work path failed, prepare install dir failed", t, prepareInstallDirFailed)
}

func setWorkPathSuccess() {
	defer clearEnv(testDir)
	err := setWorkPathTask.Run()
	convey.So(err, convey.ShouldBeNil)
}

func prepareInstallDirFailed() {
	convey.Convey("clean target software install dir failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p.Reset()

		err := setWorkPathTask.Run()
		expectErr := fmt.Errorf("clean target software install dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("create target software install dir failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p.Reset()

		err := setWorkPathTask.Run()
		expectErr := fmt.Errorf("create target software install dir failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
