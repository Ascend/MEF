// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package commands
package commands

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/veripkgutils"
	"edge-installer/pkg/installer/edgectl/common"
)

var testCrl = "./test.crl"

func TestUpdateCrlCmd(t *testing.T) {
	convey.Convey("test update crl cmd methods", t, updateCrlCmdMethods)
	convey.Convey("test update crl cmd successful", t, updateCrlCmdSuccess)
	convey.Convey("test update crl cmd failed", t, func() {
		convey.Convey("execute update crl failed", executeUpdateCrlFailed)
		convey.Convey("run update crl flow failed", runUpdateCrlFlowFailed)
	})
}

func updateCrlCmdMethods() {
	convey.So(UpdateCrlCmd().Name(), convey.ShouldEqual, common.UpdateCrl)
	convey.So(UpdateCrlCmd().Description(), convey.ShouldEqual, common.UpdateCrlDesc)
	convey.So(UpdateCrlCmd().BindFlag(), convey.ShouldBeTrue)
	convey.So(UpdateCrlCmd().LockFlag(), convey.ShouldBeTrue)
}

func updateCrlCmdSuccess() {
	p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "", nil).
		ApplyFuncReturn(common.InitEdgeOmResource, nil).
		ApplyFuncReturn(veripkgutils.PrepareVerifyCrl, true, testCrl, nil).
		ApplyFuncReturn(veripkgutils.UpdateLocalCrl, nil)
	defer p.Reset()
	err := UpdateCrlCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	UpdateCrlCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeUpdateCrlFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := UpdateCrlCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("param crl_path is invalid failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "./", test.ErrTest)
		defer p.Reset()
		err := UpdateCrlCmd().Execute(ctx)
		checkErr := errors.New("param crl_path is invalid")
		expectErr := fmt.Errorf("check param failed, error: %v", checkErr)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	UpdateCrlCmd().PrintOpLogFail(userRoot, ipLocalhost)
}

func runUpdateCrlFlowFailed() {
	convey.Convey("prepare crl for verifying package failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "", nil).
			ApplyFuncReturn(common.InitEdgeOmResource, nil).
			ApplyFuncReturn(veripkgutils.PrepareVerifyCrl, false, "", test.ErrTest)
		defer p.Reset()
		err := UpdateCrlCmd().Execute(ctx)
		expectErr := fmt.Errorf("update crl failed, error: %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("update crl file failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.RealFileCheck, "", nil).
			ApplyFuncReturn(common.InitEdgeOmResource, nil).
			ApplyFuncReturn(veripkgutils.PrepareVerifyCrl, true, testCrl, nil).
			ApplyFuncReturn(veripkgutils.UpdateLocalCrl, test.ErrTest)
		defer p.Reset()
		err := UpdateCrlCmd().Execute(ctx)
		expectErr := errors.New("update crl file failed")
		convey.So(err, convey.ShouldResemble, fmt.Errorf("update crl failed, error: %v", expectErr))
	})
}
