// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package innercommands
package innercommands

import (
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

func TestExchangeCertsCmd(t *testing.T) {
	convey.Convey("test exchange certs cmd methods", t, exchangeCertsCmdMethods)
	convey.Convey("test exchange certs cmd successful", t, exchangeCertsCmdSuccess)
	convey.Convey("test exchange certs cmd failed", t, func() {
		convey.Convey("ctx is nil", exchangeCtxIsNilFailed)
		convey.Convey("get install root dir failed", getInstallRootDirFailed)
		convey.Convey("execute exchange flow failed", exchangeCaFlowFailed)
	})
}

func exchangeCertsCmdMethods() {
	convey.So(ExchangeCertsCmd().Name(), convey.ShouldEqual, common.ExchangeCertsCmd)
	convey.So(ExchangeCertsCmd().Description(), convey.ShouldEqual, common.ExchangeCertsDesc)
	convey.So(ExchangeCertsCmd().BindFlag(), convey.ShouldBeTrue)
	convey.So(ExchangeCertsCmd().LockFlag(), convey.ShouldBeFalse)
}

func exchangeCertsCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(ExchangeCaFlow{}, "RunTasks", nil)
	defer p.Reset()
	err := ExchangeCertsCmd().Execute(&common.Context{})
	convey.So(err, convey.ShouldBeNil)
	ExchangeCertsCmd().PrintOpLogOk("root", "localhost")
}

func exchangeCtxIsNilFailed() {
	err := ExchangeCertsCmd().Execute(nil)
	expectErr := errors.New("ctx is nil")
	convey.So(err, convey.ShouldResemble, expectErr)
	ExchangeCertsCmd().PrintOpLogFail("root", "localhost")
}

func getInstallRootDirFailed() {
	p := gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, test.ErrTest)
	defer p.Reset()
	err := ExchangeCertsCmd().Execute(&common.Context{})
	expectErr := errors.New("get config path manager failed")
	convey.So(err, convey.ShouldResemble, expectErr)
}

func exchangeCaFlowFailed() {
	p := gomonkey.ApplyMethodReturn(ExchangeCaFlow{}, "RunTasks", test.ErrTest)
	defer p.Reset()
	err := ExchangeCertsCmd().Execute(&common.Context{})
	expectErr := errors.New("execute exchange flow failed")
	convey.So(err, convey.ShouldResemble, expectErr)
}

func TestRecoveryCmd(t *testing.T) {
	convey.Convey("test recovery cmd methods", t, recoveryCmdMethods)
	convey.Convey("test recovery cmd successful", t, recoveryCmdSuccess)
	convey.Convey("test recovery cmd failed", t, recoveryCmdFailed)
}

func recoveryCmdMethods() {
	convey.So(RecoveryCmd().Name(), convey.ShouldEqual, common.RecoveryCmd)
	convey.So(RecoveryCmd().Description(), convey.ShouldEqual, common.RecoveryDesc)
	convey.So(RecoveryCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(RecoveryCmd().LockFlag(), convey.ShouldBeFalse)
}

func recoveryCmdSuccess() {
	p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(util.UnSetImmutable, nil).
		ApplyFuncReturn(os.RemoveAll, nil)
	defer p.Reset()
	err := RecoveryCmd().Execute(&common.Context{})
	convey.So(err, convey.ShouldBeNil)
	RecoveryCmd().PrintOpLogOk("root", "localhost")
}

func recoveryCmdFailed() {
	convey.Convey("ctx is nil", func() {
		err := RecoveryCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
		RecoveryCmd().PrintOpLogFail("root", "localhost")
	})

	convey.Convey("get install root dir failed", func() {
		p := gomonkey.ApplyFuncReturn(path.GetWorkPathMgr, nil, test.ErrTest)
		defer p.Reset()
		err := RecoveryCmd().Execute(&common.Context{})
		expectErr := errors.New("get work path manager failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("clean target install path failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(util.UnSetImmutable, test.ErrTest).
			ApplyFuncReturn(os.RemoveAll, test.ErrTest).
			ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer p.Reset()
		err := RecoveryCmd().Execute(&common.Context{})
		convey.So(err, convey.ShouldResemble, test.ErrTest)
	})
}

func TestPrepareEdgeCoreCmd(t *testing.T) {
	convey.Convey("test prepare edge core cmd methods", t, prepareEdgeCoreCmdMethods)
	convey.Convey("test prepare edge core cmd successful", t, prepareEdgeCoreCmdSuccess)
	convey.Convey("test prepare edge core cmd failed", t, prepareEdgeCoreCmdFailed)
}

func prepareEdgeCoreCmdMethods() {
	convey.So(PrepareEdgecoreCmd().Name(), convey.ShouldEqual, common.PrepareEdgecoreCmd)
	convey.So(PrepareEdgecoreCmd().Description(), convey.ShouldEqual, common.PrepareEdgecoreDesc)
	convey.So(PrepareEdgecoreCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(PrepareEdgecoreCmd().LockFlag(), convey.ShouldBeFalse)
}

func prepareEdgeCoreCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(&PrepareEdgecoreFlow{}, "Run", nil)
	defer p.Reset()
	err := PrepareEdgecoreCmd().Execute(&common.Context{})
	convey.So(err, convey.ShouldBeNil)
	PrepareEdgecoreCmd().PrintOpLogOk("root", "localhost")
}

func prepareEdgeCoreCmdFailed() {
	convey.Convey("ctx is nil", func() {
		err := PrepareEdgecoreCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
		PrepareEdgecoreCmd().PrintOpLogFail("root", "localhost")
	})

	convey.Convey("prepare edgecore pipe file failed", func() {
		p := gomonkey.ApplyMethodReturn(&PrepareEdgecoreFlow{}, "Run", test.ErrTest)
		defer p.Reset()
		err := PrepareEdgecoreCmd().Execute(&common.Context{})
		expectErr := errors.New("prepare edgecore pipe file failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func TestRecoverLogCmdCmd(t *testing.T) {
	convey.Convey("test recover log cmd methods", t, recoverLogCmdMethods)
	convey.Convey("test recover log cmd successful", t, recoverLogCmdSuccess)
	convey.Convey("test recover log cmd failed", t, recoverLogCmdFailed)
}

func recoverLogCmdMethods() {
	convey.So(NewRecoverLogCmd().Name(), convey.ShouldEqual, common.RecoverLogCmd)
	convey.So(NewRecoverLogCmd().Description(), convey.ShouldEqual, common.RecoverLogDesc)
	convey.So(NewRecoverLogCmd().BindFlag(), convey.ShouldBeFalse)
	convey.So(NewRecoverLogCmd().LockFlag(), convey.ShouldBeFalse)
}

func recoverLogCmdSuccess() {
	p := gomonkey.ApplyMethodReturn(&util.LogSyncMgr{}, "RecoverLogs", nil)
	defer p.Reset()
	err := NewRecoverLogCmd().Execute(&common.Context{})
	convey.So(err, convey.ShouldBeNil)
	NewRecoverLogCmd().PrintOpLogOk("root", "localhost")
}

func recoverLogCmdFailed() {
	convey.Convey("ctx is nil", func() {
		err := NewRecoverLogCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
		NewRecoverLogCmd().PrintOpLogFail("root", "localhost")
	})

	convey.Convey("recover log failed", func() {
		p := gomonkey.ApplyMethodReturn(&util.LogSyncMgr{}, "RecoverLogs", test.ErrTest)
		defer p.Reset()
		err := NewRecoverLogCmd().Execute(&common.Context{})
		expectErr := errors.New("recover log failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
