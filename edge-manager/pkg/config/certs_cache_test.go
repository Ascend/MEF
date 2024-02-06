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
}

func testConcurrency() {
	outputsGetRootCa := []gomonkey.OutputCell{
		{Values: gomonkey.Params{"", test.ErrTest}},
		{Values: gomonkey.Params{testData, nil}},
		{Values: gomonkey.Params{testData2, nil}},
	}
	p1 := gomonkey.ApplyMethodSeq(&requests.ReqCertParams{}, "GetRootCa", outputsGetRootCa)
	defer p1.Reset()

	var softwareCert, imageCert string
	var err1, err2, err3 error

	group := sync.WaitGroup{}
	group.Add(1)
	softwareCert, err1 = GetCertCache(common.SoftwareCertName)
	go func() {
		imageCert, err2 = GetCertCache(common.ImageCertName)
		group.Done()
	}()
	group.Wait()
	convey.So(softwareCert, convey.ShouldResemble, "")
	convey.So(err1, convey.ShouldResemble, test.ErrTest)
	convey.So(err2, convey.ShouldBeNil)
	convey.So(imageCert, convey.ShouldResemble, testData)

	softwareCert, err3 = GetCertCache(common.SoftwareCertName)
	convey.So(err3, convey.ShouldBeNil)
	convey.So(softwareCert, convey.ShouldResemble, testData2)
}

func testUpdateCert() {
	SetCertCache(common.SoftwareCertName, testData2)
	certString, err := GetCertCache(common.SoftwareCertName)
	convey.So(certString, convey.ShouldResemble, testData2)
	convey.So(err, convey.ShouldBeNil)
}

func testErrorName() {
	_, err := GetCertCache("errorName")
	convey.So(err, convey.ShouldResemble, errors.New("unknown cert name"))
}
