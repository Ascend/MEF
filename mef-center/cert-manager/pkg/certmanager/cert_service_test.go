// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package certmanager

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/backuputils"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/modulemgr/model"
	"huawei.com/mindx/common/rand"
	"huawei.com/mindx/common/test"
	commonX509 "huawei.com/mindx/common/x509"
	"huawei.com/mindx/common/x509/certutils"

	"huawei.com/mindxedge/base/common"
)

var testCrt = `LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVyRENDQXhTZ0F3SUJBZ0lVR3RBTGxtL2ZuMkFnZFlUTDVLUnBMOEYyRDdrd
0RRWUpLb1pJaHZjTkFRRUwKQlFBd1h6RUxNQWtHQTFVRUJoTUNRMDR4RURBT0JnTlZCQWNNQjBOb1pXNW5aSFV4RHpBTkJnTlZCQW9NQmtoMQpZWGRsY
VRFV01CUUdBMVVFQ3d3TlNIVmhkMlZwSUVGelkyVnVaREVWTUJNR0ExVUVBd3dNVkVWVFZDQlNUMDlVCklFTkJNQjRYRFRJME1ESXlNREF5TlRVMU1Wb
1hEVE0wTURJeE56QXlOVFUxTVZvd1h6RUxNQWtHQTFVRUJoTUMKUTA0eEVEQU9CZ05WQkFjTUIwTm9aVzVuWkhVeER6QU5CZ05WQkFvTUJraDFZWGRsY
VRFV01CUUdBMVVFQ3d3TgpTSFZoZDJWcElFRnpZMlZ1WkRFVk1CTUdBMVVFQXd3TVZFVlRWQ0JTVDA5VUlFTkJNSUlCb2pBTkJna3Foa2lHCjl3MEJBU
UVGQUFPQ0FZOEFNSUlCaWdLQ0FZRUF2UkkxdjZuaGRsdlFPSlp1RW1CQmIwRWd5RlRBRXRFOG90VVUKZG9qb0d3Rk5SNkJxbmIxTFl1V1k5UGNjUlZKM
GNOZmh2K1NWdktNZVVmWFR0cHQ5MGJiQWdGZHFDTlpEYUZkagoyUklhVzVmTGVvUlNNWDRmTE1IZjViWTYxT2xrYlRIWURPYzNIMHdtQWRraHZFajdPR
CtTUjB3RVZRazBvUnVJCitZbllBQ1dLSU1zZTc2MHhaak9aVkMvMnpLTDhCZnI1RllLUlZadDN5dkt1TzZid3QrdVRnbi9KQTlMelNidmIKMms1SHJ5V
HlMWUlyTi94QmNJMjU2WUo0Q1FqdFJ3NGZpZ1Q2SE5UVFBmaUpITFVHZENqZ24wbjN2T1N6T3Nqdwp5OHpEVjJ3YXVTM0p5QmxuRVF2Tzc0OXB4cFpob
y9POEtVWWNoNVg0NVNiZjNsTFBLRmcxUnhvUTd5aUhMYytiCnJUUmZQTklWZnFtOVVSbjQ5TlpyRjQ5QndYTXovMVRMV0pvcWYwQ0wxNC9USC9ucW41U
Hk0a3RWeDF1YWdsL3IKRko1QnBJVml3TzhhcE9wMTJKR3haa3NVNnREVHpTN3djdjZRSmRXYjZsOTVNanZYNlJpSHRaZk9rZUMzT1QwUwpZV2JScmpjZ
C9IOUZnaXk0U29ONWNMUEtPb3dEQWdNQkFBR2pZREJlTUIwR0ExVWREZ1FXQkJUdUF0MGdEOTFKCi83TTMvazF4QzFSd1NnYlFsREFmQmdOVkhTTUVHR
EFXZ0JUdUF0MGdEOTFKLzdNMy9rMXhDMVJ3U2diUWxEQVAKQmdOVkhSTUJBZjhFQlRBREFRSC9NQXNHQTFVZER3UUVBd0lCQmpBTkJna3Foa2lHOXcwQ
kFRc0ZBQU9DQVlFQQpwVHljWXJTZzA1SVJSMXJnblllY1A5bXZUWHlnNEQ3TEhuU0IwZFloaVJIa2hiYU03U09FWjJFVldCVFRGc1QxClZRZmlZbGxrZ
nNOcnRtbUNJaUlROHBadExYREcyUGZyVFVuajlMWW9OWmh4NGxzREJXNGVLQm9KS1BqN3Z1NUUKVVh4OEUyMnpGUVhFTFNja2c3RWRqd2phb2wzUGtOb
2doLzhka2dIcENCdU9sMEF6ZTZzcXhqdzRjNnBDQVh1SgpqZUZRWGQzaVpGb2FmMnptUVpPamdaSFcvZklISVRFNmRzcVp0aXNFVnJ2em9SY2pibjFWW
DhrTU9sTjBrM0x0CmpKS3BMU29OWlpDTEdrMkZYcWhyd0ZnbS8zbWUyVGU4ZDFoUXExMlRWZGUwd3FCYXl2aklzbUovbDRVNnRyRVIKN0t1Vk1ZeXE5V
GNsZHFyeE0yejAvVVk4OW5PbG1lVjBTeGhmTVZ4VGJ6N0FXYkJPNnU5b2RldkVhM1hGMktlWgpoWE1CRU1WOGZ5VVNRTERmRm1ET0ZMZzJtZk5pR1RLc
UdjbVZWZElEMmkrQjQyWnZrdW80dDBlQ0w3MW5QYjJqClcvTW9JVW90Z20zK3FpNHg5OUkySDg1N0tHRkQ0bjdreGcrZC9wZzRkRXM5R0FtK2VkaGJnb
WJISU9xeTdoeDcKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQ==`

