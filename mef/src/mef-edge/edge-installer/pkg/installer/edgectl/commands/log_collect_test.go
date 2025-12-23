// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package commands
package commands

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/logmgmt/logcollect"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/installer/edgectl/common"
)

const (
	logDir       = "/var/alog/MEFEdge_log"
	logBackupDir = "/home/log/MEFEdge_logbackup"
	tarGzPath    = "./testLog.tar.gz"
)

func TestLogCollectCmd(t *testing.T) {
	convey.Convey("test log collect cmd methods", t, logCollectCmdMethods)
	convey.Convey("test log collect cmd successful", t, logCollectCmdSuccess)
	convey.Convey("test log collect cmd failed", t, func() {
		convey.Convey("execute func failed", executeFailed)
		convey.Convey("get log dirs failed", getLogDirsFailed)
		convey.Convey("log collect func failed", logCollectFailed)
	})
}

func logCollectCmdMethods() {
	convey.So(LogCollectCmd().Name(), convey.ShouldEqual, common.CollectLog)
	convey.So(LogCollectCmd().Description(), convey.ShouldEqual, common.CollectLogDesc)
	convey.So(LogCollectCmd().BindFlag(), convey.ShouldBeTrue)
	convey.So(LogCollectCmd().LockFlag(), convey.ShouldBeFalse)
}

func logCollectCmdSuccess() {
	coll := logcollect.NewCollector(tarGzPath, []logcollect.LogGroup{}, 1, []string{tarGzPath})
	p := gomonkey.ApplyFuncReturn(LogCollectCmd, &logCollectCmd{module: "all", tarGzPath: tarGzPath}).
		ApplyFuncSeq(fileutils.ReadLink, []gomonkey.OutputCell{
			{Values: gomonkey.Params{logDir, nil}},
			{Values: gomonkey.Params{logBackupDir, nil}},
		}).
		ApplyMethodReturn(coll, "Collect", "", nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(fileutils.RealDirCheck, "", nil)
	defer p.Reset()
	err := LogCollectCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	LogCollectCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeFailed() {
	convey.Convey("container log collection is not supported yet", func() {
		p := gomonkey.ApplyFuncReturn(LogCollectCmd, &logCollectCmd{module: "APP"})
		defer p.Reset()
		err := LogCollectCmd().Execute(ctx)
		expectErr := errors.New("container log collection is not supported yet")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("unsupported module parameter", func() {
		p := gomonkey.ApplyFuncReturn(LogCollectCmd, &logCollectCmd{module: "container"})
		defer p.Reset()
		err := LogCollectCmd().Execute(ctx)
		expectErr := errors.New("unsupported module parameter")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
	LogCollectCmd().PrintOpLogFail(userRoot, ipLocalhost)
}

func getLogDirsFailed() {
	expectErr := errors.New("get log real dirs failed")
	convey.Convey("get log dir failed", func() {
		p := gomonkey.ApplyFuncReturn(LogCollectCmd, &logCollectCmd{module: "all"}).
			ApplyFuncReturn(fileutils.ReadLink, "", test.ErrTest)
		defer p.Reset()
		err := LogCollectCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get log backup dir failed", func() {
		p := gomonkey.ApplyFuncReturn(LogCollectCmd, &logCollectCmd{module: "all"}).
			ApplyFuncSeq(fileutils.ReadLink, []gomonkey.OutputCell{
				{Values: gomonkey.Params{logDir, nil}},
				{Values: gomonkey.Params{"", test.ErrTest}},
			})
		defer p.Reset()
		err := LogCollectCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func logCollectFailed() {
	p := gomonkey.ApplyFuncReturn(LogCollectCmd, &logCollectCmd{module: "all", tarGzPath: tarGzPath}).
		ApplyFuncSeq(fileutils.ReadLink, []gomonkey.OutputCell{
			{Values: gomonkey.Params{logDir, nil}},
			{Values: gomonkey.Params{logBackupDir, nil}},
		})
	defer p.Reset()

	convey.Convey("collect log failed", func() {
		coll := logcollect.NewCollector(tarGzPath, []logcollect.LogGroup{}, 1, []string{tarGzPath})
		p1 := gomonkey.ApplyMethodReturn(coll, "Collect", "", test.ErrTest).
			ApplyFuncReturn(fileutils.RealDirCheck, "", nil)
		defer p1.Reset()
		err := LogCollectCmd().Execute(ctx)
		expectErr := errors.New("collect log failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("chmod failed", func() {
		coll := logcollect.NewCollector(tarGzPath, []logcollect.LogGroup{}, 1, []string{tarGzPath})
		p2 := gomonkey.ApplyMethodReturn(coll, "Collect", "", nil).
			ApplyFuncReturn(os.Chmod, test.ErrTest).
			ApplyFuncReturn(fileutils.SetPathPermission, test.ErrTest).
			ApplyFuncReturn(fileutils.RealDirCheck, "", nil)
		defer p2.Reset()
		err := LogCollectCmd().Execute(ctx)
		expectErr := errors.New("chmod failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
