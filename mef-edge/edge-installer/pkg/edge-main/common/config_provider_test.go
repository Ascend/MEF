// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package common for test config provider
package common

import (
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/test"
	"huawei.com/mindx/common/x509/certutils"

	"edge-installer/pkg/common/constants"
	"edge-installer/pkg/common/path"
)

var (
	testDir       = "/tmp/test_config_provider_dir"
	rootCaContent = []byte("test root cert content")
	peerCaContent = []byte("test peer cert content")
)

func TestGetWsCertContent(t *testing.T) {
	innerCertDir := filepath.Join(testDir, constants.InnerCertPathName)
	p := gomonkey.ApplyFuncReturn(path.GetCompSpecificDir, innerCertDir, nil).
		ApplyFuncReturn(fileutils.IsExist, true).
		ApplyFuncReturn(certutils.GetCertContentWithBackup, rootCaContent, nil)
	defer p.Reset()

	convey.Convey("get ws cert content should be success", t, getWsCertContentSuccess)
	convey.Convey("get ws cert content should be failed, get inner cert path failed", t, getInnerCertPathFailed)
	convey.Convey("get ws cert content should be failed, get root cert content failed", t, getRootCertContentFailed)
	convey.Convey("get ws cert content should be failed, get peer cert path failed", t, getPeerCertPathFailed)
	convey.Convey("get ws cert content should be failed, get peer cert content failed", t, getPeerCertContentFailed)
}

func getWsCertContentSuccess() {
	p := gomonkey.ApplyFuncSeq(certutils.GetCertContentWithBackup, []gomonkey.OutputCell{
		{Values: gomonkey.Params{rootCaContent, nil}},
		{Values: gomonkey.Params{peerCaContent, nil}},
	})
	defer p.Reset()
	expectContent := append(rootCaContent, peerCaContent...)
	certContents, err := GetWsCertContent()
	convey.So(certContents.RootCaContent, convey.ShouldResemble, expectContent)
	convey.So(err, convey.ShouldBeNil)
}

func getInnerCertPathFailed() {
	p := gomonkey.ApplyFuncReturn(path.GetCompSpecificDir, "", test.ErrTest)
	defer p.Reset()
	certContents, err := GetWsCertContent()
	convey.So(certContents, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func getRootCertContentFailed() {
	p := gomonkey.ApplyFuncReturn(certutils.GetCertContentWithBackup, []byte(""), test.ErrTest)
	defer p.Reset()
	certContents, err := GetWsCertContent()
	convey.So(certContents, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func getPeerCertPathFailed() {
	p := gomonkey.ApplyFuncReturn(path.GetCompSpecificDir, "", test.ErrTest)
	defer p.Reset()
	certContents, err := GetWsCertContent()
	convey.So(certContents, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}

func getPeerCertContentFailed() {
	p := gomonkey.ApplyFuncSeq(certutils.GetCertContentWithBackup, []gomonkey.OutputCell{
		{Values: gomonkey.Params{rootCaContent, nil}},
		{Values: gomonkey.Params{[]byte(""), test.ErrTest}},
	})
	defer p.Reset()
	certContents, err := GetWsCertContent()
	convey.So(certContents, convey.ShouldBeNil)
	convey.So(err, convey.ShouldResemble, test.ErrTest)
}
