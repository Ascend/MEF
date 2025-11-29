// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

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
