// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certmgr this file for cert manager
package x509

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
)

const (
	certFileName        = "./root_cert.pem"
	certKeyFileName     = "./private_key.pem"
	moreThanOneDayHours = 25
	lessThanOneDayHours = 23
	yearHours           = 24 * 365
)

type argsVerify struct {
	KeyPath     string
	SvcCertData []byte
	SvcCert     x509.Certificate
	CommonName  string
	OverdueTime int
}
type expectedVerify struct {
	checkExtension            error
	verifyCertWithKey         error
	checkOverdueTime          error
	checkSignAlgoAndKeyLength error
}
type testsVerify struct {
	name     string
	args     argsVerify
	Expected expectedVerify
}

func initLog() {
	logConfig := &hwlog.LogConfig{OnlyToStdout: true}
	if err := hwlog.InitHwLogger(logConfig, logConfig); err != nil {
		hwlog.RunLog.Errorf("init hwlog failed, %v", err)
	}
}

func CreatCertFile(t *testing.T) (*os.File, *os.File, *os.File) {
	certContent := []byte(`
-----BEGIN CERTIFICATE-----
MIIFATCCA2mgAwIBAgIUYHT2cpot/pRJDkTuJO6SRqreriQwDQYJKoZIhvcNAQEL
BQAwejELMAkGA1UEBhMCQ04xEjAQBgNVBAgMCUd1YW5nZG9uZzERMA8GA1UEBwwI
U2hlbnpoZW4xFTATBgNVBAoMDFRlc3QgQ29tcGFueTEYMBYGA1UECwwPVGVzdCBE
ZXBhcnRtZW50MRMwEQYDVQQDDAprdWJlcm5ldGVzMB4XDTI0MDYxNzA2MTU0M1oX
DTQ0MDYxMjA2MTU0M1owejELMAkGA1UEBhMCQ04xEjAQBgNVBAgMCUd1YW5nZG9u
ZzERMA8GA1UEBwwIU2hlbnpoZW4xFTATBgNVBAoMDFRlc3QgQ29tcGFueTEYMBYG
A1UECwwPVGVzdCBEZXBhcnRtZW50MRMwEQYDVQQDDAprdWJlcm5ldGVzMIIBojAN
BgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAwGaWjdtiuJlBwzchD5VBHgOJTLn0
W/G3ieMkFoJQK7yot+KzGdtQxEWu5hF6C09sitxqiYgXZKDN0NWwHGMagRcX2HnG
YqIa3dEnBIN7nXJ/8Znmb0keUReNF8VPCllVz6OQ3GK6hKrk1qiOqtXrs4+TfaIS
2JBjuq2GBjYDVse/yePr5TCIKaQ2J4/yyr/7+nKIs0Iyt1QDJd0ZONiOsmRgopxT
s+d5e3Sg43fvx8uFCddIBtXB5gDDCZPOL/RRbGlse5+rrMjlkFXgCbnK0Gr/dFTa
QBlbwP61RVS/6sZ6x1Sp4z4EKBuxSOU3gJqsG2VGqWCbyBb+9BvUndlpORsEiSxV
z3NbXUqmVbwBye7gtx9C5MpslEVHLe2gUQGBBjx8ZkPHyeSdJ0YuB4VgNRLrb9no
nRBKaWjBe1Rt6ze7lT52kyQzWvN/P4mhlY+5jibdsC7Q2bkXxoH/HsZL6PVP+88g
CASETxGhmOYxZtfJljM8XL0hR4J97kbeTDsxAgMBAAGjfzB9MB0GA1UdDgQWBBRL
Jc7zuKq+zNdeCQl8xcxfSgmO4jAfBgNVHSMEGDAWgBRLJc7zuKq+zNdeCQl8xcxf
SgmO4jAPBgNVHRMBAf8EBTADAQH/MAsGA1UdDwQEAwIBhjAdBgNVHSUEFjAUBggr
BgEFBQcDAQYIKwYBBQUHAwIwDQYJKoZIhvcNAQELBQADggGBAFkDzjrWmvn5O96I
Qc9WZtcEa4CvlKRd67ozUlmQnU3WOUAGN06Zd8fcQJ6JlUprGu5JtNqEUQexUUyb
qB/2Om7izlMwWGLXETBLlR3RXarU2K6CsEsXG90XGuAC8u4QB9Mtc9yK7WErvQ6r
LSfAb/Kq/6+9RjKN9Fo5fGHQuRK7QvbBisUh66KU7J8t52KMXqmYXo1yuUSGEk4s
L5pdOXfTZ/wrjJjz+NiWwhxh9MWiWTLEp8PrW8S61zrF0mAkV4gFhsO30TUgD9N8
n4WcuEY9cYILTHuRdqya8tPPxcmLlaiKmMSLbBgJekFOPR3U/T2ldbkr91XTFimN
3OfvEAkn2NAoDEaTkHVbPCjlxBQkDszDVgoD1t11/eO3LTMRNwLuizTxzTYR84Mh
7Ult0VdNOykbnHO5VlJd8c2bVm3r50fALXspd0EsoMNPJxaLc53QRc0nrOKG8mqz
GcACLYiQcff3gmD5++AHxMh3CBoq0NyCXKVZU81F1Z3Ny7mUyw==
-----END CERTIFICATE-----
`)
	tmpfile, err := os.CreateTemp("", "tmpCert")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write(certContent); err != nil {
		t.Fatal(err)
	}
	cert, err := os.CreateTemp("", "cert")
	if err != nil {
		t.Fatal(err)
	}
	certBak, err := os.CreateTemp("", "certbak")
	if err != nil {
		t.Fatal(err)
	}
	return tmpfile, cert, certBak
}

