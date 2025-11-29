//  Copyright(c) 2023. Huawei Technologies Co.,Ltd.  All rights reserved.

package x509

import (
	"encoding/base64"
	"errors"
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCheckCertChain(t *testing.T) {
	convey.Convey("test for CheckCertChain", t, func() {
		convey.Convey("check cert chain success", func() {
			content, err := os.ReadFile("./testdata/ca_chain.crt")
			convey.So(err, convey.ShouldBeNil)
			mgr, err := NewCaChainMgr(content)
			convey.So(err, convey.ShouldBeNil)
			err = mgr.CheckCertChain()
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("check cert chain fail", func() {
			content, err := os.ReadFile("./testdata/ca_chain_err.crt")
			convey.So(err, convey.ShouldBeNil)
			mgr, err := NewCaChainMgr(content)
			convey.So(err, convey.ShouldBeNil)
			err = mgr.CheckCertChain()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestCheckCertsChainReturnContent(t *testing.T) {
	convey.Convey(" test for CheckCertsChainReturnContent", t, func() {
		convey.Convey("check cert chain success", func() {
			content, err := os.ReadFile("./testdata/ca_chain.crt")
			convey.So(err, convey.ShouldBeNil)
			contentByFunc, err := CheckCertsChainReturnContent("./testdata/ca_chain.crt")
			convey.So(err, convey.ShouldBeNil)
			convey.So(content, convey.ShouldResemble, contentByFunc)
		})

		convey.Convey("check cert chain file not exist", func() {
			_, err := CheckCertsChainReturnContent("./testdata/notexists")
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})

		convey.Convey("check cert chain failed", func() {
			_, err := CheckCertsChainReturnContent("./testdata/ca_chain_err.crt")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestGetCerts(t *testing.T) {
	convey.Convey("test for GetCerts func", t, func() {
		convey.Convey("check GetCerts success", func() {
			_, err := GetCerts("./testdata/ca_chain.crt")
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("check GetCerts no file", func() {
			_, err := GetCerts("./testdata/notexists")
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})

		convey.Convey("check GetCerts no PEM cert", func() {
			_, err := GetCerts("./testdata/mainks")
			convey.So(err, convey.ShouldResemble, errors.New("new ca Chain Mgr failed: no PEM cert contains"))
		})

		convey.Convey("cehck GetCerts failed: wrong cert", func() {
			_, err := GetCerts("./testdata/wrong_cert.crt")
			convey.So(err, convey.ShouldResemble, errors.New("new ca Chain Mgr failed: parse cert failed"))
		})
	})
}

func TestCheckDerCertChain(t *testing.T) {
	convey.Convey("test for CheckDerCertChain func", t, func() {
		convey.Convey("check DerCerts success", func() {
			content, err := os.ReadFile("./testdata/der_chain.crt")
			convey.So(err, convey.ShouldBeNil)
			certContent, err := base64.StdEncoding.DecodeString(string(content))
			convey.So(err, convey.ShouldBeNil)
			err = CheckDerCertChain(certContent)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("check DerCerts failed, content is nil", func() {
			var content []byte
			err := CheckDerCertChain(content)
			convey.So(err, convey.ShouldResemble, errors.New("new ca Chain Mgr failed: content is empty"))
		})

		convey.Convey("check DerCerts failed, not DER cert", func() {
			content, err := os.ReadFile("./testdata/ca_chain.crt")
			convey.So(err, convey.ShouldBeNil)
			err = CheckDerCertChain(content)
			convey.So(err, convey.ShouldResemble,
				errors.New("new ca Chain Mgr failed: parse Certificates failed: x509: malformed certificate"))
		})
	})
}

func TestCheckPemCertChain(t *testing.T) {
	convey.Convey("test for CheckPemCertChain func", t, func() {
		content, err := os.ReadFile("./testdata/ca_chain.crt")
		convey.So(err, convey.ShouldBeNil)
		err = CheckPemCertChain(content)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCheckCertsOverdue(t *testing.T) {
	const validOverdueTime = 100
	const invalidOverdueTime = 100000
	convey.Convey("test for CheckCertsOverdue func", t, func() {
		content, err := os.ReadFile("./testdata/ca_chain.crt")
		convey.So(err, convey.ShouldBeNil)
		err = CheckCertsOverdue(content, validOverdueTime)
		convey.So(err, convey.ShouldBeNil)
		err = CheckCertsOverdue(content, invalidOverdueTime)
		convey.So(err.Error(), convey.ShouldContainSubstring, "need to update certification")
	})
}
