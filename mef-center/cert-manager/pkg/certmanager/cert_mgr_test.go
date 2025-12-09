// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package certmanager

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gin-gonic/gin"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509/certutils"

	"cert-manager/pkg/certmanager/certchecker"
	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

func TestIsCertImported(t *testing.T) {
	convey.Convey("case: cert imported", t, func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.IsExist, true)
		defer patch.Reset()
		imported := isCertImported(common.ImageCertName)
		convey.So(imported, convey.ShouldBeTrue)
	})
}

func TestIsExternalCrlImported(t *testing.T) {
	convey.Convey("case: crl imported", t, func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.IsExist, true)
		defer patch.Reset()
		imported := isExternalCrlImported(common.ImageCertName)
		convey.So(imported, convey.ShouldBeTrue)
	})
}

func TestGetCertByCertName(t *testing.T) {
	convey.Convey("case: normal success", t, func() {
		patch := gomonkey.ApplyFuncReturn(certutils.GetCertContentWithBackup, []byte(testContent), nil)
		defer patch.Reset()
		data, err := getCertByCertName(common.ImageCertName)
		convey.So(data, convey.ShouldResemble, []byte(testContent))
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("case: get cert failed", t, func() {
		patch := gomonkey.ApplyFuncReturn(certutils.GetCertContentWithBackup, []byte(testContent), test.ErrTest)
		defer patch.Reset()
		_, err := getCertByCertName(common.ImageCertName)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("load root cert failed: %v", test.ErrTest))
	})
}

func TestCreateCaIfNotExit(t *testing.T) {
	convey.Convey("case: ca filed exist", t, func() {
		patch := gomonkey.ApplyFuncReturn(isRootCaFilesExist, true)
		defer patch.Reset()
		convey.So(CreateCaIfNotExit(common.ImageCertName), convey.ShouldBeNil)
	})

	convey.Convey("case: create ca filed succeed", t, func() {
		patches := gomonkey.ApplyFuncReturn(isRootCaFilesExist, false).
			ApplyPrivateMethod(&certutils.RootCertMgr{}, "NewRootCaWithBackup",
				func(*certutils.RootCertMgr) (*certutils.CaPairInfo, error) {
					return nil, nil
				})
		defer patches.Reset()
		convey.So(CreateCaIfNotExit(common.ImageCertName), convey.ShouldBeNil)
	})
}

func TestCreateTempCaCert(t *testing.T) {
	convey.Convey("case: previously created certs exist", t, func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(fileutils.LoadFile, []byte(testContent), nil)
		defer patch.Reset()
		cert, err := CreateTempCaCert(common.ImageCertName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(cert, convey.ShouldResemble, testContent)
	})

	convey.Convey("case: create ca filed succeed", t, func() {
		patches := gomonkey.ApplyFuncReturn(isRootCaFilesExist, false).
			ApplyPrivateMethod(&certutils.RootCertMgr{}, "NewRootCa",
				func(*certutils.RootCertMgr) (*certutils.CaPairInfo, error) {
					return nil, nil
				}).
			ApplyFuncReturn(fileutils.LoadFile, []byte(testContent), nil)
		defer patches.Reset()
		caStr, err := CreateTempCaCert(common.ImageCertName)
		convey.So(err, convey.ShouldBeNil)
		convey.So(caStr, convey.ShouldResemble, testContent)
	})

	convey.Convey("case: new ca failed", t, func() {
		patches := gomonkey.ApplyFuncReturn(isRootCaFilesExist, false).
			ApplyPrivateMethod(&certutils.RootCertMgr{}, "NewRootCa",
				func(*certutils.RootCertMgr) (*certutils.CaPairInfo, error) {
					return nil, test.ErrTest
				})
		defer patches.Reset()
		_, err := CreateTempCaCert(common.ImageCertName)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("create new root ca [%v] failed: %v", common.ImageCertName, test.ErrTest))
	})

	convey.Convey("case: load temp ca failed", t, func() {
		patches := gomonkey.ApplyFuncReturn(isRootCaFilesExist, false).
			ApplyPrivateMethod(&certutils.RootCertMgr{}, "NewRootCa",
				func(*certutils.RootCertMgr) (*certutils.CaPairInfo, error) {
					return nil, nil
				}).
			ApplyFuncReturn(fileutils.LoadFile, nil, test.ErrTest)
		defer patches.Reset()
		_, err := CreateTempCaCert(common.ImageCertName)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("load new temp root ca failed: %v", test.ErrTest))
	})
}

