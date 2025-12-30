// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package config test for version xml manager
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/fileutils"

	"edge-installer/pkg/common/constants"
)

func prepareXmlFile() string {
	xmlContent := `<?xml version="1.0" encoding="utf-8"?>
<SoftwarePackage version="V1">
    <Package>
        <FileName>Ascend-mindxedge-mefedge_5.0.RC1_linux-aarch64.tar.gz</FileName>
        <OutterName>MEFEdge</OutterName>
        <Version>5.0.RC2</Version>
        <InnerVersion>1.0</InnerVersion>
        <FileType>Software</FileType>
        <Vendor>Huawei Technologies Co., Ltd</Vendor>
        <SupportModel>Linux</SupportModel>
        <ProcessorArchitecture>aarch64</ProcessorArchitecture>
        <CommitId>66d9f1921ddb3900d947fe7c5221b061c68ca779</CommitId>
    </Package>
</SoftwarePackage>`
	xmlFile := filepath.Join("./", constants.VersionXml)
	if err := os.Remove(xmlFile); err != nil && errors.Is(err, os.ErrExist) {
		fmt.Printf("cleanup version xml file failed, error: %v", err)
		return ""
	}
	if err := os.WriteFile(xmlFile, []byte(xmlContent), constants.Mode600); err != nil {
		fmt.Printf("write file failed, error: %v\n", err)
		return ""
	}
	xmlRealFile, err := filepath.Abs(xmlFile)
	if err != nil {
		fmt.Printf("get xml abs path failed: %v", err)
		return ""
	}
	return xmlRealFile
}

func TestNewVersionXmlMgr(t *testing.T) {
	convey.Convey("new version xml mgr should be success", t, func() {
		xmlMgr := NewVersionXmlMgr("./")
		convey.So(xmlMgr.versionPath, convey.ShouldEqual, "./")
	})
}

func TestGetInnerVersion(t *testing.T) {
	convey.Convey("get inner version should be success, test method GetSftPkgName", t, testGetSftPkgName)
	convey.Convey("get inner version should be success, test method GetInnerVersion", t, testGetInnerVersion)
	convey.Convey("get inner version should be failed, real file checker failed", t, testGetInnerVersionErrChecker)
	convey.Convey("get inner version should be failed, read file failed", t, testGetInnerVersionErrReadFile)
	convey.Convey("get inner version should be failed, get regex str failed", t, testGetInnerVersionErrGetRegexStr)
	convey.Convey("get inner version should be failed, sub match len error", t, testGetInnerVersionErrSubMatchLen)
	convey.Convey("get inner version should be failed, regexp compile failed", t, testGetInnerVersionErrRegexp)
}

func testGetSftPkgName() {
	xmlFile := prepareXmlFile()
	versionXmlMgr := NewVersionXmlMgr(xmlFile)
	version, err := versionXmlMgr.GetSftPkgName()
	convey.So(err, convey.ShouldBeNil)
	convey.So(version, convey.ShouldEqual, "MEFEdge")
}

func testGetInnerVersion() {
	xmlFile := prepareXmlFile()
	versionXmlMgr := NewVersionXmlMgr(xmlFile)
	version, err := versionXmlMgr.GetInnerVersion()
	convey.So(err, convey.ShouldBeNil)
	convey.So(version, convey.ShouldEqual, "1.0")
}

func testGetInnerVersionErrChecker() {
	xmlFile := prepareXmlFile()
	versionXmlMgr := NewVersionXmlMgr(xmlFile)
	var p1 = gomonkey.ApplyFunc(fileutils.RealFileCheck,
		func(path string, checkParent, allowLink bool, size int64) (string, error) {
			return "", testErr
		})
	defer p1.Reset()
	_, err := versionXmlMgr.GetInnerVersion()
	convey.So(err, convey.ShouldResemble, testErr)
}

func testGetInnerVersionErrReadFile() {
	xmlFile := prepareXmlFile()
	versionXmlMgr := NewVersionXmlMgr(xmlFile)
	var p1 = gomonkey.ApplyFunc(fileutils.LoadFile,
		func(filePath string, _ ...fileutils.FileChecker) ([]byte, error) {
			return []byte{}, testErr
		})
	defer p1.Reset()
	_, err := versionXmlMgr.GetInnerVersion()
	convey.So(err, convey.ShouldResemble, testErr)
}

func testGetInnerVersionErrGetRegexStr() {
	xmlFile := prepareXmlFile()
	versionXmlMgr := NewVersionXmlMgr(xmlFile)
	var c VersionXmlMgr
	var p1 = gomonkey.ApplyPrivateMethod(reflect.TypeOf(c), "getRegexStr",
		func(vxm VersionXmlMgr, fieldName string) (string, error) {
			return "", testErr
		})
	defer p1.Reset()
	_, err := versionXmlMgr.GetInnerVersion()
	convey.So(err, convey.ShouldResemble, testErr)
}

func testGetInnerVersionErrSubMatchLen() {
	xmlFile := prepareXmlFile()
	versionXmlMgr := NewVersionXmlMgr(xmlFile)
	var c *regexp.Regexp
	var p1 = gomonkey.ApplyMethod(reflect.TypeOf(c), "FindSubmatch",
		func(re *regexp.Regexp, b []byte) [][]byte {
			return [][]byte{}
		})
	defer p1.Reset()
	_, err := versionXmlMgr.GetInnerVersion()
	convey.So(err, convey.ShouldResemble, errors.New("find value from version xml data failed"))
}

func testGetInnerVersionErrRegexp() {
	xmlFile := prepareXmlFile()
	versionXmlMgr := NewVersionXmlMgr(xmlFile)
	var p1 = gomonkey.ApplyFunc(regexp.Compile,
		func(expr string) (*regexp.Regexp, error) {
			return nil, testErr
		})
	defer p1.Reset()
	_, err := versionXmlMgr.GetInnerVersion()
	convey.So(err, convey.ShouldResemble, testErr)
}

func TestGetVersion(t *testing.T) {
	xmlFile := prepareXmlFile()
	versionXmlMgr := NewVersionXmlMgr(xmlFile)
	convey.Convey("get version should be success", t, func() {
		version, err := versionXmlMgr.GetVersion()
		convey.So(err, convey.ShouldBeNil)
		convey.So(version, convey.ShouldEqual, "5.0.RC2")
	})
}
