// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
//go:build MEFEdge_SDK

// Package commands
package commands

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/installer/edgectl/common"
	"edge-installer/pkg/installer/edgectl/importcrl"
)

func TestImportCrlCmd(t *testing.T) {
	convey.Convey("test import crl cmd methods", t, importCrlCmdMethods)
	convey.Convey("test import crl cmd successful", t, importCrlCmdSuccess)
	convey.Convey("test import crl cmd failed", t, executeImportCrlFailed)
}

func importCrlCmdMethods() {
	convey.So(ImportCrlCmd().Name(), convey.ShouldEqual, common.ImportCrl)
	convey.So(ImportCrlCmd().Description(), convey.ShouldEqual, common.ImportCrlDesc)
	convey.So(ImportCrlCmd().LockFlag(), convey.ShouldBeFalse)
}

func importCrlCmdSuccess() {
	p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "", nil).
		ApplyMethodReturn(&importcrl.CrlImportFlow{}, "RunFlow", nil)
	defer p.Reset()
	err := ImportCrlCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	ImportCrlCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeImportCrlFailed() {
	convey.Convey("context is nil failed", func() {
		err := ImportCrlCmd().Execute(nil)
		expectErr := errors.New("context is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check crl path failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "./", test.ErrTest)
		defer p.Reset()
		err := ImportCrlCmd().Execute(ctx)
		expectErr := errors.New("check crl path failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("run crl import flow failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "", nil).
			ApplyMethodReturn(&importcrl.CrlImportFlow{}, "RunFlow", test.ErrTest)
		defer p.Reset()
		err := ImportCrlCmd().Execute(ctx)
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
	ImportCrlCmd().PrintOpLogFail(userRoot, ipLocalhost)
}
