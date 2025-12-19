// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package x509 certificate test file
package x509

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/kmc"
)

func init() {
	config := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&config, context.Background())
}

// TestCertStatus test for checkCertStatus
func TestCertStatus(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2022-03-18T00:00:00Z")
	if err != nil {
		fmt.Printf("Parse time failed %#v\n", err)
	}
	t2, err := time.Parse(time.RFC3339, "2022-03-20T00:00:00Z")
	if err != nil {
		fmt.Printf("Parse time failed %#v\n", err)
	}
	cs := &CertStatus{
		NotBefore: t1,
		NotAfter:  t2,
		IsCA:      true,
	}
	convey.Convey("overdue 1day", t, func() {
		x, err := time.Parse(time.RFC3339, "2022-03-19T00:00:00Z")
		if err != nil {
			fmt.Printf("Parse time failed %#v\n", err)
		}
		checkCertStatus(x, cs)
	})
}

// TestCheckCaCert test for CheckCaCert
func TestCheckCaCert(t *testing.T) {
	convey.Convey("test for CheckCaCert", t, func() {
		convey.Convey("normal situation,no err returned", func() {
			_, err := CheckCaCert("./testdata/ca.crt", InvalidNum)
			convey.So(err, convey.ShouldEqual, nil)
		})
		convey.Convey("cert is nil", func() {
			_, err := CheckCaCert("", InvalidNum)
			convey.So(err, convey.ShouldEqual, nil)
		})
		convey.Convey("cert file is not exsit", func() {
			_, err := CheckCaCert("/djdsk.../dsd", InvalidNum)
			convey.So(err, convey.ShouldNotBeEmpty)
		})
		convey.Convey("ca file not right", func() {
			_, err := CheckCaCert("./testdata/ca_err.crt", InvalidNum)
			convey.So(err, convey.ShouldNotBeEmpty)
		})
		convey.Convey("cert file is not ca", func() {
			_, err := CheckCaCert("./testdata/server.crt", InvalidNum)
			convey.So(err, convey.ShouldNotBeEmpty)
		})
	})
}

// TestLoadCertsFromPEM test LoadCertsFromPEM
func TestLoadCertsFromPEM(t *testing.T) {
	convey.Convey("test for DecryptPrivateKey", t, func() {
		convey.Convey("normal cert", func() {
			caByte, err := fileutils.ReadLimitBytes("./testdata/ca.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			ca, err := LoadCertsFromPEM(caByte)
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(ca.IsCA, convey.ShouldBeTrue)
		})
		convey.Convey("abnormal cert", func() {
			caByte, err := fileutils.ReadLimitBytes("./testdata/ca_err.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			ca, err := LoadCertsFromPEM(caByte)
			convey.So(ca, convey.ShouldEqual, nil)
			convey.So(err, convey.ShouldNotBeEmpty)
		})
	})
}

// TestGetPrivateKeyLength
func TestCheckValidityPeriodWithError(t *testing.T) {
	convey.Convey("normal", t, func() {
		now := time.Now()
		cert := &x509.Certificate{
			NotBefore: now,
			NotAfter:  now.AddDate(1, 0, 0),
		}
		err := CheckValidityPeriodWithError(cert, 1)
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("need update", t, func() {
		now := time.Now()
		cert := &x509.Certificate{
			NotBefore: now,
			NotAfter:  now.AddDate(0, 0, 1),
		}
		err := CheckValidityPeriodWithError(cert, 1)
		convey.So(err.Error(), convey.ShouldContainSubstring, "need to update certification")
	})
	convey.Convey("overdue", t, func() {
		now := time.Now()
		cert := &x509.Certificate{
			NotBefore: now,
			NotAfter:  now.AddDate(0, 0, -1),
		}
		err := CheckValidityPeriodWithError(cert, 1)
		convey.So(err.Error(), convey.ShouldContainSubstring, "the certificate overdue")
	})
}

// TestCheckSignatureAlgorithm test CheckSignatureAlgorithm
func TestCheckSignatureAlgorithm(t *testing.T) {
	convey.Convey("test for CheckSignatureAlgorithm", t, func() {
		convey.Convey("normal cert", func() {
			caByte, err := fileutils.ReadLimitBytes("./testdata/ca.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			ca, err := LoadCertsFromPEM(caByte)
			err = CheckSignatureAlgorithm(ca)
			convey.So(err, convey.ShouldEqual, nil)
		})
	})
}

func createGetPrivateKeyLengthTestData(curve elliptic.Curve) (*x509.Certificate, *tls.Certificate) {
	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Printf("create ecdsa private key failed: %#v\n", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"This is Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Printf("Failed to create certificate: %s\n", err)
	}
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		fmt.Printf("Parse certificate failed: %s\n", err)
	}
	c := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		fmt.Printf("x509.MarshalECPrivateKey failed: %s\n", err)
	}
	k := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keypair, err := tls.X509KeyPair(c, k)
	if err != nil {
		fmt.Printf("tls.X509KeyPair failed: %s\n", err)
	}
	return cert, &keypair
}

func createRSAPrivateKeyLengthTestData(bits int) (*x509.Certificate, *tls.Certificate) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		fmt.Printf("create ecdsa private key failed: %#v\n", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"This is Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Printf("Failed to create certificate: %s\n", err)
	}
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		fmt.Printf("Parse certificate failed: %s\n", err)
	}
	c := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	b := x509.MarshalPKCS1PrivateKey(priv)
	if err != nil {
		fmt.Printf("x509.MarshalECPrivateKey failed: %s\n", err)
	}
	k := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keypair, err := tls.X509KeyPair(c, k)
	if err != nil {
		fmt.Printf("tls.X509KeyPair failed: %s\n", err)
	}
	return cert, &keypair
}

func createED25519PrivateKeyLengthTestData() (*x509.Certificate, *tls.Certificate) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("create ecdsa private key failed: %#v\n", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"This is Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	if err != nil {
		fmt.Printf("Failed to create certificate: %s\n", err)
	}
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		fmt.Printf("Parse certificate failed: %s\n", err)
	}
	c := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		fmt.Printf("x509.MarshalECPrivateKey failed: %s\n", err)
	}
	k := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	keypair, err := tls.X509KeyPair(c, k)
	if err != nil {
		fmt.Printf("tls.X509KeyPair failed: %s\n", err)
	}
	return cert, &keypair
}

