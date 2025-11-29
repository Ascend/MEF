// Copyright (c) 2024. Huawei Technologies Co., Ltd.  All rights reserved.

package config

import (
	"errors"
	"sync"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"huawei.com/mindxedge/base/common"
	"huawei.com/mindxedge/base/common/requests"
)

var testData = "test cert data"
var testData2 = "test cert data2"

func TestCertsCache(t *testing.T) {
	convey.Convey("test concurrently get certs should be success", t, testConcurrency)
	convey.Convey("test get certs after update certs' cache should be success", t, testUpdateCert)
	convey.Convey("test get certs with wrong name should be failed", t, testErrorName)
	convey.Convey("test get crl should be success", t, testGetCrl)
}

func testConcurrency() {
	outputsGetRootCa := []gomonkey.OutputCell{
		{Values: gomonkey.Params{"", test.ErrTest}},
		{Values: gomonkey.Params{testData, nil}},
		{Values: gomonkey.Params{testData2, nil}},
	}
	p1 := gomonkey.ApplyMethodSeq(&requests.ReqCertParams{}, "GetRootCa", outputsGetRootCa).
		ApplyMethodReturn(&requests.ReqCertParams{}, "GetCrl", "", nil)
	defer p1.Reset()

	var softwareCertCrlPair, imageCertCrlPair CertCrlPair
	var err1, err2, err3 error

	group := sync.WaitGroup{}
	group.Add(1)
	softwareCertCrlPair, err1 = GetCertCrlPairCache(common.SoftwareCertName)
	go func() {
		imageCertCrlPair, err2 = GetCertCrlPairCache(common.ImageCertName)
		group.Done()
	}()
	group.Wait()
	convey.So(softwareCertCrlPair.CertPEM, convey.ShouldResemble, "")
	convey.So(err1, convey.ShouldResemble, test.ErrTest)
	convey.So(err2, convey.ShouldBeNil)
	convey.So(imageCertCrlPair.CertPEM, convey.ShouldResemble, testData)

	softwareCertCrlPair, err3 = GetCertCrlPairCache(common.SoftwareCertName)
	convey.So(err3, convey.ShouldBeNil)
	convey.So(softwareCertCrlPair.CertPEM, convey.ShouldResemble, testData2)
}

func testUpdateCert() {
	SetCertCrlPairCache(common.SoftwareCertName, testData2, "")
	certCrlPair, err := GetCertCrlPairCache(common.SoftwareCertName)
	convey.So(certCrlPair.CertPEM, convey.ShouldResemble, testData2)
	convey.So(err, convey.ShouldBeNil)
}

func testErrorName() {
	_, err := GetCertCrlPairCache("errorName")
	convey.So(err, convey.ShouldResemble, errors.New("unknown cert name"))
}

func testGetCrl() {
	testcases := []struct {
		description string
		crlStr      string
		crlErr      error
		wantErr     bool
	}{
		{
			description: "get crl successfully",
		},
		{
			description: "get crl failed",
			crlErr:      errors.New("get crl failed"),
			wantErr:     true,
		},
	}

	for _, tc := range testcases {
		convey.Convey(tc.description, func() {
			patches := gomonkey.ApplyMethodReturn(&requests.ReqCertParams{}, "GetRootCa", "", nil).
				ApplyMethodReturn(&requests.ReqCertParams{}, "GetCrl", tc.crlStr, tc.crlErr)
			defer patches.Reset()

			verb := convey.ShouldBeNil
			if tc.wantErr {
				verb = convey.ShouldNotBeNil
			}
			_, err := getCertFromCertMgr(common.SoftwareCertName)
			convey.So(err, verb)
			if !tc.wantErr {
				softwareCertCrlPair, err := GetCertCrlPairCache(common.SoftwareCertName)
				convey.So(softwareCertCrlPair.CrlPEM, convey.ShouldEqual, tc.crlStr)
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}
