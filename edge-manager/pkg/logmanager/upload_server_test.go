// Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

// Package logmanager
package logmanager

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	utils2 "huawei.com/mindx/common/utils"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/taskschedule"

	"edge-manager/pkg/constants"
	"edge-manager/pkg/logmanager/testutils"
	"edge-manager/pkg/logmanager/utils"
)

func TestMain(m *testing.M) {
	if err := testutils.PrepareHwlog(); err != nil {
		fmt.Printf("init hwlog failed, %v\n", err)
	}
	if err := testutils.PrepareTempDirs(); err != nil {
		fmt.Printf("prepare dirs failed, %v\n", err)
	}
	m.Run()
	if err := testutils.CleanupTempDirs(); err != nil {
		fmt.Printf("cleanup dirs failed, %v\n", err)
	}
}

func TestProcessUpload(t *testing.T) {
	convey.Convey("test process upload", t, func() {
		env := testutils.DummyTaskSchedule()

		type responseWriter struct {
			http.ResponseWriter
		}

		p := uploadProcess{responseWriter: &responseWriter{}}

		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, env.Scheduler).
			ApplyMethodReturn(env.TaskCtx, "UpdateStatus", nil).
			ApplyPrivateMethod(p, "receiveFile", func(uploadProcess, taskschedule.TaskContext) error { return nil }).
			ApplyPrivateMethod(p, "verifyFile", func(uploadProcess, taskschedule.TaskContext) error { return nil }).
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

		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, env.Scheduler).
			ApplyMethodReturn(env.TaskCtx, "UpdateStatus", nil).
			ApplyMethodReturn(env.TaskCtx, "GracefulShutdown", nil).
			ApplyMethodReturn(env.TaskCtx, "UpdateLiveness", nil).
			ApplyFunc(utils.WithDiskPressureProtect, testutils.WithoutDiskPressureProtect)
		defer patch.Reset()
		p := uploadProcess{httpRequest: &http.Request{Body: io.NopCloser(strings.NewReader("hello"))}}
		err := p.receiveFile(env.TaskCtx)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestVerifyFile(t *testing.T) {
	convey.Convey("test verify file", t, func() {
		env := testutils.DummyTaskSchedule()

		patch := gomonkey.ApplyFuncReturn(taskschedule.DefaultScheduler, env.Scheduler).
			ApplyMethodReturn(env.TaskCtx, "Spec", taskschedule.TaskSpec{Id: "1"})
		defer patch.Reset()
		filePath := filepath.Join(constants.LogDumpTempDir, "1"+common.TarGzSuffix)

		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, common.Mode600)
		convey.So(err, convey.ShouldBeNil)
		gw := gzip.NewWriter(file)
		tw := tar.NewWriter(gw)
		const data = "hello"
		err = tw.WriteHeader(&tar.Header{Name: "a.log", Size: int64(len(data)), Mode: common.Mode400})
		convey.So(err, convey.ShouldBeNil)
		_, err = tw.Write([]byte(data))
		convey.So(err, convey.ShouldBeNil)
		convey.So(tw.Close(), convey.ShouldBeNil)
		convey.So(gw.Close(), convey.ShouldBeNil)
		convey.So(file.Close(), convey.ShouldBeNil)

		convey.So(err, convey.ShouldBeNil)
		sha256Checksum, err := utils2.GetFileSha256(filePath)
		convey.So(err, convey.ShouldBeNil)
		p := uploadProcess{sha256Checksum: sha256Checksum}
		err = p.verifyFile(env.TaskCtx)
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