func TestCheckSvcCertTask(t *testing.T) {
	initLog()
	// 准备证书和秘钥
	certfile, cert, certbak := CreatCertFile(t)
	keyfile := CreatKeyFile(t)

	files := []string{
		certfile.Name(),
		keyfile.Name(),
		cert.Name(),
		certbak.Name(),
	}
	defer func() {
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	tests := preparingDataVerify(t, certfile, keyfile)

	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			cct := NewCheckSvcCertTaskForTest(tt.args.KeyPath, tt.args.SvcCertData)
			err := cct.initTask()
			convey.So(err, convey.ShouldBeNil)
			convey.Convey("TestcheckExtension", func() {
				convey.So(cct.checkExtension(), convey.ShouldNotResemble, tt.Expected.checkExtension)
			})
			convey.Convey("TestverifyCertWithKey", func() {
				patches1 := gomonkey.ApplyFunc(kmc.DecryptContent, MockDecryptContent)
				defer patches1.Reset()
				convey.So(cct.verifyCertWithKey(), convey.ShouldResemble, tt.Expected.verifyCertWithKey)
			})
			convey.Convey("TestcheckOverdueTime", func() {
				convey.So(cct.checkOverdueTime(), convey.ShouldResemble, tt.Expected.checkOverdueTime)
			})
			convey.Convey("TestcheckSignAlgoAndKeyLength", func() {
				convey.So(cct.checkSignAlgoAndKeyLength(), convey.ShouldResemble, tt.Expected.checkSignAlgoAndKeyLength)
			})
		})
	}
	convey.Convey("Test cert is empty", t, func() {
		cct := NewCheckSvcCertTaskForTest("", nil)
		err := cct.initTask()
		convey.So(err.Error(), convey.ShouldContainSubstring, "service cert content is empty")
	})
}

