// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.
//go:build MEFEdge_SDK

// Package handlers for testing dump log handler
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/modulemgr"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/edge-main/common/cloudcert"
	"edge-installer/pkg/edge-main/common/configpara"
)

const testDumpLogMsg = `{
    "header":{
        "id":"5dea927e-a961-458b-9a6a-b7714e446a55",
        "timestamp":1695454325180,
        "isSync":true
    },
    "route":{
        "source":"websocket",
        "destination":"handler-manager",
        "operation":"post",
        "resource":"/logmgmt/dump/task"
    },
    "content":{}
}`

var (
	dumpLogMsg     model.Message
	dumpLogContent = DumpLogReq{
		Module: "edgeNode",
		TaskId: "dumpSingleNodeLog.2102312NSF10K8000130.859a88487aac6f2f12d4e094",
	}
	process = dumpLogProcess{handler: getDumpLogHandler(), taskId: dumpLogContent.TaskId}
)

func setupDumpLogHandler() error {
	if err := json.Unmarshal([]byte(testDumpLogMsg), &dumpLogMsg); err != nil {
		hwlog.RunLog.Errorf("unmarshal test dump log message failed, error: %v", err)
		return err
	}
	if err := dumpLogMsg.FillContent(dumpLogContent); err != nil {
		hwlog.RunLog.Errorf("fill test dump log content failed, error: %v", err)
		return err
	}
	return nil
}

func TestDumpLogHandler(t *testing.T) {
	if err := setupDumpLogHandler(); err != nil {
		panic(err)
	}

	p := gomonkey.ApplyPrivateMethod(&dumpLogProcess{}, "process", func(*dumpLogProcess) error { return nil })
	defer p.Reset()

	convey.Convey("test dump log handler should be success", t, dumpLogHandlerSuccess)
	convey.Convey("test dump log handler should be failed, parse and check args failed", t, parseAndCheckArgsFailed)
	convey.Convey("test dump log handler should be failed, dump log handler busy", t, dumpLogHandlerBusyFailed)
}

func dumpLogHandlerSuccess() {
	p1 := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
	defer p1.Reset()
	err := getDumpLogHandler().Handle(&dumpLogMsg)
	convey.So(err, convey.ShouldBeNil)
}

func parseAndCheckArgsFailed() {
	invalidDumpLogMsg := dumpLogMsg
	convey.Convey("parse content failed", func() {
		p1 := gomonkey.ApplyMethodReturn(&model.Message{}, "ParseContent", test.ErrTest)
		defer p1.Reset()
		err := getDumpLogHandler().Handle(&dumpLogMsg)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("failed to parse or check args,"+
			" parma convert error: %v", test.ErrTest))
	})

	convey.Convey("invalid Module argument", func() {
		err := invalidDumpLogMsg.FillContent(DumpLogReq{Module: ""})
		convey.So(err, convey.ShouldBeNil)

		err = getDumpLogHandler().Handle(&invalidDumpLogMsg)
		convey.So(err, convey.ShouldResemble, errors.New("failed to parse or check args, invalid argument"))
	})

	convey.Convey("invalid TaskId argument", func() {
		err := invalidDumpLogMsg.FillContent(DumpLogReq{
			Module: "edgeNode",
			TaskId: "2102312NSF10K8000130.859a88487aac6f2f12d4e094",
		})
		convey.So(err, convey.ShouldBeNil)

		err = getDumpLogHandler().Handle(&invalidDumpLogMsg)
		convey.So(err, convey.ShouldResemble, errors.New("failed to parse or check args, invalid argument"))
	})
}

func dumpLogHandlerBusyFailed() {
	handler := getDumpLogHandler()
	handler.running = 1
	convey.Convey("dump log handler busy, feedback error success", func() {
		p1 := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, nil)
		defer p1.Reset()
		err := handler.Handle(&dumpLogMsg)
		convey.So(err, convey.ShouldResemble, errors.New("dump log handler busy"))
	})

	convey.Convey("dump log handler busy, feedback error failed", func() {
		p1 := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, test.ErrTest)
		defer p1.Reset()
		err := handler.Handle(&dumpLogMsg)
		convey.So(err, convey.ShouldResemble, errors.New("dump log handler busy"))
	})
}

func TestDumpLogProcess(t *testing.T) {
	p := gomonkey.ApplyFuncReturn(configpara.GetNetConfig, config.NetManager{NetType: "MEF", IP: "127.0.0.1"}).
		ApplyPrivateMethod(&dumpLogProcess{}, "packLogs", func(*dumpLogProcess) error { return nil }).
		ApplyFuncReturn(cloudcert.GetEdgeHubCertInfo, &certutils.TlsCertInfo{}, nil).
		ApplyPrivateMethod(&dumpLogProcess{}, "uploadLogs", func(*dumpLogProcess) error { return nil })
	defer p.Reset()

	convey.Convey("test dump log process should be success", t, dumpLogProcessSuccess)
	convey.Convey("test dump log process should be failed, get net config failed", t, getNetConfigFailed)
	convey.Convey("test dump log process should be failed, get edge hub cert info failed", t, getCertInfoFailed)
	convey.Convey("test dump log process should be failed, pack logs failed", t, packLogsFailed)
	convey.Convey("test dump log process should be failed, upload logs failed", t, uploadLogsFailed)

}

