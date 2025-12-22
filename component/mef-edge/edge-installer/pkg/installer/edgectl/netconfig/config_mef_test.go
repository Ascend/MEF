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

// Package netconfig
package netconfig

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/envutils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/httpsmgr"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/terminal"
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/certmgr"
	"edge-installer/pkg/common/config"
	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

const (
	testToken   = "12345678910ABCDEFG"
	testRootDir = "/tmp/test_mef_config/MEFEdge"
	testRootCa  = `-----BEGIN CERTIFICATE-----
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
)

var (
	testRootCaPath = filepath.Join(testRootDir, "rootCa.crt")
	testParam      = Param{
		NetType:       constants.MEF,
		Ip:            "192.168.0.0",
		Port:          constants.DefaultWsPort,
		AuthPort:      constants.DefaultWsTestPort,
		RootCa:        testRootCaPath,
		TestConnect:   true,
		ConfigPathMgr: pathmgr.NewConfigPathMgr(filepath.Dir(testRootDir)),
	}
	testErr = errors.New("test error")
)

func setupMefCfg() error {
	var err error
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err = util.InitHwLogger(logConfig, logConfig); err != nil {
		return err
	}
	testCfgDir := pathmgr.NewConfigPathMgr(filepath.Dir(testRootDir)).GetConfigDir()
	if err = fileutils.CreateDir(testCfgDir, constants.Mode700); err != nil {
		return fmt.Errorf("create test config dir [%s] failed, error: %v", testCfgDir, err)
	}

	rootCa, err := os.OpenFile(testRootCaPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constants.Mode600)
	if err != nil {
		return fmt.Errorf("open test root ca cert [%s] failed, error: %v", testRootCaPath, err)
	}
	defer func() {
		if err = rootCa.Close(); err != nil {
			fmt.Printf("close test root ca cert [%s] failed, error: %v\n", testRootCaPath, err)
			return
		}
	}()

	if _, err = rootCa.Write([]byte(testRootCa)); err != nil {
		return fmt.Errorf("write content to test root ca cert failed, error: %v", err)
	}
	return nil
}

func teardownMefCfg() {
	if err := os.RemoveAll(testRootDir); err != nil {
		hwlog.RunLog.Errorf("remove root dir for mef config failed, error: %v", err)
	}
}

func TestMefConfigFlow(t *testing.T) {
	if err := setupMefCfg(); err != nil {
		fmt.Printf("setup test environment for mef config failed: %v\n", err)
		return
	}
	defer teardownMefCfg()

	convey.Convey("test mef net config flow failed", t, func() {
		convey.Convey("check parameters task failed", func() {
			convey.Convey("check param ip failed", checkParamIpFailed)
			convey.Convey("check param port failed", checkParamPortFailed)
			convey.Convey("check param root ca failed", checkParamRootCaFailed)
		})

		convey.Convey("set config to db task failed", func() {
			convey.Convey("get token failed", getTokenFailed)
			convey.Convey("test connection failed", testConnectionFailed)
			convey.Convey("set net manager to db failed", setNetManagerToDbFailed)
		})

		convey.Convey("import root ca task failed", func() {
			convey.Convey("import root ca failed", importRootCaFailed)
			convey.Convey("get ca path failed", getCaPathFailed)
			convey.Convey("process root ca failed", processRootCaFailed)
			convey.Convey("remove mef invalid certs failed", removeMefInvalidCertsFailed)
		})
	})

	convey.Convey("test mef net config flow successful", t, MefConfigSuccess)
}

func MefConfigSuccess() {
	p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
		ApplyFuncReturn(kmc.EncryptContent, []byte(testToken), nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil).
		ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "Get", []byte{}, nil).
		ApplyFuncReturn(config.SetNetManager, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil).
		ApplyFuncReturn(removeFilesByMEFEdgeUser, nil)
	defer p.Reset()

	err := NewMefConfigFlow(testParam).RunTasks()
	convey.So(err, convey.ShouldBeNil)
}

func checkParamIpFailed() {
	testInvalidParam := testParam
	convey.Convey("param ip is invalid", func() {
		testInvalidParam.Ip = "255.255.255.255"
		err := NewMefConfigFlow(testInvalidParam).RunTasks()
		expectErr := errors.New("param ip is invalid")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get local ip failed", func() {
		p := gomonkey.ApplyFuncReturn(net.InterfaceAddrs, nil, testErr)
		defer p.Reset()
		testInvalidParam.Ip = "127.0.0.1"
		err := NewMefConfigFlow(testInvalidParam).RunTasks()
		expectErr := errors.New("get local ip failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("param ip is the same as local ip", func() {
		testInvalidParam.Ip = "127.0.0.1"
		err := NewMefConfigFlow(testInvalidParam).RunTasks()
		expectErr := errors.New("param ip is the same as local ip")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func checkParamPortFailed() {
	testInvalidParam := testParam
	convey.Convey("param port is out of range", func() {
		testInvalidParam.Port = 0
		err := NewMefConfigFlow(testInvalidParam).RunTasks()
		expectErr := fmt.Errorf("param port is out of range [%d, %d]", constants.MinPort, constants.MaxPort)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("param auth_port is out of range", func() {
		testInvalidParam.Port = constants.DefaultWsPort
		testInvalidParam.AuthPort = 0
		err := NewMefConfigFlow(testInvalidParam).RunTasks()
		expectErr := fmt.Errorf("param auth_port is out of range [%d, %d]", constants.MinPort, constants.MaxPort)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("auth_port cannot equal to port", func() {
		testInvalidParam.Port = constants.DefaultWsPort
		testInvalidParam.AuthPort = constants.DefaultWsPort
		err := NewMefConfigFlow(testInvalidParam).RunTasks()
		expectErr := errors.New("auth_port cannot equal to port")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func checkParamRootCaFailed() {
	convey.Convey("param root_ca is invalid", func() {
		symlinkRootCa := filepath.Join(testRootDir, "rootCa_symlink.crt")
		if err := os.Symlink(testRootCaPath, symlinkRootCa); err != nil {
			hwlog.RunLog.Errorf("create symlink for root ca failed, error: %v", err)
			return
		}
		testInvalidParam := testParam
		testInvalidParam.RootCa = symlinkRootCa
		err := NewMefConfigFlow(testInvalidParam).RunTasks()
		expectErr := errors.New("param root_ca is invalid")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("create temp cert dir failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.CreateDir, testErr)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("create temp cert dir failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("save root ca to tmp_certs failed", func() {
		p := gomonkey.ApplyMethodReturn(&certmgr.CertManager{}, "SaveCertByFile", testErr)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := fmt.Errorf("save root ca to %s failed", constants.TmpCerts)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("check ca content failed", func() {
		p := gomonkey.ApplyFuncReturn(x509.CheckCertsChainReturnContent, nil, testErr)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("check importing cert failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get file sha256 sum failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.GetFileSha256, "", testErr)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("get file sha256 sum failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func getTokenFailed() {
	convey.Convey("get token failed", func() {
		p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, nil, testErr)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("get token failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("input token length invalid", func() {
		p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte("123456"), nil)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("input token length invalid")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("token complex does not meet the requirement", func() {
		p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte("12345678910111213"), nil)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("token complex does not meet the requirement")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get kmc config failed when encrypt token", func() {
		p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
			ApplyFuncReturn(util.GetKmcConfig, nil, testErr)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, testErr)
	})

	convey.Convey("encrypt token failed", func() {
		p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
			ApplyFuncReturn(kmc.EncryptContent, nil, testErr)
		defer p.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, testErr)
	})
}

func testConnectionFailed() {
	expectErr := errors.New("test connection between MEF Edge and Center failed")
	p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
		ApplyFuncReturn(kmc.EncryptContent, []byte(testToken), nil)
	defer p.Reset()

	convey.Convey("decrypt token failed", func() {
		p1 := gomonkey.ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), testErr)
		defer p1.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, testErr)
	})

	convey.Convey("auth failed by center: token is incorrect", func() {
		patchErr := errors.New("https return error status code: 401")
		p2 := gomonkey.ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil).
			ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "Get", nil, patchErr)
		defer p2.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("auth failed by center: ip is lock", func() {
		patchErr := errors.New("https return error status code: 423")
		p3 := gomonkey.ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil).
			ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "Get", nil, patchErr)
		defer p3.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func setNetManagerToDbFailed() {
	p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
		ApplyFuncReturn(kmc.EncryptContent, []byte(testToken), nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil).
		ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "Get", []byte{}, nil)
	defer p.Reset()

	convey.Convey("get install root dir failed", func() {
		p1 := gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, testErr)
		defer p1.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("get config path manager failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set net manager to database failed", func() {
		p2 := gomonkey.ApplyFuncReturn(config.SetNetManager, testErr)
		defer p2.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("set net manager to database failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func importRootCaFailed() {
	p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
		ApplyFuncReturn(kmc.EncryptContent, []byte(testToken), nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil).
		ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "Get", []byte{}, nil).
		ApplyFuncReturn(config.SetNetManager, nil)
	defer p.Reset()

	convey.Convey("get edge user id failed", func() {
		p1 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), testErr)
		defer p1.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("get edge user id failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get edge group id failed", func() {
		p2 := gomonkey.ApplyFuncReturn(envutils.GetUid, uint32(0), nil).
			ApplyFuncReturn(envutils.GetGid, uint32(0), testErr)
		defer p2.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("get edge group id failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func getCaPathFailed() {
	expectErr := errors.New("get root ca cert paths failed")
	p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
		ApplyFuncReturn(kmc.EncryptContent, []byte(testToken), nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil).
		ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "Get", []byte{}, nil).
		ApplyFuncReturn(config.SetNetManager, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()

	convey.Convey("temp root ca or backup root ca path is soft link", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.IsSoftLink, testErr)
		defer p1.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set dest cert dir owner failed", func() {
		p3 := gomonkey.ApplyFuncReturn(fileutils.SetPathOwnerGroup, testErr)
		defer p3.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func processRootCaFailed() {
	expectErr := errors.New("import root ca failed")
	p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
		ApplyFuncReturn(kmc.EncryptContent, []byte(testToken), nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil).
		ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "Get", []byte{}, nil).
		ApplyFuncReturn(config.SetNetManager, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()

	convey.Convey("remove old root ca certs failed", func() {
		p1 := gomonkey.ApplyFuncReturn(envutils.RunCommandWithUser, "", testErr)
		defer p1.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("copy temp root ca to edge-main failed", func() {
		p2 := gomonkey.ApplyFuncSeq(envutils.RunCommandWithUser, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", testErr}},
		})
		defer p2.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("copy temp backup root ca to edge-main failed", func() {
		p3 := gomonkey.ApplyFuncSeq(envutils.RunCommandWithUser, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", testErr}},
		})
		defer p3.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func removeMefInvalidCertsFailed() {
	p := gomonkey.ApplyFuncReturn(terminal.ReadPasswordWithTimeout, []byte(testToken), nil).
		ApplyFuncReturn(kmc.EncryptContent, []byte(testToken), nil).
		ApplyFuncReturn(kmc.DecryptContent, []byte(testToken), nil).
		ApplyMethodReturn(&httpsmgr.HttpsRequest{}, "Get", []byte{}, nil).
		ApplyFuncReturn(config.SetNetManager, nil).
		ApplyFuncReturn(envutils.GetUid, uint32(os.Geteuid()), nil).
		ApplyFuncReturn(envutils.GetGid, uint32(os.Getegid()), nil)
	defer p.Reset()

	convey.Convey("remove invalid certs failed", func() {
		p1 := gomonkey.ApplyFuncReturn(removeFilesByMEFEdgeUser, testErr)
		defer p1.Reset()
		err := NewMefConfigFlow(testParam).RunTasks()
		expectErr := errors.New("remove invalid certs failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