// TestGetPrivateKeyLength
func TestGetPrivateKeyLength(t *testing.T) {
	convey.Convey("get key length of Curve 384", t, func() {
		// P384 curve key length is 388
		const bitLengthP384 = 384
		cert, keypair := createGetPrivateKeyLengthTestData(elliptic.P384())
		keyLen, keyType, err := GetPrivateKeyLength(cert, keypair)
		if err != nil {
			fmt.Printf("GetPrivateKeyLength failed %#v\n", err)
		}
		fmt.Printf("private key length is %#v, key type is %#v\n", keyLen, keyType)
		convey.So(keyLen, convey.ShouldEqual, bitLengthP384)
	})

	convey.Convey("get key length of Curve 256", t, func() {
		// P521 curve key length is 256. the byte lengh is in 256
		const bitLengthP256 = 256
		cert, keypair := createGetPrivateKeyLengthTestData(elliptic.P256())
		keyLen, keyType, err := GetPrivateKeyLength(cert, keypair)
		if err != nil {
			fmt.Printf("GetPrivateKeyLength failed %#v\n", err)
		}
		fmt.Printf("private key length is %#v, key type is %#v\n", keyLen, keyType)
		convey.So(keyLen, convey.ShouldEqual, bitLengthP256)
	})

	convey.Convey("given cert nil return error", t, func() {
		const empty = 0
		cert, _ := createGetPrivateKeyLengthTestData(elliptic.P256())
		keyLen, keyType, err := GetPrivateKeyLength(cert, nil)
		convey.So(keyLen, convey.ShouldEqual, empty)
		convey.So(keyType, convey.ShouldBeEmpty)
		convey.So(err.Error(), convey.ShouldEqual, "certificate is nil")
	})

	convey.Convey("get key length of RSA 2048", t, func() {
		const rsaLength = 2048
		cert, keypair := createRSAPrivateKeyLengthTestData(rsaLength)
		keyLen, keyType, err := GetPrivateKeyLength(cert, keypair)
		convey.So(keyLen, convey.ShouldEqual, rsaLength)
		convey.So(keyType, convey.ShouldEqual, "RSA")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("get key length of ED25519", t, func() {
		const ED25519Length = 32
		cert, keypair := createED25519PrivateKeyLengthTestData()
		keyLen, keyType, err := GetPrivateKeyLength(cert, keypair)
		convey.So(keyLen, convey.ShouldEqual, ED25519Length)
		convey.So(keyType, convey.ShouldEqual, "ED25519")
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestCheckRevokedCert test RevokedCert
func TestCheckRevokedCert(t *testing.T) {
	convey.Convey("test for CheckRevokedCert", t, func() {
		convey.Convey("cert revoked", func() {
			certByte, err := fileutils.ReadLimitBytes("./testdata/crl/certificate.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			cert, _ := LoadCertsFromPEM(certByte)
			cacert, err := fileutils.ReadLimitBytes("./testdata/crl/ca.crt", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			ca, _ := LoadCertsFromPEM(cacert)
			r := &http.Request{
				TLS: &tls.ConnectionState{
					VerifiedChains:   [][]*x509.Certificate{{cert, ca}},
					PeerCertificates: []*x509.Certificate{{SerialNumber: big.NewInt(1)}, cert},
				},
			}
			crlByte, err := fileutils.ReadLimitBytes("./testdata/crl/certificate_revokelist.crl", fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			crl, err := x509.ParseCRL(crlByte)
			if err == nil {
				convey.So(err, convey.ShouldEqual, nil)
			}
			res := CheckRevokedCert(r, crl)
			convey.So(res, convey.ShouldBeTrue)
		})
		convey.Convey("cert not revoked", func() {
			r := &http.Request{TLS: &tls.ConnectionState{}}
			crlcerList1 := &pkix.CertificateList{
				TBSCertList: pkix.TBSCertificateList{
					RevokedCertificates: []pkix.RevokedCertificate{{
						SerialNumber:   big.NewInt(1),
						RevocationTime: time.Time{},
						Extensions:     nil,
					}},
				},
			}
			crlcerList2 := &pkix.CertificateList{
				TBSCertList: pkix.TBSCertificateList{RevokedCertificates: []pkix.RevokedCertificate{}},
			}
			res := CheckRevokedCert(r, nil)
			convey.So(res, convey.ShouldBeFalse)
			res = CheckRevokedCert(r, crlcerList1)
			convey.So(res, convey.ShouldBeFalse)
			res = CheckRevokedCert(r, crlcerList2)
			convey.So(res, convey.ShouldBeFalse)
		})
	})
}

// TestCheckCRL test CheckCRL
func TestCheckCRL(t *testing.T) {
	convey.Convey("CheckCRL test", t, func() {
		convey.Convey("crl update time not match,return error", func() {
			_, err := CheckCRL("./testdata/client.crl")
			convey.So(err, convey.ShouldNotBeEmpty)
		})
		convey.Convey("directory no exist,no err returned", func() {
			_, err := CheckCRL("./testdata/xxx.crl")
			convey.So(err, convey.ShouldNotBeEmpty)
		})
		convey.Convey("crl file content wrong,err returned", func() {
			_, err := CheckCRL("./testdata/client_err.crl")
			convey.So(err, convey.ShouldNotBeEmpty)
		})

	})
}

// TestEncryptPrivateKeyAgain test EncryptPrivateKeyAgain
func TestEncryptPrivateKeyAgain(t *testing.T) {
	mainks := getAbsPath("./testdata/mainPd", t)
	backupks := getAbsPath("./testdata/backupPd", t)
	convey.Convey("test for EncryptPrivateKey", t, func() {
		// mock kmcInit
		initStub := gomonkey.ApplyFunc(kmc.Initialize, func(sdpAlgID int, primaryKey, standbyKey string) error {
			return nil
		})
		defer initStub.Reset()
		encryptStub := gomonkey.ApplyFunc(kmc.Encrypt, func(domainID uint, data []byte) ([]byte, error) {
			return []byte("test"), nil
		})
		defer encryptStub.Reset()
		keyBytes, err := fileutils.ReadLimitBytes("./testdata/client.key", fileutils.Size10M)
		convey.So(err, convey.ShouldEqual, nil)
		block, _ := pem.Decode(keyBytes)
		convey.Convey("read from main file", func() {
			encryptedBlock, err := EncryptPrivateKeyAgain(block, mainks, backupks, 0)
			convey.So(err, convey.ShouldEqual, nil)
			_, ok := encryptedBlock.Headers["DEK-Info"]
			convey.So(ok, convey.ShouldBeTrue)
			pd, err := fileutils.ReadLimitBytes(mainks, fileutils.Size10M)
			convey.So(err, convey.ShouldEqual, nil)
			convey.So(pd, convey.ShouldNotBeEmpty)
		})

	})

}

// TestCheckCertFilesPart1 test CheckCertFiles
func TestCheckCertFilesPart1(t *testing.T) {
	convey.Convey("CheckCertFiles test", t, func() {
		convey.Convey("one of paths is empty, return error containing empty", func() {
			pathMap1 := map[string]string{
				KeyStorePath: "keyFile",
			}
			pathMap2 := map[string]string{
				CertStorePath: "cert",
			}
			pathMap3 := map[string]string{
				CertStorePath:       "cert",
				CertStoreBackupPath: "certBkp",
			}
			pathMap4 := map[string]string{
				CertStorePath:       "cert",
				CertStoreBackupPath: "certBkp",
				KeyStorePath:        "keyFile",
			}
			pathMap5 := map[string]string{
				CertStorePath:       "cert",
				CertStoreBackupPath: "certBkp",
				KeyStorePath:        "keyFile",
				KeyStoreBackupPath:  "keyBkp",
			}
			pathMap6 := map[string]string{
				CertStorePath:       "cert",
				CertStoreBackupPath: "certBkp",
				KeyStorePath:        "keyFile",
				KeyStoreBackupPath:  "keyBkp",
				PassFilePath:        "pass",
			}
			err1 := CheckCertFiles(pathMap1)
			err2 := CheckCertFiles(pathMap2)
			err3 := CheckCertFiles(pathMap3)
			err4 := CheckCertFiles(pathMap4)
			err5 := CheckCertFiles(pathMap5)
			err6 := CheckCertFiles(pathMap6)
			convey.So(err1.Error(), convey.ShouldContainSubstring, "is empty")
			convey.So(err2.Error(), convey.ShouldContainSubstring, "is empty")
			convey.So(err3.Error(), convey.ShouldContainSubstring, "is empty")
			convey.So(err4.Error(), convey.ShouldContainSubstring, "is empty")
			convey.So(err5.Error(), convey.ShouldContainSubstring, "is empty")
			convey.So(err6.Error(), convey.ShouldContainSubstring, "is empty")

		})
	})
}

// TestCheckCertFilesPart1 test CheckCertFiles
func TestCheckCertFilesPart2(t *testing.T) {
	convey.Convey("CheckCertFiles test", t, func() {
		pathMap := map[string]string{
			CertStorePath:       "cert",
			CertStoreBackupPath: "certBkp",
			KeyStorePath:        "keyFile",
			KeyStoreBackupPath:  "keyBkp",
			PassFilePath:        "pass",
			PassFileBackUpPath:  "passbkp",
		}
		convey.Convey("ps file not exist", func() {
			pathMap[PassFilePath] = "./testdata/pass"
			pathMap[PassFileBackUpPath] = "./testdata/passbkp"
			err := CheckCertFiles(pathMap)
			convey.So(err, convey.ShouldEqual, os.ErrNotExist)
		})

		convey.Convey("keyPath file not exist", func() {
			pathMap[PassFilePath] = "./testdata/mainks"
			pathMap[PassFileBackUpPath] = "./testdata/mainks"
			err := CheckCertFiles(pathMap)
			convey.So(err, convey.ShouldEqual, os.ErrNotExist)
		})

		convey.Convey("cert file exist", func() {
			pathMap[KeyStorePath] = "./testdata/mainks"
			pathMap[KeyStoreBackupPath] = "./testdata/mainks"
			err := CheckCertFiles(pathMap)
			convey.So(err, convey.ShouldEqual, os.ErrNotExist)
		})

		convey.Convey("all file exist but checkpath failed", func() {
			stub := gomonkey.ApplyFunc(fileutils.CheckOriginPath, func(path string) (string, error) {
				return "", os.ErrNotExist
			})
			defer stub.Reset()
			err := CheckCertFiles(pathMap)
			convey.So(err, convey.ShouldEqual, os.ErrNotExist)
		})
	})
}

// TestLoadCertPairByte test LoadCertPairByte
func TestLoadCertPairByte(t *testing.T) {
	const encryptAlgorithm = 0
	convey.Convey("LoadCertPairByte test", t, func() {
		convey.Convey("CheckCertFiles failed", func() {
			pathMap := map[string]string{
				KeyStorePath:       "keyFile",
				CertStorePath:      "cert",
				PassFilePath:       "psFile",
				PassFileBackUpPath: "psFileBk",
			}

			certBytes, keyPem, err := LoadCertPairByte(pathMap, encryptAlgorithm, fileutils.Mode600)
			convey.So(certBytes, convey.ShouldBeNil)
			convey.So(keyPem, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeEmpty)
		})

		convey.Convey("normal load", func() {
			initStub := gomonkey.ApplyFunc(kmc.Initialize, func(sdpAlgID int, primaryKey, standbyKey string) error {
				return nil
			})
			defer initStub.Reset()

			decryptStub := gomonkey.ApplyFunc(kmc.Decrypt, func(domainID uint, data []byte) ([]byte, error) {
				return []byte("111111"), nil
			})
			defer decryptStub.Reset()

			isEncryptedStub := gomonkey.ApplyFunc(IsEncryptedKey, func(keyFile string) (bool, error) {
				return true, nil
			})
			defer isEncryptedStub.Reset()

			pathMap := map[string]string{
				KeyStorePath:        getAbsPath("./testdata/client.key", t),
				KeyStoreBackupPath:  getAbsPath("./testdata/client.key", t),
				CertStorePath:       getAbsPath("./testdata/client-v3.crt", t),
				CertStoreBackupPath: getAbsPath("./testdata/client-v3.crt", t),
				PassFilePath:        getAbsPath("./testdata/mainks", t),
				PassFileBackUpPath:  getAbsPath("./testdata/mainks", t),
			}
			var ins *BackUpInstance
			mock := gomonkey.ApplyMethod(reflect.TypeOf(ins), "WriteToDisk",
				func(_ *BackUpInstance, mode os.FileMode, needPadding bool) error {
					return nil
				})
			defer mock.Reset()
			certBytes, keyPem, err := LoadCertPairByte(pathMap, encryptAlgorithm, fileutils.Mode600)
			convey.So(certBytes, convey.ShouldNotBeEmpty)
			convey.So(keyPem, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})

	})
}

// TestIsEncryptedKey test IsEncryptedKey
func TestIsEncryptedKey(t *testing.T) {
	convey.Convey("test for IsEncryptedKey", t, func() {
		convey.Convey("file not exist", func() {
			ok, err := IsEncryptedKey("xxx")
			convey.So(ok, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("decode failed", func() {
			ok, err := IsEncryptedKey(getAbsPath("./testdata/mainks", t))
			convey.So(ok, convey.ShouldBeFalse)
			convey.So(err.Error(), convey.ShouldEqual, "decode key file failed")
		})

		convey.Convey("not encrypted PEM block", func() {
			ok, err := IsEncryptedKey(getAbsPath("./testdata/client.key", t))
			hwlog.RunLog.Error(err)
			convey.So(ok, convey.ShouldBeFalse)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestCheckExtension test CheckExtension
func TestCheckExtension(t *testing.T) {
	convey.Convey("test for CheckExtension", t, func() {
		convey.Convey("should return err when version is not 3", func() {
			const x509version = 1
			cert := &x509.Certificate{Version: x509version}
			err := CheckExtension(cert)
			convey.So(err.Error(), convey.ShouldEqual, "the certificate must be x509v3")
		})

		convey.Convey("should return nil when not isCA", func() {
			const x509version = 3
			cert := &x509.Certificate{Version: x509version, IsCA: false}
			err := CheckExtension(cert)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return err when not include KeyUsageCertSign", func() {
			const x509version = 3
			cert := &x509.Certificate{Version: x509version, IsCA: true, KeyUsage: x509.KeyUsageDigitalSignature}
			err := CheckExtension(cert)
			convey.So(err.Error(), convey.ShouldEqual, "CA certificate keyUsage didn't include keyCertSign")
		})
	})
}

func TestCheckCaExtension(t *testing.T) {
	convey.Convey("test for CheckCaExtension", t, func() {
		convey.Convey("should return err when version is not 3", testCheckCaExtWithVersionWrong)

		convey.Convey("should return err when not isCA", testCheckCaExtWithNotCa)

		convey.Convey("should return err when not include KeyUsageCertSign", testCheckCaExtWithNotKeySign)
	})
}

func testCheckCaExtWithVersionWrong() {
	const x509version = 1
	cert := &x509.Certificate{Version: x509version}
	err := CheckCaExtension(cert)
	convey.So(err.Error(), convey.ShouldEqual, "the certificate must be x509v3")
}

func testCheckCaExtWithNotCa() {
	const x509version = 3
	cert := &x509.Certificate{Version: x509version, IsCA: false}
	err := CheckCaExtension(cert)
	convey.So(err, convey.ShouldNotBeNil)
}

func testCheckCaExtWithNotKeySign() {
	const x509version = 3
	cert := &x509.Certificate{Version: x509version, IsCA: true, KeyUsage: x509.KeyUsageDigitalSignature}
	err := CheckCaExtension(cert)
	convey.So(err, convey.ShouldNotBeNil)
}

func TestCheckPubKeyLength(t *testing.T) {
	convey.Convey("test for CheckPubKeyLength", t, func() {
		convey.Convey("rsa cert with 3072 key length should success", testRSA3072KeyLength)

		convey.Convey("rsa cert with 2048 key length should fail", testRSA2048KeyLength)

		convey.Convey("ecdsa cert with 384 key length should success", testECDSA256KeyLength)
	})
}

func testRSA3072KeyLength() {
	const keyLength = 3072
	rsaKey, err := rsa.GenerateKey(rand.Reader, keyLength)
	convey.So(err, convey.ShouldBeNil)
	cert := &x509.Certificate{PublicKey: rsaKey.Public(), PublicKeyAlgorithm: x509.RSA}
	err = CheckPubKeyLength(cert)
	convey.So(err, convey.ShouldBeNil)
}

func testRSA2048KeyLength() {
	const keyLength = 2048
	rsaKey, err := rsa.GenerateKey(rand.Reader, keyLength)
	convey.So(err, convey.ShouldBeNil)
	cert := &x509.Certificate{PublicKey: rsaKey.Public(), PublicKeyAlgorithm: x509.RSA}
	err = CheckPubKeyLength(cert)
	convey.So(err, convey.ShouldNotBeNil)
}

func testECDSA256KeyLength() {
	curve := elliptic.P256()
	ecdsaKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	convey.So(err, convey.ShouldBeNil)
	cert := &x509.Certificate{PublicKey: ecdsaKey.Public(), PublicKeyAlgorithm: x509.ECDSA}
	err = CheckPubKeyLength(cert)
	convey.So(err, convey.ShouldBeNil)
}

func TestInterceptor(t *testing.T) {
	crlBytes, err := fileutils.LoadFile("./testdata/client.crl")
	if err != nil {
		t.Fatal(err)
	}
	crlist, err := x509.ParseCRL(crlBytes)
	if err != nil {
		t.Fatal(err)
	}
	r := &http.Request{
		URL: &url.URL{
			Path: "test.com",
		},
		Header: map[string][]string{"userID": {"1"}, "reqID": {"requestIDxxxx"}},
		Method: "GET",
	}
	h := http.DefaultServeMux
	mockSever := gomonkey.ApplyMethodFunc(h, "ServeHTTP", func(w http.ResponseWriter, r *http.Request) {
		return
	})
	defer mockSever.Reset()
	convey.Convey("cert revoked ,no error returned", t, func() {
		mock := gomonkey.ApplyFunc(CheckRevokedCert, func(r *http.Request, crlcerList *pkix.CertificateList) bool {
			return true
		})
		defer mock.Reset()
		w := &testWriter{}
		Interceptor(http.DefaultServeMux, crlist).ServeHTTP(w, r)
		convey.So(w.Header().Get("Strict-Transport-Security"), convey.ShouldEqual, "")
	})
	convey.Convey("cert didn't revoked ,no error returned", t, func() {
		mock := gomonkey.ApplyFunc(CheckRevokedCert, func(r *http.Request, crlcerList *pkix.CertificateList) bool {
			return false
		})
		defer mock.Reset()
		w := &testWriter{}
		Interceptor(http.DefaultServeMux, crlist).ServeHTTP(w, r)
		convey.So(w.Header().Get("Strict-Transport-Security"), convey.ShouldEqual, "")
	})
}

type testWriter struct {
	http.ResponseWriter
}

func (w *testWriter) WriteHeader(statusCode int) {
	w.Header().Set("status", "200")
}
func (w *testWriter) Header() http.Header {
	return http.Header{"content": {"xxx"}}
}
