// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package certmgr this file for cert manager
package certmgr

import (
	"fmt"
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/test"
)

func TestMain(m *testing.M) {
	tcBase := &test.TcBase{}
	test.RunWithPatches(tcBase, m, nil)
}

type argsMgr struct {
	certDir        string
	certName       string
	certBackUpName string
	tempCertPath   string
}
type expectedMgr struct {
	SaveCertByFile error
	IsCertExist    bool
	LoadCert       error
}

type testsMgr struct {
	name     string
	args     argsMgr
	Expected expectedMgr
}

func TestCertManager(t *testing.T) {

	tmpfile, cert, certBak := CreatCertFile(t)
	files := []string{
		tmpfile.Name(),
		cert.Name(),
		certBak.Name(),
	}
	defer func() {
		for _, file := range files {
			err := os.Remove(file)
			if err != nil {
				t.Fatal(err)
			}
		}
	}()

	tests := preparingDataMgr(cert, certBak, tmpfile)

	for _, tt := range tests {
		convey.Convey(tt.name, t, func() {
			cm := NewCertMgr(tt.args.certDir, tt.args.certName, tt.args.certBackUpName)

			convey.Convey("TestIsCertExist", func() {
				convey.So(cm.IsCertExist(), convey.ShouldResemble, tt.Expected.IsCertExist)
			})
			convey.Convey("TestSaveCertByFile", func() {
				convey.So(cm.SaveCertByFile(tt.args.tempCertPath), convey.ShouldResemble, tt.Expected.SaveCertByFile)
			})
			convey.Convey("TestLoadCert", func() {
				_, err := cm.LoadCert()
				convey.So(err, convey.ShouldResemble, tt.Expected.LoadCert)
			})
		})
	}
}

func preparingDataMgr(cert *os.File, certBak *os.File, tmpfile *os.File) []testsMgr {
	tests := []testsMgr{{
		name: "Case 1: Normal",
		args: argsMgr{
			certDir:        "",
			certName:       cert.Name(),
			certBackUpName: certBak.Name(),
			tempCertPath:   tmpfile.Name(),
		},
		Expected: expectedMgr{
			SaveCertByFile: nil,
			IsCertExist:    true,
			LoadCert:       nil,
		},
	},
		{
			name: "Case 2: certName not exit",
			args: argsMgr{
				certDir:        "",
				certName:       "",
				certBackUpName: certBak.Name(),
				tempCertPath:   tmpfile.Name(),
			},
			Expected: expectedMgr{
				SaveCertByFile: fmt.Errorf("create backup instance failed"),
				IsCertExist:    false,
				LoadCert:       fmt.Errorf("create backup instance failed"),
			},
		},
		{
			name: "Case 3: BackUpName not exit",
			args: argsMgr{
				certDir:        "",
				certName:       cert.Name(),
				certBackUpName: "",
				tempCertPath:   tmpfile.Name(),
			},
			Expected: expectedMgr{
				SaveCertByFile: fmt.Errorf("create backup instance failed"),
				IsCertExist:    true,
				LoadCert:       fmt.Errorf("create backup instance failed"),
			},
		}}
	return tests
}

func CreatCertFile(t *testing.T) (*os.File, *os.File, *os.File) {
	certContent := []byte(`
-----BEGIN CERTIFICATE-----
MIIENzCCApygAwIBAgIUEIcq8Bihffh1fP6839E25ARqssYwDQYJKoZIhvcNAQEL
BQAwIDEeMBwGA1UECgwVQ2VydGlmaWNhdGUgQXV0aG9yaXR5MB4XDTIzMDYwNzEy
NTMwN1oXDTIzMTIwNDEyNTMwN1owIDEeMBwGA1UECgwVQ2VydGlmaWNhdGUgQXV0
aG9yaXR5MIIBpTANBgkqhkiG9w0BAQEFAAOCAZIAMIIBjQKCAYQAt1PkXHVM3hlO
Ll62FmqozJo/XWOnCmGo975MM4cKhnNd1HSwh/Sbwz5HD83zRQpO+nukJtMMHM5z
4Gs4Y1Qw65tDchLSDTNTVsN+j0D8oS59TReXw1zI4icAsh6u/ycL0jZ4rxUZjdZE
jEDNwaS1C8iDSgrOmYOv9E2G6XuqooT7JZ+Vw6bllhGN2e6Q68gAicJclpBCFVrF
F/IcIjJED5Icm71aXuo2XhR9XBJMLO4hDfAt4xro5q8krsSUCW3WjgR9L1UhB54j
AojQUx2NWRLGyV7dSiqCSBnsoIhR6Iu5+fyJyFWQwI+ycEW/MNd+Ykw2TmopIosj
2Pdq5nlNRhRSKDdrCCb4ExL6726boAABLLbMmAVy85OpsLlctZIaotvf/fE/nwwp
rgP+QPM8wVKjeqe6KX41D6xxZ40XAwUGJsFKi34k3f5uX+qYHG/2qezcCkSp2Uvo
mHymtUuymz6hj7tdxIabG7tCvGnk3eCYEr2UkF3IcxPlg7rtYR0R0apPAgMBAAGj
YzBhMB0GA1UdDgQWBBQVW/d1Qc1C3MiN3IFiqupAm+a7QDAfBgNVHSMEGDAWgBQV
W/d1Qc1C3MiN3IFiqupAm+a7QDAPBgNVHRMBAf8EBTADAQH/MA4GA1UdDwEB/wQE
AwIBhjANBgkqhkiG9w0BAQsFAAOCAYQAfBmv8uK+tGy77HFVpp76TFA232GEWlyS
wTeWo7k88+1e7qS7dWqRWY7ipZZdMzzmwBsoUMV47FEbX4F3X86P4PVxVdnr6CbC
MH6sPk/tOZcsp4E8S09TwD6SlEer60F931J4Pkc1dHKOJjNplzTaz8VP4mUFEYLp
UIgbYQfost2E7IkJgnohTjUdhbF3+n28hmq3Aq/UCk3Xh+lwH/6yzhvb+AvvADKC
/M5WEMqGRKEEZ7fUKDqKql+XDBoJa9wf7aMbaDFuE2TLlbNpHKs/xjfVOW0EbCvo
yU5kw89Yf+fMqOo+RIp78YyKVFYjEVZaL34X2d/eBX2V+o1Ux5OJSwuk12+3Uvsz
tkyUKX8H5wYViPcNPw2hmPQKIubpinzQ9RBnSHFblNwWHlqJ4wpf5YjTGSg8uNWK
e6Bt4vpbqfUDMCvCE8yFBxNs8u9wQ8ifDw5g2rYjAjTJRs4yNENrXXlPr4QUgQys
lqo88spVNrbsauGVVeNcOTfGDj9iQ9n1f6AS
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
