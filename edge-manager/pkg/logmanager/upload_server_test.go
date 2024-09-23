// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logmanager
package logmanager

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/testutils"
	"edge-manager/pkg/logmanager/utils"
)

func TestProcessUpload(t *testing.T) {
	convey.Convey("test process upload", t, func() {
		env := testutils.DummyTaskSchedule()

		type responseWriter struct {
			http.ResponseWriter
		}

		p := uploadProcess{responseWriter: &responseWriter{}}

		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, env.Scheduler).
			ApplyMethodReturn(env.TaskCtx, "UpdateStatus", nil).
			ApplyPrivateMethod(p, "receiveFile", func(*uploadProcess, taskschedule.TaskContext) error { return nil }).
			ApplyPrivateMethod(p, "verifyFile", func(*uploadProcess, taskschedule.TaskContext) error { return nil }).
			ApplyMethodReturn(p.responseWriter, "Write", 0, nil).
			ApplyFunc(utils.FeedbackTaskError, func(ctx taskschedule.TaskContext, err error) {})
		defer patch.Reset()

		err := p.processUpload()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestReceiveFile(t *testing.T) {
	convey.Convey("test receive file", t, func() {
		env := testutils.DummyTaskSchedule()
		err := os.MkdirAll(constants.LogDumpTempDir, common.Mode600)
		convey.So(err, convey.ShouldBeNil)

		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, env.Scheduler).
			ApplyMethodReturn(env.TaskCtx, "UpdateStatus", nil).
			ApplyMethodReturn(env.TaskCtx, "GracefulShutdown", nil).
			ApplyMethodReturn(env.TaskCtx, "UpdateLiveness", nil).
			ApplyFunc(utils.WithDiskPressureProtect, testutils.WithoutDiskPressureProtect)
		defer patch.Reset()
		p := uploadProcess{httpRequest: &http.Request{Body: io.NopCloser(strings.NewReader("hello"))}}
		err = p.receiveFile(env.TaskCtx)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCreateUploadProcess(t *testing.T) {
	convey.Convey("test create upload process", t, func() {
		testcases := []struct {
			header http.Header
			result bool
		}{
			{header: map[string][]string{
				headerTaskId:      {constants.DumpSingleNodeLogTaskName + ".test.1"},
				headerPackageSize: {strconv.Itoa(constants.LogUploadMaxSize)},
				"X-Forwarded-For": {"test"},
			}, result: true},
			{header: map[string][]string{
				headerTaskId:      {constants.DumpSingleNodeLogTaskName + ".test.1"},
				headerPackageSize: {strconv.Itoa(constants.LogUploadMaxSize + 1)},
				"X-Forwarded-For": {"test"},
			}},
			{header: map[string][]string{
				headerTaskId:      {constants.DumpSingleNodeLogTaskName + "/.test.1"},
				headerPackageSize: {strconv.Itoa(constants.LogUploadMaxSize)},
				"X-Forwarded-For": {"test"},
			}},
		}

		taskScheduler := testutils.DummyTaskSchedule().Scheduler
		p := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, taskScheduler)
		defer p.Reset()

		for _, testcase := range testcases {
			assertion := convey.ShouldNotBeNil
			if testcase.result {
				assertion = convey.ShouldBeNil
			}
			_, err := createUploadProcess(nil, &http.Request{Header: testcase.header})
			convey.So(err, assertion)
		}
	})
}