func TestUpdateCaCertWithTemp(t *testing.T) {
	convey.Convey("case: update ca cert succeed", t, func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.SetPathPermission, nil).
			ApplyFuncReturn(RemoveTempCaCert, nil).
			ApplyFuncReturn(fileutils.CopyFile, nil).
			ApplyFuncReturn(backuputils.BackUpFiles, nil)
		defer patches.Reset()
		convey.So(UpdateCaCertWithTemp(common.ImageCertName), convey.ShouldBeNil)
	})
}

func TestRemoveTempCaCert(t *testing.T) {
	convey.Convey("case: normal success", t, func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, nil)
		defer patches.Reset()
		convey.So(RemoveTempCaCert(common.ImageCertName), convey.ShouldBeNil)
	})

	convey.Convey("case: temp ca cert not exist", t, func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.IsExist, false)
		defer patch.Reset()
		convey.So(RemoveTempCaCert(common.ImageCertName), convey.ShouldBeNil)
	})

	convey.Convey("case: remove temp ca cert failed", t, func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(fileutils.DeleteAllFileWithConfusion, test.ErrTest)
		defer patches.Reset()
		convey.So(RemoveTempCaCert(common.ImageCertName), convey.ShouldResemble, test.ErrTest)
	})
}

func TestIssueServiceCert(t *testing.T) {
	convey.Convey("case: normal success", t, func() {
		patches := gomonkey.ApplyPrivateMethod(base64.StdEncoding, "DecodeString",
			func(*base64.Encoding, string) ([]byte, error) {
				return []byte(testContent), nil
			}).
			ApplyFuncReturn(pem.Decode, &pem.Block{}, nil).
			ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(isRootCaFilesExist, false).
			ApplyPrivateMethod(&certutils.RootCertMgr{}, "NewRootCaWithBackup",
				func(*certutils.RootCertMgr) (*certutils.CaPairInfo, error) {
					return nil, nil
				}).
			ApplyPrivateMethod(&certutils.RootCertMgr{}, "IssueServiceCertWithBackup",
				func(*certutils.RootCertMgr, []byte) ([]byte, error) {
					return nil, nil
				}).
			ApplyFuncReturn(certutils.PemWrapCert, []byte(testContent))
		defer patches.Reset()
		cert, err := issueServiceCert(common.WsCltName, testContent)
		convey.So(err, convey.ShouldBeNil)
		convey.So(cert, convey.ShouldResemble, []byte(testContent))
	})
}

func TestIsRootCaFilesExist(t *testing.T) {
	convey.Convey("case: root ca exist", t, func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.IsLexist, true)
		defer patch.Reset()
		convey.So(isRootCaFilesExist("", ""), convey.ShouldBeTrue)
	})

	convey.Convey("case: root ca not exist", t, func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.IsLexist, false)
		defer patch.Reset()
		convey.So(isRootCaFilesExist("", ""), convey.ShouldBeFalse)
	})
}

func TestSaveCaContent(t *testing.T) {
	convey.Convey("succeed cases", t, testSaveCaContentSucceedCases)
	convey.Convey("failed cases", t, testSaveCaContentFailedCases)
}

func testSaveCaContentSucceedCases() {
	convey.Convey("case: normal success with old cert saved", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.MakeSureDir, nil).
			ApplyFuncReturn(certutils.GetCertContentWithBackup, []byte(testContent), nil).
			ApplyFuncReturn(fileutils.WriteData, nil).
			ApplyFuncReturn(backuputils.BackUpFiles, test.ErrTest)
		defer patches.Reset()
		convey.So(saveCaContent(common.ImageCertName, []byte{}), convey.ShouldBeNil)
	})

	convey.Convey("case: success with old cert not exist", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.MakeSureDir, nil).
			ApplyFuncReturn(certutils.GetCertContentWithBackup, nil, test.ErrTest).
			ApplyFuncReturn(fileutils.WriteData, nil).
			ApplyFuncReturn(backuputils.BackUpFiles, nil)
		defer patches.Reset()
		convey.So(saveCaContent(common.ImageCertName, []byte{}), convey.ShouldBeNil)
	})
}