func preparingDataVerify(t *testing.T, certfile *os.File, keyfile *os.File) []testsVerify {

	// 获取证书内容
	certBytes, err := fileutils.LoadFile(certfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	// 解析证书
	block, _ := pem.Decode(certBytes)
	if block == nil {
		t.Fatal("failed to parse certificate")
	}
	certContent, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	tests := []testsVerify{{
		name: "Case1 : Normal",
		args: argsVerify{
			KeyPath:     keyfile.Name(),
			SvcCertData: certBytes,
			SvcCert:     *certContent,
			CommonName:  " ",
			OverdueTime: 1,
		},
		Expected: expectedVerify{
			checkExtension:            nil,
			verifyCertWithKey:         nil,
			checkOverdueTime:          nil,
			checkSignAlgoAndKeyLength: nil,
		},
	}}
	return tests
}

func NewCheckSvcCertTaskForTest(KeyPath string, SvcCertData []byte) *CheckSvcCertTask {
	return &CheckSvcCertTask{
		KeyPath:     KeyPath,
		SvcCertData: SvcCertData,
		KmcConfig:   nil,
	}
}

func CreatKeyFile(t *testing.T) *os.File {
	keyContent := GetKeyContent()
	tmpfile, err := os.CreateTemp("", "tmpKey")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write(keyContent); err != nil {
		t.Fatal(err)
	}
	return tmpfile
}

func GetKeyContent() []byte {
	keyContent := []byte(`
-----BEGIN PRIVATE KEY-----
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
KnT3PuKGAKSzBuHD4LCjTxqU
-----END PRIVATE KEY-----
`)
	return keyContent
}

func MockDecryptContent(encryptByte []byte, kmcCfg *kmc.SubConfig) ([]byte, error) {
	kmcCfg = nil
	return encryptByte, nil
}

func TestGetValidityPeriod(t *testing.T) {
	clearExistingCertFiles()
	defer clearExistingCertFiles()
	convey.Convey("GetValidityPeriod", t, func() {
		// add 25h, makes 'now' before 25h
		err := prepareNewCert(time.Now().Add(moreThanOneDayHours * time.Hour))
		certFile, err := os.OpenFile(certFileName, os.O_RDONLY, fileutils.Mode400)
		certBytes := make([]byte, 2048)
		_, err = certFile.Read(certBytes)
		convey.So(err, convey.ShouldBeNil)
		block, _ := pem.Decode(certBytes)
		svcCert, err := x509.ParseCertificate(block.Bytes)
		fmt.Println(svcCert.NotBefore, time.Now())
		convey.So(err, convey.ShouldBeNil)
		_, err = GetValidityPeriod(svcCert, true)
		convey.So(err, convey.ShouldNotBeNil)

		err = prepareNewCert(time.Now().Add(lessThanOneDayHours * time.Hour))
		certFile, err = os.OpenFile(certFileName, os.O_RDONLY, fileutils.Mode400)
		certBytes = make([]byte, 2048)
		_, err = certFile.Read(certBytes)
		convey.So(err, convey.ShouldBeNil)
		block, _ = pem.Decode(certBytes)
		svcCert, err = x509.ParseCertificate(block.Bytes)
		fmt.Println(svcCert.NotBefore, time.Now())
		convey.So(err, convey.ShouldBeNil)
		_, err = GetValidityPeriod(svcCert, true)
		convey.So(err, convey.ShouldBeNil)

		err = prepareNewCert(time.Now().Add(lessThanOneDayHours * time.Hour))
		certFile, err = os.OpenFile(certFileName, os.O_RDONLY, fileutils.Mode400)
		certBytes = make([]byte, 2048)
		_, err = certFile.Read(certBytes)
		convey.So(err, convey.ShouldBeNil)
		block, _ = pem.Decode(certBytes)
		svcCert, err = x509.ParseCertificate(block.Bytes)
		fmt.Println(svcCert.NotBefore, time.Now())
		convey.So(err, convey.ShouldBeNil)
		_, err = GetValidityPeriod(svcCert, false)
		convey.So(err, convey.ShouldNotBeNil)
	})
}

func clearExistingCertFiles() {
	files := []string{certFileName, certKeyFileName}
	for _, file := range files {
		err := fileutils.DeleteFile(file)
		if err != nil {
			hwlog.RunLog.Errorf("Error deleting, err: %s", err.Error())
		}
	}
	return
}

// prepareNewCert is to obtain new cert which notBefore is before 'time.Now'
func prepareNewCert(notBeforeTime time.Time) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return err
	}
	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"MEFTest"}},
		NotBefore:             notBeforeTime,
		NotAfter:              time.Now().Add(yearHours * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	certFile, err := os.OpenFile(certFileName, os.O_WRONLY|os.O_CREATE, fileutils.Mode600)
	defer certFile.Close()
	if err != nil {
		return err
	}
	if err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		hwlog.RunLog.Error(err)
		return err
	}
	keyFile, err := os.OpenFile(certKeyFileName, os.O_WRONLY|os.O_CREATE, fileutils.Mode600)
	defer keyFile.Close()
	if err != nil {
		return err
	}
	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}
	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	if err != nil {
		return err
	}
	return nil
}