func TestQueryRootCa(t *testing.T) {
	convey.Convey("test queryRootCa success cases", t, testQueryRootCaSuccessfulCases)
	convey.Convey("test queryRootCa error cases", t, testQueryRootCaFailedCases)
}

func testQueryRootCaSuccessfulCases() {
	convey.Convey("case: cert does not exist", func() {
		patches := gomonkey.ApplyFuncReturn(isCertImported, false)
		defer patches.Reset()
		msg := newMsgWithContentForUT(common.NorthernCertName)
		resp := queryRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, fmt.Sprintf("%s cert is no imported yet", common.NorthernCertName))
	})

	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(isCertImported, true).
			ApplyFuncReturn(getCertByCertName, []byte("test cert content"), nil)
		defer patches.Reset()
		msg := newMsgWithContentForUT(common.NorthernCertName)
		resp := queryRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "query ca success")
	})
}

func testQueryRootCaFailedCases() {
	convey.Convey("case: incorrect cert name", func() {
		msg := newMsgWithContentForUT("error-type")
		resp := queryRootCa(msg)
		convey.So(resp.Status, convey.ShouldResemble, common.ErrorParamInvalid)
	})

	convey.Convey("case: failed to get cert", func() {
		patches := gomonkey.ApplyFuncReturn(isCertImported, true).
			ApplyFuncReturn(getCertByCertName, []byte{}, test.ErrTest)
		defer patches.Reset()
		msg := newMsgWithContentForUT(common.NorthernCertName)
		resp := queryRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "Query root ca failed")
	})

	convey.Convey("case: failed to get temp hub_client cert", func() {
		patches := gomonkey.ApplyFuncReturn(isCertImported, true).
			ApplyFuncReturn(getCertByCertName, []byte("test cert content"), nil).
			ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(certutils.GetCertContent, []byte(""), test.ErrTest)
		defer patches.Reset()
		msg := newMsgWithContentForUT(common.WsCltName)
		resp := queryRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "failed to load new temp root ca")
	})
}

func TestIssueServiceCa(t *testing.T) {
	convey.Convey("test issueServiceCa success cases", t, testIssueServiceCaSuccessfulCases)
	convey.Convey("test issueServiceCa error cases", t, testIssueServiceCaFailedCases)
}

func getTestCsrReq() csrJson {
	const validBitSizeForKey = 3072
	key, err := rsa.GenerateKey(rand.Reader, validBitSizeForKey)
	if err != nil {
		panic(err)
	}
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Organization:       []string{"Huawei"},
			OrganizationalUnit: []string{"CPL Ascend"},
			CommonName:         "test-cert-name",
		},
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, &template, key)
	if err != nil {
		panic(err)
	}
	csrReq := csrJson{
		CertName: common.NginxCertName,
		Csr:      base64.StdEncoding.EncodeToString(certutils.PemWrapCert(csr)),
	}
	return csrReq
}

func testIssueServiceCaSuccessfulCases() {
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(issueServiceCert, []byte("test-cert"), nil)
		defer patches.Reset()
		msg := newMsgWithContentForUT(getTestCsrReq())
		resp := issueServiceCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "issue success")
	})
}

