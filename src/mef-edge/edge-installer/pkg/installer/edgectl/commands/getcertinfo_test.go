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

// Package commands
package commands

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/util"
	"edge-installer/pkg/installer/edgectl/common"
)

const testCert = `-----BEGIN CERTIFICATE-----
MIIErDCCAxSgAwIBAgIUGE2hCetqId/wezLWeToR6CHQ6YAwDQYJKoZIhvcNAQEL
BQAwQTELMAkGA1UEBhMCQ04xDzANBgNVBAoTBkh1YXdlaTEPMA0GA1UECxMGQXNj
ZW5kMRAwDgYDVQQDDAdodWJfc3ZyMB4XDTIzMDYyNzEyMzk1MloXDTMzMDYyNzEy
Mzk1MlowQTELMAkGA1UEBhMCQ04xDzANBgNVBAoTBkh1YXdlaTEPMA0GA1UECxMG
QXNjZW5kMRAwDgYDVQQDDAdodWJfc3ZyMIIBojANBgkqhkiG9w0BAQEFAAOCAY8A
MIIBigKCAYEAq6JPBWr62AAeWv0/cT3PpbwD1trKm+QOm+3ipHg+06EzjQsJ5G65
TFMi4gqjlFFdZEM+i6jjb4lAHLYTLM9F2jeTMM8QGcjAmp64f4jPenWnAdCF0SM/
sZFYgwdcHWbAcCRbxpt3HCasLFSri6CyyE3CCa/jNrtiYC4iNHOLOcNa3C6e0IBq
I5cuZxE3Il7DEfVZ4dVJQJDf7hcLn56CcKjQuLaT33XyRltzghCZyWdwdEd1B9DM
lSknu45+EoY2k7vGzW70U1jJT6QdMGlzjfFOY5656YRkZfI2IVjn+iv+wGulCXOE
8D/iSPRqxMdsW7sznzYa0I0mzg1wPTleqVbXbG28cDjN5luYoKny2oNR+cE7lf95
7N84HoCIGDVqJqOihVHpa0+dSc/0r9ZGDAV6Y0DV3vWlBWZLpRIXUDQd4ZtxFRsn
whJkq4K4DPNKAVofOAtEB05WBOPUx3PfdNMF4e6ROdLqLH188NUqVT4LpOg1OdNG
HU7fq927/vV7AgMBAAGjgZswgZgwDgYDVR0PAQH/BAQDAgKEMB0GA1UdJQQWMBQG
CCsGAQUFBwMCBggrBgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdDgQiBCC5
dI7PSI0Hwgod4bsds6o8M4PFLDPW37wCb7QJ5sZ4fDArBgNVHSMEJDAigCC5dI7P
SI0Hwgod4bsds6o8M4PFLDPW37wCb7QJ5sZ4fDANBgkqhkiG9w0BAQsFAAOCAYEA
hhknqoM2qP6NKBfh7Hf5+RttD7B2707ZRj8AUdQcHDt63Z81g37bLub+nIgyIfXU
PIWgbUcyIFavZPUcOHKTGUcwa1fbCug7SrCq8sjbolh1LBHlOyismxVJ/syV19IN
k4Y4HleuNsjEimqH150LlQYjIgbjdsv4QAQFG157+ktsPVeQMcjEGVfqUMlcnQw3
0r81wzoGL1sy7x0MuyqxWEe081ivdVqqj5mNI8aKZHkybq4+NPH+DGxBdr6jXUFu
/B2bBCkRPLg8rNp/fuN+AK4P6QZOpjFCdLttvKL7vrZGFi62Y6R74Q7R26klIMAc
IWINLbYQXWCJqQNBytbkQ61HugFAbVaIRFWD0MAr9qV/LHEJq6Yll0QdlFMJTMef
YQQU5rk4ynSsCN3tCwIIq7Z+oLYw8F73D5kq+SaLKM0cy/2fTpKXb7FAXknr5hcX
oPIW7kIti3EBn90JSqdVEkxJs4iGoGm3Ee/6Ak7ETjTAh36fovD4p2l1aKUzLC88
-----END CERTIFICATE-----`

func TestNewGetCertInfoCmd(t *testing.T) {
	patch := gomonkey.ApplyMethodReturn(&util.EdgeGUidMgr{}, "SetEUGidToEdge", nil).
		ApplyMethodReturn(&util.EdgeGUidMgr{}, "ResetEUGid", nil)
	defer patch.Reset()
	convey.Convey("test get cert info cmd methods", t, getCertInfoCmdMethods)
	convey.Convey("test get cert info cmd successful", t, getCertInfoCmdSuccess)
	convey.Convey("test get cert info cmd failed", t, func() {
		convey.Convey("execute get cert info failed", executeGetCertInfoCmdFailed)
		convey.Convey("print cert info failed", printCertFailed)
	})
}

func getCertInfoCmdMethods() {
	convey.So(NewGetCertInfoCmd().Name(), convey.ShouldEqual, common.GetCertInfo)
	convey.So(NewGetCertInfoCmd().Description(), convey.ShouldEqual, common.GetCertInfoDesc)
	convey.So(NewGetCertInfoCmd().BindFlag(), convey.ShouldBeTrue)
	convey.So(NewGetCertInfoCmd().LockFlag(), convey.ShouldBeTrue)
}

func getCertInfoCmdSuccess() {
	p := gomonkey.ApplyFuncReturn(NewGetCertInfoCmd, &getCertInfoCmd{certName: "center"}).
		ApplyFuncReturn(fileutils.LoadFile, []byte(testCert), nil)
	defer p.Reset()
	err := NewGetCertInfoCmd().Execute(ctx)
	convey.So(err, convey.ShouldBeNil)
	NewGetCertInfoCmd().PrintOpLogOk(userRoot, ipLocalhost)
}

func executeGetCertInfoCmdFailed() {
	convey.Convey("ctx is nil failed", func() {
		err := NewGetCertInfoCmd().Execute(nil)
		expectErr := errors.New("ctx is nil")
		convey.So(err, convey.ShouldResemble, expectErr)
		NewGetCertInfoCmd().PrintOpLogFail(userRoot, ipLocalhost)
	})

	convey.Convey("unsupported cert name parameter", func() {
		p := gomonkey.ApplyFuncReturn(NewGetCertInfoCmd, &getCertInfoCmd{certName: "cloud"})
		defer p.Reset()
		err := NewGetCertInfoCmd().Execute(ctx)
		expectErr := errors.New("the certificate name is not supported")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func printCertFailed() {
	convey.Convey("load cert failed", func() {
		p := gomonkey.ApplyFuncReturn(NewGetCertInfoCmd, &getCertInfoCmd{certName: "center"}).
			ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer p.Reset()
		err := NewGetCertInfoCmd().Execute(ctx)
		expectErr := errors.New("print cert failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get ca cert chain mgr failed", func() {
		p := gomonkey.ApplyFuncReturn(NewGetCertInfoCmd, &getCertInfoCmd{certName: "center"}).
			ApplyFuncReturn(fileutils.LoadFile, []byte{}, nil)
		defer p.Reset()
		err := NewGetCertInfoCmd().Execute(ctx)
		expectErr := errors.New("print cert failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
