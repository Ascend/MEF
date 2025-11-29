// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package innercommands
package innercommands

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/installer/common"
)

var exchangeFlow = NewExchangeCaFlow("./import", "./export", pathmgr.NewConfigPathMgr("./"))

func TestExchangeCertsFlow(t *testing.T) {
	convey.Convey("test exchange certs flow successful", t, exchangeCertsFlowSuccess)
	convey.Convey("test exchange certs flow failed", t, func() {
		convey.Convey("check exchange ca param failed", checkParamTaskFailed)
		convey.Convey("get uid or gid failed", getUidOrGidFailed)
		convey.Convey("import ca cert failed", importCaTaskFailed)
		convey.Convey("make sure edge certs failed", makeSureEdgeCertsFailed)
		convey.Convey("export ca cert failed", exportCaTaskFailed)
	})
}

func exchangeCertsFlowSuccess() {
	p := gomonkey.ApplyPrivateMethod(&checkParamTask{}, "runTask", func() error { return nil }).
		ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
		ApplyMethodReturn(&common.ImportCaTask{}, "RunTask", nil).
		ApplyFuncSeq(path.GetInstallRootDir, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"./", nil}},
			{Values: gomonkey.Params{"./", nil}},
		}).
		ApplyMethodReturn(&common.GenerateCertsTask{}, "MakeSureEdgeCerts", nil).
		ApplyFuncReturn(fileutils.CopyFile, nil)
	defer p.Reset()
	err := exchangeFlow.RunTasks()
	convey.So(err, convey.ShouldBeNil)
	ExchangeCertsCmd().PrintOpLogOk("root", "localhost")
}

func checkParamTaskFailed() {
	expectErr := errors.New("check exchange ca param failed")
	convey.Convey("importPath file check failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "", test.ErrTest)
		defer p.Reset()
		err := exchangeFlow.RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("exportPath dir check failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "", nil).
			ApplyFuncReturn(fileutils.RealDirCheck, "", test.ErrTest)
		defer p.Reset()
		err := exchangeFlow.RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("exportPath file check failed", func() {
		p := gomonkey.ApplyFuncSeq(fileutils.RealFileCheck, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", test.ErrTest}},
		}).
			ApplyFuncReturn(fileutils.RealDirCheck, "", nil).
			ApplyFuncReturn(fileutils.IsExist, true)
		defer p.Reset()
		err := exchangeFlow.RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func getUidOrGidFailed() {
	p := gomonkey.ApplyPrivateMethod(&checkParamTask{}, "runTask", func() error { return nil })
	defer p.Reset()

	convey.Convey("get uid failed", func() {
		p1 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), test.ErrTest)
		defer p1.Reset()
		err := exchangeFlow.RunTasks()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})

	convey.Convey("get gid failed", func() {
		p2 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
			ApplyFuncReturn(envutils.GetGid, uint32(0), test.ErrTest)
		defer p2.Reset()
		err := exchangeFlow.RunTasks()
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func importCaTaskFailed() {
	p := gomonkey.ApplyPrivateMethod(&checkParamTask{}, "runTask", func() error { return nil }).
		ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
		ApplyMethodReturn(&common.ImportCaTask{}, "RunTask", test.ErrTest)
	defer p.Reset()
	err := exchangeFlow.RunTasks()
	expectErr := errors.New("import ca failed")
	convey.So(err, convey.ShouldResemble, expectErr)
}

func makeSureEdgeCertsFailed() {
	p := gomonkey.ApplyPrivateMethod(&checkParamTask{}, "runTask", func() error { return nil }).
		ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
		ApplyMethodReturn(&common.ImportCaTask{}, "RunTask", nil)
	defer p.Reset()

	convey.Convey("get install root dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(path.GetInstallRootDir, "", test.ErrTest)
		defer p1.Reset()
		err := exchangeFlow.RunTasks()
		expectErr := errors.New("get install root dir failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("make sure certs failed", func() {
		p2 := gomonkey.ApplyFuncReturn(path.GetInstallRootDir, "./", nil).
			ApplyMethodReturn(&common.GenerateCertsTask{}, "MakeSureEdgeCerts", test.ErrTest)
		defer p2.Reset()
		err := exchangeFlow.RunTasks()
		expectErr := errors.New("make sure certs failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func exportCaTaskFailed() {
	p := gomonkey.ApplyPrivateMethod(&checkParamTask{}, "runTask", func() error { return nil }).
		ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
		ApplyMethodReturn(&common.ImportCaTask{}, "RunTask", nil).
		ApplyFuncReturn(path.GetInstallRootDir, "./", nil).
		ApplyMethodReturn(&common.GenerateCertsTask{}, "MakeSureEdgeCerts", nil).
		ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
	defer p.Reset()
	err := exchangeFlow.RunTasks()
	expectErr := errors.New("export ca failed")
	convey.So(err, convey.ShouldResemble, expectErr)
}