func testIssueServiceCaFailedCases() {
	convey.Convey("case: incorrect content", func() {
		msg := newMsgWithContentForUT("test content")
		resp := issueServiceCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "parse content failed")
	})

	convey.Convey("case: check of cert name failed", func() {
		req := getTestCsrReq()
		req.CertName = "test cert name"
		msg := newMsgWithContentForUT(req)
		resp := issueServiceCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "cert issue para check failed")
	})

	convey.Convey("case: failed to issue service cert", func() {
		patches := gomonkey.ApplyFuncReturn(issueServiceCert, nil, test.ErrTest)
		defer patches.Reset()
		msg := newMsgWithContentForUT(getTestCsrReq())
		resp := issueServiceCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "issue service certificate failed")
	})
}

func TestCertsUpdateResult(t *testing.T) {
	convey.Convey("test certsUpdateResult success cases", t, testCertsUpdateResultSuccessfulCases)
	convey.Convey("test certsUpdateResult error cases", t, testCertsUpdateResultFailedCases)
}

func testCertsUpdateResultSuccessfulCases() {
	convey.Convey("case: normal success", func() {
		msg := newMsgWithContentForUT(certUpdateResult{CertType: CertTypeEdgeCa})
		go func() {
			if edgeCaResultChan != nil {
				_, _ = <-edgeCaResultChan
			}
		}()
		resp := certsUpdateResult(msg)
		convey.So(resp.Status, convey.ShouldResemble, common.Success)
	})
}

func testCertsUpdateResultFailedCases() {
	convey.Convey("case: wrong content type", func() {
		msg := newMsgWithContentForUT("test content")
		resp := certsUpdateResult(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "parse content failed")
	})

	convey.Convey("case: wrong cert type", func() {
		msg := newMsgWithContentForUT(certUpdateResult{CertType: "wrong type"})
		resp := certsUpdateResult(msg)
		convey.So(resp.Status, convey.ShouldResemble, common.ErrorCertTypeError)
	})
}

func TestImportRootCa(t *testing.T) {
	convey.Convey("test importRootCa success cases", t, testImportRootCaSuccessfulCases)
	convey.Convey("test importRootCa error cases", t, testImportRootCaFailedCases)
}

func getTestImportCertReq() importCertReq {
	req := importCertReq{
		CertName: common.SoftwareCertName,
		Cert:     testCrt,
	}
	return req
}

func testImportRootCaSuccessfulCases() {
	convey.Convey("case: normal success", func() {
		msg := newMsgWithContentForUT(getTestImportCertReq())
		resp := importRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "import certificate success")
	})
}

func testImportRootCaFailedCases() {
	convey.Convey("case: incorrect content", func() {
		msg := newMsgWithContentForUT("test content")
		resp := importRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "parse content failed")
	})

	convey.Convey("case: check of name failed", func() {
		req := getTestImportCertReq()
		req.CertName = "test name"
		msg := newMsgWithContentForUT(req)
		resp := importRootCa(msg)
		convey.So(resp.Status, convey.ShouldResemble, common.ErrorParamInvalid)
	})

	convey.Convey("case: failed to save ca", func() {
		patches := gomonkey.ApplyFuncReturn(saveCaContent, test.ErrTest)
		defer patches.Reset()
		msg := newMsgWithContentForUT(getTestImportCertReq())
		resp := importRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "save ca content to file failed")
	})

}

func TestDeleteRootCa(t *testing.T) {
	convey.Convey("test deleteRootCa success cases", t, testDeleteRootCaSuccessfulCases)
	convey.Convey("test deleteRootCa error cases", t, testDeleteRootCaFailedCases)
}

func testDeleteRootCaSuccessfulCases() {
	convey.Convey("case: normal success", func() {
		msg := newMsgWithContentForUT(deleteCaReq{Type: common.ImageCertName})
		resp := deleteRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "delete ca file success")
	})
}

func testDeleteRootCaFailedCases() {
	convey.Convey("case: incorrect content", func() {
		msg := newMsgWithContentForUT("test content")
		resp := deleteRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "parse content failed")
	})

	convey.Convey("case: check of ca name failed", func() {
		msg := newMsgWithContentForUT(deleteCaReq{Type: common.NginxCertName})
		resp := deleteRootCa(msg)
		convey.So(resp.Status, convey.ShouldResemble, common.ErrorParamInvalid)
	})

	convey.Convey("case: deletion failure", func() {
		patches := gomonkey.ApplyFuncReturn(removeCaFile, test.ErrTest)
		defer patches.Reset()
		msg := newMsgWithContentForUT(deleteCaReq{Type: common.SoftwareCertName})
		resp := deleteRootCa(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "delete ca file failed")
	})
}

