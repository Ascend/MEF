// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package tls config test file
package tls

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
	"huawei.com/mindx/common/x509"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&config, context.Background())
}

// TestValidateX509Pair test ValidateX509Pair
func TestValidateX509Pair(t *testing.T) {
	convey.Convey("test for ValidateX509Pair", t, func() {
		convey.Convey("normal v1 cert", func() {
			certByte, err := fileutils.ReadLimitBytes("./testdata/client-v1.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			keyByte, err := fileutils.ReadLimitBytes("./testdata/client.key", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			// validate period is 10 years, after that this case maybe failed
			c, err := ValidateX509Pair(certByte, keyByte, x509.InvalidNum)
			convey.So(err, convey.ShouldNotBeEmpty)
			convey.So(c, convey.ShouldEqual, nil)
		})
		convey.Convey("normal cert", func() {
			certByte, err := fileutils.ReadLimitBytes("./testdata/client-v3.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			keyByte, err := fileutils.ReadLimitBytes("./testdata/client.key", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			// validate period is 10 years, after that this case maybe failed
			c, err := ValidateX509Pair(certByte, keyByte, x509.InvalidNum)
			convey.So(err.Error(), convey.ShouldEqual, "the certificate overdue ")
			convey.So(c, convey.ShouldNotBeEmpty)
		})
		convey.Convey("not match cert", func() {
			certByte, err := fileutils.ReadLimitBytes("./testdata/server.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			keyByte, err := fileutils.ReadLimitBytes("./testdata/client.key", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			c, err := ValidateX509Pair(certByte, keyByte, x509.InvalidNum)
			convey.So(err, convey.ShouldNotBeEmpty)
			convey.So(c, convey.ShouldEqual, nil)
		})
	})
}

// TestLoadEncryptedCertPair  test load function
func TestLoadEncryptedCertPair(t *testing.T) {
	convey.Convey("test for LoadCertPair", t, func() {
		// mock kmcInit
		initStub := gomonkey.ApplyFunc(kmc.Initialize, func(sdpAlgID int, primaryKey, standbyKey string) error {
			return nil
		})
		defer initStub.Reset()
		convey.Convey("normal cert", func() {
			certByte, err := fileutils.ReadLimitBytes("./testdata/client-v3.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			keyByte, err := fileutils.ReadLimitBytes("./testdata/client.key", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			loadStub := gomonkey.ApplyFunc(x509.LoadCertPairByte, func(pathMap map[string]string,
				encryptAlgorithm int, mode os.FileMode) ([]byte, []byte, error) {
				return certByte, keyByte, nil
			})
			defer loadStub.Reset()
			c, err := LoadCertPair(mockMap(), 0)
			convey.So(c, convey.ShouldNotBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "the certificate overdue ")
		})
		convey.Convey("cert not match", func() {
			certByte, err := fileutils.ReadLimitBytes("./testdata/server.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			keyByte, err := fileutils.ReadLimitBytes("./testdata/client.key", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			loadStub := gomonkey.ApplyFunc(x509.LoadCertPairByte, func(pathMap map[string]string,
				encryptAlgorithm int, mode os.FileMode) ([]byte, []byte, error) {
				return certByte, keyByte, nil
			})
			defer loadStub.Reset()
			c, err := LoadCertPair(mockMap(), 0)
			convey.So(c, convey.ShouldEqual, nil)
			convey.So(err, convey.ShouldNotBeEmpty)
		})
		convey.Convey("cert not exist", func() {
			loadStub := gomonkey.ApplyFunc(x509.LoadCertPairByte, func(pathMap map[string]string,
				encryptAlgorithm int, mode os.FileMode) ([]byte, []byte, error) {
				return nil, nil, errors.New("mock err")
			})
			defer loadStub.Reset()
			c, err := LoadCertPair(mockMap(), 0)
			convey.So(c, convey.ShouldEqual, nil)
			convey.So(err, convey.ShouldNotBeEmpty)
		})
	})
}

func mockMap() map[string]string {
	return map[string]string{
		x509.CertStorePath:       "./testdata/client-v3.crt",
		x509.CertStoreBackupPath: "./testdata/backup/client-v3.crt",
		x509.KeyStorePath:        "./testdata/client.key",
		x509.KeyStoreBackupPath:  "./testdata/client.key",
		x509.PassFilePath:        "./testdata/mainks",
		x509.PassFileBackUpPath:  "./testdata/mainks",
	}
}
