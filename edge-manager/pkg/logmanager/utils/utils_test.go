// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package utils
package utils

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/testutils"
)

const mode777 = 0777

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
		convey.So(err, convey.ShouldBeNil)
		convey.So(exists, convey.ShouldBeTrue)
	})
	convey.Convey("test clean temp files", t, func() {
		err := os.Chmod(constants.LogDumpTempDir, common.Mode755)
		convey.So(err, convey.ShouldBeNil)
		err = os.Chmod(constants.LogDumpPublicDir, mode777)
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
