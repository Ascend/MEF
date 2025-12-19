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

// Package importcrl
package importcrl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

var (
	testPath           = "/tmp/test_import_crl/MEFEdge"
	configPathMgr      = pathmgr.NewConfigPathMgr(filepath.Dir(testPath))
	testImportCrlPath  = filepath.Join(testPath, "test.crl")
	testCenterPath     = filepath.Join(testPath, constants.Config, constants.EdgeMain, constants.MefCertImportPathName)
	testCenterCertPath = filepath.Join(testCenterPath, constants.RootCertName)
)

func setup() error {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := util.InitHwLogger(logConfig, logConfig); err != nil {
		return err
	}

	if err := fileutils.CreateDir(testCenterPath, constants.Mode700); err != nil {
		return fmt.Errorf("create center cert import dir failed, error: %v", err)
	}
	if err := fileutils.CreateFile(testCenterCertPath, constants.Mode600); err != nil {
		return err
	}
	if err := fileutils.CreateFile(testImportCrlPath, constants.Mode600); err != nil {
		return err
	}
	return nil
}

func teardown() {
	if err := os.RemoveAll(testPath); err != nil {
		hwlog.RunLog.Errorf("remove test path failed, error: %v", err)
	}
}

// TestMain run test main
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		fmt.Printf("setup test environment failed: %v\n", err)
		return
	}
	defer teardown()
	exitCode := m.Run()
	fmt.Printf("test complete, exitCode=%d\n", exitCode)
}

func TestCrlImportFlow(t *testing.T) {
	convey.Convey("test crl import successful", t, crlImportSuccess)
	convey.Convey("test crl import failed", t, func() {
		convey.Convey("set cert pair failed", setCertPairFailed)
		convey.Convey("check crl failed", checkCrlFailed)
		convey.Convey("crl import failed", func() {
			convey.Convey("copy crl to tmp failed", copyCrlToTmpFailed)
			convey.Convey("copy crl to edge main failed", copyCrlToEdgeMainFailed)
		})
	})
}

func crlImportSuccess() {
	p := gomonkey.ApplyFuncReturn(x509.CheckCrlsChainReturnContent, []byte{}, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
		ApplyFuncReturn(util.CreateBackupWithMefOwner, nil).
		ApplyFuncSeq(envutils.RunCommandWithUser, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", nil}},
		})
	defer p.Reset()
	crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, constants.MefCenterPeer)
	err := crlImportFlow.RunFlow()
	convey.So(err, convey.ShouldBeNil)
}

func setCertPairFailed() {
	crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, "")
	err := crlImportFlow.RunFlow()
	expectErr := errors.New("unsupported peer param")
	convey.So(err, convey.ShouldResemble, expectErr)
}

func checkCrlFailed() {
	convey.Convey("peer's cert has not yet imported failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.IsLexist, false)
		defer p.Reset()
		crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, constants.MefCenterPeer)
		err := crlImportFlow.RunFlow()
		expectErr := errors.New("peer's cert has not yet imported")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check crl chain failed", func() {
		p := gomonkey.ApplyFuncReturn(x509.CheckCrlsChainReturnContent, []byte{}, test.ErrTest)
		defer p.Reset()
		crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, constants.MefCenterPeer)
		err := crlImportFlow.RunFlow()
		expectErr := fmt.Errorf("check crl chain failed: %s", test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func copyCrlToTmpFailed() {
	p := gomonkey.ApplyFuncReturn(x509.CheckCrlsChainReturnContent, []byte{}, nil)
	defer p.Reset()

	convey.Convey("init tmp dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CreateDir, test.ErrTest)
		defer p1.Reset()
		crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, constants.MefCenterPeer)
		err := crlImportFlow.RunFlow()
		expectErr := fmt.Errorf("init tmp dir failed: %s", test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("copy file to tmp dir failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
		defer p2.Reset()
		crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, constants.MefCenterPeer)
		err := crlImportFlow.RunFlow()
		expectErr := fmt.Errorf("copy file to tmp dir failed: %s", test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set tmp dir path failed", func() {
		p3 := gomonkey.ApplyFuncReturn(fileutils.SetPathPermission, test.ErrTest)
		defer p3.Reset()
		crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, constants.MefCenterPeer)
		err := crlImportFlow.RunFlow()
		expectErr := fmt.Errorf("set tmp dir path failed: %s", test.ErrTest.Error())
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func copyCrlToEdgeMainFailed() {
	p := gomonkey.ApplyFuncReturn(x509.CheckCrlsChainReturnContent, []byte{}, nil)
	defer p.Reset()

	convey.Convey("copy temp crl to dst failed", func() {
		p1 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
			ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
			ApplyFuncReturn(envutils.RunCommandWithUser, "", test.ErrTest)
		defer p1.Reset()
		crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, constants.MefCenterPeer)
		err := crlImportFlow.RunFlow()
		expectErr := errors.New("copy temp crl to dst failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set save crl right failed", func() {
		p2 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(1225), nil).
			ApplyFuncReturn(envutils.GetGid, uint32(1225), nil).
			ApplyFuncSeq(envutils.RunCommandWithUser, []gomonkey.OutputCell{
				{Values: gomonkey.Params{"", nil}},
				{Values: gomonkey.Params{"", test.ErrTest}},
			})
		defer p2.Reset()
		crlImportFlow := NewCrlImportFlow(configPathMgr, testImportCrlPath, constants.MefCenterPeer)
		err := crlImportFlow.RunFlow()
		expectErr := errors.New("set save crl right failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