func TestGetCertInfo(t *testing.T) {
	convey.Convey("test getCertInfo success cases", t, testGetCertInfoSuccessfulCases)
	convey.Convey("test getCertInfo error cases", t, testGetCertInfoFailedCases)
}

func testGetCertInfoSuccessfulCases() {
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(getCertByCertName, []byte{}, nil).
			ApplyFuncReturn(parseNorthernRootCa, []map[string]interface{}{}, nil)
		defer patches.Reset()
		msg := newMsgWithContentForUT(common.NorthernCertName)
		resp := getCertInfo(msg)
		convey.So(resp.Status, convey.ShouldResemble, common.Success)
	})

}

func testGetCertInfoFailedCases() {
	convey.Convey("case: wrong cert name", func() {
		msg := newMsgWithContentForUT(common.NginxCertName)
		resp := getCertInfo(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "the cert name not support")
	})

	convey.Convey("case: failed to get cert ", func() {
		msg := newMsgWithContentForUT(common.NginxCertName)
		resp := getCertInfo(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "the cert name not support")
	})

	convey.Convey("case: failed to parse cert", func() {
		patches := gomonkey.ApplyFuncReturn(parseNorthernRootCa, []byte{}, test.ErrTest)
		defer patches.Reset()
		msg := newMsgWithContentForUT(common.NginxCertName)
		resp := getCertInfo(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "the cert name not support")
	})
}

func TestImportCrl(t *testing.T) {
	convey.Convey("test importCrl success cases", t, testImportCrlSuccessfulCases)
	convey.Convey("test importCrl error cases", t, testImportCrlFailedCases)
}

func getTestCrlReq() importCrlReq {
	req := importCrlReq{
		CrlName: common.NorthernCertName,
		Crl: `LS0tLS1CRUdJTiBYNTA5IENSTC0tLS0tCk1JSUNPRENCb1FJQkFUQU5CZ2txaGtpRzl3MEJBUXNGQURCZk1Rc3dDUVlEVlFRR0V3Sk
RUakVRTUE0R0ExVUUKQnd3SFEyaGxibWRrZFRFUE1BMEdBMVVFQ2d3R1NIVmhkMlZwTVJZd0ZBWURWUVFMREExSWRXRjNaV2tnUVhOagpaVzVrTVJVd0
V3WURWUVFEREF4VVJWTlVJRkpQVDFRZ1EwRVhEVEkwTURJeU1EQXlOVGt4TWxvWERUTTBNREl4Ck56QXlOVGt4TWxxZ0RqQU1NQW9HQTFVZEZBUURBZ0
VCTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCZ1FDeXlZVDcKWmdRb1RMWWVEVksrVjZmcmJLcFhXaTlpK290YjcyNWhvbjVBK3I3djJHaXZaVDN0TVlYT1
h1U0lBOUpzMEg2NQoveHpWck1KOGhDNlY5Q3FSaUd0WDZpcVRpODAweDYzTExXUGxUV0ZTMWdjQSt2aGtTNHViNzR2MjA4NjJiWVl6CnllVGlVRXpPVV
d1b0RJQkRVMmFRQ1NSYnJiMUVSU20xd25RQmhxdWNUUVA4T0VpUHhneEZobUJPalhXeVMyTnQKakJweWlxWDFEUVZseXV1VUFhUitubkwxbDdjUUdFOC
trbmU0OStuTGcvYWZXYXF0VFVSSS9VR1MraVVjd3BILwpXbjdUVlArTHJBWlFMQW9NZDdNVGd6YTl4TTNoWmM2UDErS1dhU0Y1MGRhYkdYN3IzUFUxWD
dQT25lcVkxUW9BCnYyTkw0RWtTMjkxdHh0NDRYYXlyYnFxOGR2dmtVamliQ3h3bDhtVnhxalg0bXZvT3IxaytuZEdlSEZJT2NHWFAKdS9TYm5DNW1jYk
hwYUNWOTkrcVBnbVRXbWhOQXFyNUtmL1BKSDNzYkVBSVkzK1oyN1NzaUU2OEhNN0hGRUlQKwpCeVhHTEZENU51eUpQS0tnM25GaGJMQ20xdGpBVk5IWU
dsSW5SOHpScDc2VWhSZmh6c0RPV3FnVkNrdz0KLS0tLS1FTkQgWDUwOSBDUkwtLS0tLQ==`,
	}
	return req
}

