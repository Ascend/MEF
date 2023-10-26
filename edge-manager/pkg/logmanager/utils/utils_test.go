// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/testutils"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"
)

func TestMain(m *testing.M) {
	if err := testutils.PrepareHwlog(); err != nil {
		fmt.Println("prepare hwlog failed")
	}
	if err := testutils.PrepareTempDirs(); err != nil {
		fmt.Printf("prepare dirs failed, %v\n", err)
	}
	m.Run()
	if err := testutils.CleanupTempDirs(); err != nil {
		fmt.Printf("cleanup dirs failed, %v\n", err)
	}
}

func TestCleanTempFiles(t *testing.T) {
	convey.Convey("test clean temp files", t, func() {
		err := fileutils.DeleteAllFileWithConfusion(constants.LogDumpTempDir)
		convey.So(err, convey.ShouldBeNil)
		err = fileutils.DeleteAllFileWithConfusion(constants.LogDumpPublicDir)
		convey.So(err, convey.ShouldBeNil)

		exists, err := CleanTempFiles()
		convey.So(err, convey.ShouldBeNil)
		convey.So(exists, convey.ShouldBeFalse)
	})
	convey.Convey("test clean temp files", t, func() {
		err := fileutils.CreateDir(constants.LogDumpTempDir, common.Mode700)
		convey.So(err, convey.ShouldBeNil)
		err = fileutils.CreateDir(constants.LogDumpPublicDir, common.Mode700)
		convey.So(err, convey.ShouldBeNil)

		exists, err := CleanTempFiles()
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(exists, convey.ShouldBeFalse)
	})
	convey.Convey("test clean temp files", t, func() {
		err := os.Chmod(constants.LogDumpTempDir, common.Mode755)
		convey.So(err, convey.ShouldBeNil)
		err = os.Chmod(constants.LogDumpPublicDir, common.Mode755)
		convey.So(err, convey.ShouldBeNil)

		_, err = CleanTempFiles()
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func TestFeedbackTaskError(t *testing.T) {
	convey.Convey("test feedback task error", t, func() {
		env := testutils.DummyTaskSchedule()
		gomonkey.ApplyMethodReturn(env.TaskCtx, "UpdateStatus", nil).
			ApplyMethodReturn(env.TaskCtx, "Spec", taskschedule.TaskSpec{})
		FeedbackTaskError(env.TaskCtx, errors.New("my error"))
	})
}