func testSaveCaContentFailedCases() {
	convey.Convey("case: prepare dir failed", func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.MakeSureDir, test.ErrTest)
		defer patch.Reset()
		convey.So(saveCaContent(common.ImageCertName, []byte{}), convey.ShouldResemble,
			fmt.Errorf("create %s ca folder failed, error: %v", common.ImageCertName, test.ErrTest))
	})

	convey.Convey("case: save old cert failed", func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.MakeSureDir, nil).
			ApplyFuncReturn(certutils.GetCertContentWithBackup, []byte(testContent), nil).
			ApplyFuncReturn(fileutils.WriteData, test.ErrTest)
		defer patch.Reset()
		convey.So(saveCaContent(common.ImageCertName, []byte{}), convey.ShouldResemble,
			fmt.Errorf("write previous ca of %s failed, error: %v", common.ImageCertName, test.ErrTest))
	})

	convey.Convey("case: save cert failed", func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.MakeSureDir, nil).
			ApplyFuncReturn(certutils.GetCertContentWithBackup, nil, test.ErrTest).
			ApplyFuncReturn(fileutils.WriteData, test.ErrTest)
		defer patch.Reset()
		convey.So(saveCaContent(common.ImageCertName, []byte{}), convey.ShouldResemble,
			fmt.Errorf("save %s ca file failed", common.ImageCertName))
	})
}

func TestRemoveCaFile(t *testing.T) {
	convey.Convey("case: normal success", t, func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, nil)
		defer patch.Reset()
		convey.So(removeCaFile(common.ImageCertName), convey.ShouldBeNil)
	})

	convey.Convey("case: ca file not exist", t, func() {
		patch := gomonkey.ApplyFuncReturn(fileutils.DeleteFile, test.ErrTest)
		defer patch.Reset()
		convey.So(removeCaFile(common.ImageCertName), convey.ShouldResemble,
			fmt.Errorf("remove %s ca file failed, error: %v", common.ImageCertName, test.ErrTest))
	})
}

func TestUpdateClientCert(t *testing.T) {
	convey.Convey("case: normal success", t, func() {
		patches := gomonkey.ApplyFuncReturn(getCertByCertName, []byte{}, nil).
			ApplyFuncReturn(isExternalCrlImported, true).
			ApplyFuncReturn(certutils.GetCrlContentWithBackup, []byte{}, nil).
			ApplyPrivateMethod(&requests.ReqCertParams{}, "UpdateCertFile",
				func(*requests.ReqCertParams, certutils.UpdateClientCert) (string, error) {
					return "", nil
				})
		defer patches.Reset()
		convey.So(updateClientCert(common.ImageCertName, common.Update), convey.ShouldBeNil)
	})

	convey.Convey("case: get cert by name failed", t, func() {
		patch := gomonkey.ApplyFuncReturn(getCertByCertName, nil, test.ErrTest)
		defer patch.Reset()
		convey.So(updateClientCert(common.ImageCertName, common.Update), convey.ShouldResemble,
			fmt.Errorf("load %s ca file failed", common.ImageCertName))
	})

	convey.Convey("case: get crl failed", t, func() {
		patches := gomonkey.ApplyFuncReturn(getCertByCertName, []byte{}, nil).
			ApplyFuncReturn(isExternalCrlImported, true).
			ApplyFuncReturn(certutils.GetCrlContentWithBackup, nil, test.ErrTest)
		defer patches.Reset()
		convey.So(updateClientCert(common.ImageCertName, common.Update), convey.ShouldResemble,
			fmt.Errorf("load %s crl file failed", common.ImageCertName))
	})

	convey.Convey("case: update cert file failed", t, func() {
		patch := gomonkey.ApplyPrivateMethod(&requests.ReqCertParams{}, "UpdateCertFile",
			func(*requests.ReqCertParams, certutils.UpdateClientCert) (string, error) {
				return common.ImageCertName, test.ErrTest
			})
		defer patch.Reset()
		convey.So(updateClientCert(common.ImageCertName, common.Delete), convey.ShouldResemble,
			fmt.Errorf("update %s ca file failed", common.ImageCertName))
	})
}

func TestExportRootCa(t *testing.T) {
	convey.Convey("case: normal success", t, func() {
		var invokeFlag bool
		patches := gomonkey.ApplyFuncReturn(getCertByCertName, []byte{}, nil).
			ApplyPrivateMethod(&gin.Context{}, "JSON", func(_ *gin.Context, code int, obj any) {
				invokeFlag = true
			}).
			ApplyFuncReturn(certchecker.CheckIfCanExport, true).
			ApplyFuncReturn(fileutils.IsExist, true)
		defer patches.Reset()
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		ExportRootCa(c)
		convey.So(invokeFlag, convey.ShouldBeFalse)
	})
}