func testImportCrlSuccessfulCases() {
	convey.Convey("case: normal success", func() {
		req := getTestImportCertReq()
		testCaContent, err := base64.StdEncoding.DecodeString(req.Cert)
		if err != nil {
			panic(err)
		}
		patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, testCaContent, nil).
			ApplyFuncReturn(backuputils.BackUpFiles, nil)
		defer patches.Reset()
		msg := newMsgWithContentForUT(getTestCrlReq())
		resp := importCrl(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "import crl file success")
	})
}

func testImportCrlFailedCases() {
	convey.Convey("case: wrong content", func() {
		msg := newMsgWithContentForUT("test content")
		resp := importCrl(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "parse content failed")
	})

	convey.Convey("case: failed name check", func() {
		req := getTestCrlReq()
		req.CrlName = "test name"
		msg := newMsgWithContentForUT(req)
		resp := importCrl(msg)
		convey.So(resp.Status, convey.ShouldResemble, common.ErrorParamInvalid)
	})

	convey.Convey("case: failed to save crl", func() {
		req := getTestImportCertReq()
		testCaContent, err := base64.StdEncoding.DecodeString(req.Cert)
		if err != nil {
			panic(err)
		}
		patches := gomonkey.ApplyFuncReturn(fileutils.LoadFile, testCaContent, nil).
			ApplyFuncReturn(saveCrlContentWithBackup, test.ErrTest)
		defer patches.Reset()
		msg := newMsgWithContentForUT(getTestCrlReq())
		resp := importCrl(msg)
		convey.So(resp.Status, convey.ShouldResemble, common.ErrorSaveCrl)
	})
}

func TestQueryCrl(t *testing.T) {
	convey.Convey("test queryCrl success cases", t, testQueryCrlSuccessfulCases)
	convey.Convey("test queryCrl error cases", t, testQueryCrlFailedCases)
}

func testQueryCrlSuccessfulCases() {
	convey.Convey("case: crl is not exist", func() {
		msg := newMsgWithContentForUT(common.NorthernCertName)
		resp := queryCrl(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "north crl is no imported yet")
	})

	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(certutils.GetCrlContentWithBackup, []byte("test crl"), nil)
		defer patches.Reset()
		msg := newMsgWithContentForUT(common.NorthernCertName)
		resp := queryCrl(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "query crl success")
		convey.So(resp.Data, convey.ShouldResemble, "test crl")
	})
}

func testQueryCrlFailedCases() {
	convey.Convey("case: wrong cert name", func() {
		msg := newMsgWithContentForUT("test name")
		resp := queryCrl(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "query crl failed parma is invalid")
	})

	convey.Convey("case: failed to check crl", func() {
		patches := gomonkey.ApplyFuncReturn(fileutils.IsExist, true).
			ApplyFuncReturn(certutils.GetCrlContentWithBackup, nil, test.ErrTest)
		defer patches.Reset()
		msg := newMsgWithContentForUT(common.NorthernCertName)
		resp := queryCrl(msg)
		convey.So(resp.Msg, convey.ShouldResemble, "query crl failed, crl file is damaged")
	})
}

func TestGetImportedCertsInfo(t *testing.T) {
	convey.Convey("test getImportedCertsInfo success cases", t, testGetImportedCertsInfoSuccessfulCases)
	convey.Convey("test getImportedCertsInfo error cases", t, testGetImportedCertsInfoFailedCases)
}

func testGetImportedCertsInfoSuccessfulCases() {
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(commonX509.CheckCertsChainReturnContent, []byte(""), nil)
		defer patches.Reset()
		resp := getImportedCertsInfo(&model.Message{})
		convey.So(resp.Msg, convey.ShouldResemble, "get imported certs info success")
	})
}

func testGetImportedCertsInfoFailedCases() {
	convey.Convey("case: normal success", func() {
		patches := gomonkey.ApplyFuncReturn(json.Marshal, nil, test.ErrTest)
		defer patches.Reset()
		resp := getImportedCertsInfo(&model.Message{})
		convey.So(resp.Msg, convey.ShouldResemble, "get imported certs info failed")
	})
}

func TestParseNorthernRootCa(t *testing.T) {
	convey.Convey("case: normal success", t, func() {
		caBase64, decodeErr := base64.StdEncoding.DecodeString(testCrt)
		convey.So(decodeErr, convey.ShouldBeNil)
		_, err := parseNorthernRootCa(caBase64)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("case: parse cert failed", t, func() {
		_, err := parseNorthernRootCa([]byte(""))
		convey.So(err, convey.ShouldNotBeNil)
	})
}
