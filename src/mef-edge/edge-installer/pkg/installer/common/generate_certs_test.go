// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package common for testing generate certs task
package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/util"
)

func TestGenerateCertsTask(t *testing.T) {
	var p = gomonkey.ApplyFuncReturn(util.GetMefId, uint32(constants.EdgeUserGid), uint32(constants.EdgeUserGid), nil).
		ApplyFuncReturn(kmc.InitKmcCfg, nil).
		ApplyFuncReturn(util.GetKmcConfig, nil, nil).
		ApplyMethodReturn(&certutils.RootCertMgr{}, "NewRootCa", nil, nil).
		ApplyMethodReturn(&certutils.SelfSignCert{}, "CreateSignCert", nil).
		ApplyFuncReturn(fileutils.CopyFile, nil).
		ApplyFuncReturn(fileutils.SetPathPermission, nil).
		ApplyFuncReturn(fileutils.SetPathOwnerGroup, nil).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(fileutils.LoadFile, []byte("abc"), nil).
		ApplyFuncReturn(x509.LoadCertsFromPEM, nil, nil)
	defer p.Reset()

	convey.Convey("generate certs should be success", t, generateCertsSuccess)
	convey.Convey("generate certs should be failed, prepare root cert failed", t, prepareRootCaCertFailed)
	convey.Convey("generate certs should be failed, prepare component cert failed", t, prepareCompCertFailed)
	convey.Convey("generate certs should be failed, input is nil failed", t, inputIsNilFailed)
	convey.Convey("generate certs should be failed, copy ca failed", t, copyCaFailed)
	convey.Convey("generate certs should be failed, set cert perm failed", t, setCertPermFailed)
	convey.Convey("generate certs should be failed, set cert owner failed", t, setCertOwnerFailed)
	convey.Convey("generate certs should be success, clear root ca files failed", t, clearRootCaFailed)

	convey.Convey("make sure certs should be success when cert does not exist", t, certNotExist)
	convey.Convey("make sure certs should be success when load cert file failed", t, loadFileFailed)
	convey.Convey("make sure certs should be success when cert file is nil", t, certFileIsNil)
	convey.Convey("make sure certs should be success when load cert from pem failed", t, loadCertFromPemFailed)
	convey.Convey("make sure certs should be success when cert is overdue", t, certIsOverdue)
}

func getGenerateCertsTask() *GenerateCertsTask {
	generateCerts, err := NewGenerateCertsTask("/tmp/generate_cert_test")
	if err != nil {
		hwlog.RunLog.Errorf("new generate certs task failed, error: %v", err)
		return &GenerateCertsTask{}
	}
	return generateCerts
}

func generateCertsSuccess() {
	err := getGenerateCertsTask().Run()
	convey.So(err, convey.ShouldBeNil)
}

func prepareRootCaCertFailed() {
	convey.Convey("get kmc config for root cert failed", func() {
		p := gomonkey.ApplyFuncReturn(util.GetKmcConfig, nil, test.ErrTest)
		defer p.Reset()
		err := getGenerateCertsTask().Run()
		convey.So(err, convey.ShouldResemble, errors.New("get kmc config for root cert failed"))
	})

	convey.Convey("new root ca failed", func() {
		p := gomonkey.ApplyMethodReturn(&certutils.RootCertMgr{}, "NewRootCa", nil, test.ErrTest)
		defer p.Reset()
		err := getGenerateCertsTask().Run()
		convey.So(err, convey.ShouldResemble, errors.New("new root ca failed"))
	})
}

func prepareCompCertFailed() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 2},
		{Values: gomonkey.Params{test.ErrTest}},
	}
	var p1 = gomonkey.ApplyMethodSeq(&certutils.SelfSignCert{}, "CreateSignCert", outputs)
	defer p1.Reset()

	for _, component := range getGenerateCertsTask().components {
		err := getGenerateCertsTask().Run()
		convey.So(err, convey.ShouldResemble, fmt.Errorf("prepare %s component cert failed", component.name))
	}
}

func inputIsNilFailed() {
	getGenerateCertsTask().prepareEdgeCerts(nil)

	err := getGenerateCertsTask().prepareComponentCert(nil, "")
	convey.So(err, convey.ShouldResemble, errors.New("pointer rootCertMgr is nil"))
}

func copyCaFailed() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.CopyFile, test.ErrTest)
	defer p1.Reset()

	err := getGenerateCertsTask().Run()
	expectErr := fmt.Errorf("copy root ca to %s failed", getGenerateCertsTask().components[0].name)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func setCertPermFailed() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 2},
		{Values: gomonkey.Params{test.ErrTest}},

		{Values: gomonkey.Params{nil}, Times: 3},
		{Values: gomonkey.Params{test.ErrTest}},
	}
	var p1 = gomonkey.ApplyFuncSeq(fileutils.SetPathPermission, outputs)
	defer p1.Reset()

	const testCaseNum = 4
	expectErr := fmt.Errorf("set %s permission failed", getGenerateCertsTask().components[0].name)
	for i := 0; i < testCaseNum; i++ {
		err := getGenerateCertsTask().Run()
		convey.So(err, convey.ShouldResemble, expectErr)
	}
}

func setCertOwnerFailed() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.SetPathOwnerGroup, test.ErrTest)
	defer p1.Reset()

	err := getGenerateCertsTask().Run()
	expectErr := fmt.Errorf("set %s owner failed", getGenerateCertsTask().components[0].name)
	convey.So(err, convey.ShouldResemble, expectErr)
}

func clearRootCaFailed() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{test.ErrTest}},
		{Values: gomonkey.Params{test.ErrTest}},
	}
	var p1 = gomonkey.ApplyFuncSeq(fileutils.DeleteAllFileWithConfusion, outputs)
	defer p1.Reset()

	err := getGenerateCertsTask().Run()
	convey.So(err, convey.ShouldBeNil)
}

func certNotExist() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.IsExist, false).
		ApplyFuncReturn(x509.GetValidityPeriod, float64(100), nil)
	defer p1.Reset()

	err := getGenerateCertsTask().MakeSureEdgeCerts()
	convey.So(err, convey.ShouldBeNil)
}

func loadFileFailed() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
	defer p1.Reset()

	err := getGenerateCertsTask().MakeSureEdgeCerts()
	convey.So(err, convey.ShouldBeNil)
}

func certFileIsNil() {
	var p1 = gomonkey.ApplyFuncReturn(fileutils.LoadFile, nil, nil)
	defer p1.Reset()

	err := getGenerateCertsTask().MakeSureEdgeCerts()
	convey.So(err, convey.ShouldBeNil)
}

func loadCertFromPemFailed() {
	var p1 = gomonkey.ApplyFuncReturn(x509.LoadCertsFromPEM, nil, test.ErrTest)
	defer p1.Reset()

	err := getGenerateCertsTask().MakeSureEdgeCerts()
	convey.So(err, convey.ShouldBeNil)
}

func certIsOverdue() {
	outputs := []gomonkey.OutputCell{
		{Values: gomonkey.Params{float64(0), test.ErrTest}},

		{Values: gomonkey.Params{float64(80), nil}},
	}
	var p1 = gomonkey.ApplyFuncSeq(x509.GetValidityPeriod, outputs)
	defer p1.Reset()

	err := getGenerateCertsTask().MakeSureEdgeCerts()
	convey.So(err, convey.ShouldBeNil)

	err = getGenerateCertsTask().MakeSureEdgeCerts()
	convey.So(err, convey.ShouldBeNil)
}
