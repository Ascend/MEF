// Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.

// Package common
package common

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
	"huawei.com/mindx/common/x509"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
	"edge-installer/pkg/common/path/pathmgr"
	"edge-installer/pkg/common/util"
)

const testImportCert = `-----BEGIN CERTIFICATE-----
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

var (
	certName     = constants.RootCaName
	testPath     = "/tmp/test_import_ca_task"
	importPath   = filepath.Join(testPath, "import.crt")
	savePath     = filepath.Join(testPath, "save_path")
	saveCrt      = filepath.Join(savePath, certName)
	tempPath     = filepath.Join(testPath, constants.Config, constants.TmpCerts)
	importCaTask = InitImportCaTask(importPath, savePath, certName, uint32(os.Geteuid()), uint32(os.Getegid()))
)

func setupImportCaTask() error {
	if err := fileutils.MakeSureDir(importPath); err != nil {
		return fmt.Errorf("create import dir failed, error: %v", err)
	}
	importFile, err := os.OpenFile(importPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constants.Mode600)
	if err != nil {
		return fmt.Errorf("open import cert [%s] failed, error: %v", importPath, err)
	}
	defer func() {
		if err = importFile.Close(); err != nil {
			fmt.Printf("close import cert [%s] failed, error: %v\n", importPath, err)
			return
		}
	}()

	if _, err = importFile.Write([]byte(testImportCert)); err != nil {
		return fmt.Errorf("write content to import cert failed, error: %v", err)
	}
	return nil
}

func teardownImportCaTask() {
	if err := os.RemoveAll(testPath); err != nil {
		hwlog.RunLog.Errorf("remove test path [%s] failed, error: %v", testPath, err)
	}
}

func TestImportCaTask(t *testing.T) {
	patchTempPath := filepath.Join(testPath, constants.Config, constants.TmpCerts)
	p := gomonkey.ApplyMethodReturn(&pathmgr.ConfigPathMgr{}, "GetTempCertsDir", patchTempPath)
	defer p.Reset()

	if err := setupImportCaTask(); err != nil {
		fmt.Printf("setup import ca task test environment failed: %v\n", err)
		return
	}
	defer teardownImportCaTask()
	convey.Convey("test import ca task successful", t, importCaTaskSuccess)
	convey.Convey("test import ca task failed", t, func() {
		convey.Convey("check ca failed", checkCaFailed)
		convey.Convey("prepare save path failed", prepareSavePathFailed)
		convey.Convey("import ca failed", importCaFailed)
		convey.Convey("copy ca to temp failed", copyCaToTempFailed)
		convey.Convey("copy ca to edge main failed", copyCaToEdgeMainFailed)
	})
}

func importCaTaskSuccess() {
	if err := fileutils.CreateDir(savePath, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("create save path failed, error: %v", err)
		return
	}
	p := gomonkey.ApplyFuncReturn(util.CreateBackupWithMefOwner, nil).
		ApplyPrivateMethod(&ImportCaTask{}, "copyCaToTemp", func() error { return nil }).
		ApplyPrivateMethod(&ImportCaTask{}, "copyCaToEdgeMain", func() error { return nil })
	defer p.Reset()
	err := importCaTask.RunTask()
	convey.So(err, convey.ShouldBeNil)
}

func checkCaFailed() {
	convey.Convey("check importing ca cert failed", func() {
		p := gomonkey.ApplyFuncReturn(x509.CheckCertsChainReturnContent, nil, testErr)
		defer p.Reset()
		err := importCaTask.RunTask()
		expectErr := fmt.Errorf("check importing cert failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("get file sha256 sum failed", func() {
		p := gomonkey.ApplyFuncReturn(fileutils.GetFileSha256, "", testErr)
		defer p.Reset()
		err := importCaTask.RunTask()
		expectErr := errors.New("get file sha256 sum failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func prepareSavePathFailed() {
	if err := os.RemoveAll(savePath); err != nil {
		hwlog.RunLog.Errorf("delete existed save path failed, error: %v", err)
		return
	}

	convey.Convey("create save cert dir failed", func() {
		p := gomonkey.ApplyFuncReturn(envutils.RunCommandWithUser, "", testErr)
		defer p.Reset()
		err := importCaTask.RunTask()
		expectErr := fmt.Errorf("create dir [%s] failed", savePath)
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set save path right failed", func() {
		p := gomonkey.ApplyFuncSeq(envutils.RunCommandWithUser, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", testErr}},
		})
		defer p.Reset()
		err := importCaTask.RunTask()
		expectErr := fmt.Errorf("set path [%s] right failed", savePath)
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func importCaFailed() {
	if err := fileutils.CreateDir(savePath, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("create save path failed, error: %v", err)
		return
	}
	if err := fileutils.CreateFile(saveCrt, constants.Mode444); err != nil {
		hwlog.RunLog.Errorf("create save cert failed, error: %v", err)
		return
	}

	convey.Convey("get config path manager failed", func() {
		p := gomonkey.ApplyFuncReturn(path.GetConfigPathMgr, nil, testErr)
		defer p.Reset()
		err := importCaTask.RunTask()
		expectedErr := errors.New("get config path manager failed")
		convey.So(err, convey.ShouldResemble, expectedErr)
	})

	convey.Convey("delete original crt failed", func() {
		p := gomonkey.ApplyPrivateMethod(&ImportCaTask{}, "copyCaToTemp", func() error { return nil }).
			ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, testErr)
		defer p.Reset()
		err := importCaTask.RunTask()
		expectErr := errors.New("delete original crt failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func copyCaToTempFailed() {
	if err := fileutils.CreateDir(savePath, constants.Mode700); err != nil {
		hwlog.RunLog.Errorf("create save path failed, error: %v", err)
		return
	}
	if err := fileutils.CreateFile(saveCrt, constants.Mode444); err != nil {
		hwlog.RunLog.Errorf("create save cert failed, error: %v", err)
		return
	}

	convey.Convey("create temp crt path failed", func() {
		p1 := gomonkey.ApplyFuncReturn(fileutils.CreateDir, testErr)
		defer p1.Reset()
		err := importCaTask.RunTask()
		expectErr := errors.New("create temp crt path failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set temp dir right failed", func() {
		p2 := gomonkey.ApplyFuncReturn(fileutils.SetPathPermission, testErr)
		defer p2.Reset()
		err := importCaTask.RunTask()
		expectErr := errors.New("set temp dir right failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("import cert failed", func() {
		if err := fileutils.CreateDir(tempPath, constants.Mode700); err != nil {
			hwlog.RunLog.Errorf("create temp path failed, error: %v", err)
			return
		}
		p3 := gomonkey.ApplyFuncReturn(fileutils.CopyFile, testErr)
		defer p3.Reset()
		err := importCaTask.RunTask()
		expectErr := fmt.Errorf("import [%s] cert failed, error: %s", certName, testErr.Error())
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set temp crt right failed", func() {
		if err := fileutils.CreateDir(tempPath, constants.Mode700); err != nil {
			hwlog.RunLog.Errorf("create temp path failed, error: %v", err)
			return
		}
		p4 := gomonkey.ApplyFuncReturn(fileutils.SetPathPermission, testErr)
		defer p4.Reset()
		err := importCaTask.RunTask()
		expectErr := errors.New("set temp crt right failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}

func copyCaToEdgeMainFailed() {
	p := gomonkey.ApplyPrivateMethod(&ImportCaTask{}, "copyCaToTemp", func() error { return nil })
	defer p.Reset()

	convey.Convey("copy temp crt to dst failed", func() {
		p1 := gomonkey.ApplyFuncReturn(envutils.RunCommandWithUser, "", testErr)
		defer p1.Reset()
		err := importCaTask.RunTask()
		expectErr := errors.New("copy temp crt to dst failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})

	convey.Convey("set save crt right failed", func() {
		p2 := gomonkey.ApplyFuncSeq(envutils.RunCommandWithUser, []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", nil}},
			{Values: gomonkey.Params{"", testErr}},
		}).
			ApplyFuncReturn(fileutils.DeleteFile, testErr)
		defer p2.Reset()
		err := importCaTask.RunTask()
		expectErr := errors.New("set save crt right failed")
		convey.So(err, convey.ShouldResemble, expectErr)
	})
}