func dumpLogProcessSuccess() {
	err := process.process()
	convey.So(err, convey.ShouldBeNil)
}

func getNetConfigFailed() {
	p1 := gomonkey.ApplyFuncReturn(configpara.GetNetConfig, config.NetManager{})
	defer p1.Reset()
	err := process.process()
	convey.So(err, convey.ShouldResemble, errors.New("failed to get net config, ip is invalid"))
}

func getCertInfoFailed() {
	p1 := gomonkey.ApplyFuncReturn(cloudcert.GetEdgeHubCertInfo, &certutils.TlsCertInfo{}, test.ErrTest)
	defer p1.Reset()
	err := process.process()
	convey.So(err, convey.ShouldResemble, errors.New("failed to get tls cert info"))
}

func packLogsFailed() {
	p1 := gomonkey.ApplyPrivateMethod(&dumpLogProcess{}, "packLogs", func(*dumpLogProcess) error { return test.ErrTest })
	defer p1.Reset()
	err := process.process()
	convey.So(err, convey.ShouldResemble, errors.New("failed to pack logs"))
}

func uploadLogsFailed() {
	p1 := gomonkey.ApplyPrivateMethod(&dumpLogProcess{}, "uploadLogs", func(*dumpLogProcess) error { return test.ErrTest })
	defer p1.Reset()
	err := process.process()
	convey.So(err, convey.ShouldResemble, errors.New("failed to upload logs"))
}

func TestPackLogs(t *testing.T) {
	convey.Convey("failed to create request message", t, func() {
		p := gomonkey.ApplyFuncReturn(model.NewMessage, nil, test.ErrTest)
		defer p.Reset()
		err := process.packLogs()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("failed to create request message, %v", test.ErrTest))
	})

	convey.Convey("failed to send request", t, func() {
		p := gomonkey.ApplyFuncReturn(modulemgr.SendMessage, test.ErrTest)
		defer p.Reset()
		err := process.packLogs()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("failed to send request, %v", test.ErrTest))
	})
}

func prepareLogCollectTempFile(file string) error {
	if err := fileutils.MakeSureDir(file); err != nil {
		hwlog.RunLog.Errorf("create test log collect temp dir failed, error: %v", err)
		return err
	}
	if err := fileutils.CreateFile(file, constants.Mode600); err != nil {
		hwlog.RunLog.Errorf("create test log collect temp file failed, error: %v", err)
		return err
	}
	return nil
}

func TestUploadLogs(t *testing.T) {
	testLogCollectTempDir := "/tmp/test_mef_logcollect"
	testLocalFilePath := filepath.Join(testLogCollectTempDir, constants.EdgeMain, constants.LogCollectTempFileName)
	if err := prepareLogCollectTempFile(testLocalFilePath); err != nil {
		panic(err)
	}
	defer func() {
		if err := fileutils.DeleteAllFileWithConfusion(testLogCollectTempDir); err != nil {
			hwlog.RunLog.Errorf("clear test log collect temp dir failed, error: %v", err)
		}
	}()

	p := gomonkey.ApplyGlobalVar(&localFilePath, testLocalFilePath).
		ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "PostFile", []byte(constants.OK), nil)
	defer p.Reset()

	convey.Convey("upload logs should be success", t, uploadLogsSuccess)
	convey.Convey("upload logs should be failed, open local file failed", t, openLocalFileFailed)
	convey.Convey("upload logs should be failed, check local file failed", t, checkLocalFileFailed)
	convey.Convey("upload logs should be failed, calculate checksum failed", t, calculateChecksumFailed)
	convey.Convey("upload logs should be failed, upload file failed", t, uploadFileFailed)
	convey.Convey("upload logs should be failed, unexpected response from center", t, unexpectedResponse)
}

func uploadLogsSuccess() {
	err := process.uploadLogs()
	convey.So(err, convey.ShouldBeNil)
}

func openLocalFileFailed() {
	p1 := gomonkey.ApplyFuncReturn(os.Open, nil, test.ErrTest)
	defer p1.Reset()
	err := process.uploadLogs()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("failed to open local file, %v", test.ErrTest))
}

func checkLocalFileFailed() {
	p1 := gomonkey.ApplyMethodReturn(&fileutils.FileLinkChecker{}, "Check", test.ErrTest)
	defer p1.Reset()
	err := process.uploadLogs()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("failed to check temp file, %v", test.ErrTest))
}

func calculateChecksumFailed() {
	p1 := gomonkey.ApplyFuncReturn(io.Copy, int64(0), test.ErrTest)
	defer p1.Reset()
	err := process.uploadLogs()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("failed to calculate checksum, %v", test.ErrTest))
}

func uploadFileFailed() {
	p1 := gomonkey.ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "PostFile", nil, test.ErrTest)
	defer p1.Reset()
	err := process.uploadLogs()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("failed to upload file, %v", test.ErrTest))
}

func unexpectedResponse() {
	p1 := gomonkey.ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "PostFile", []byte(constants.Failed), nil)
	defer p1.Reset()
	err := process.uploadLogs()
	convey.So(err, convey.ShouldResemble, fmt.Errorf("unexpected response from center: %s", []byte(constants.Failed)))
}
