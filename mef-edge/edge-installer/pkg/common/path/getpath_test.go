// Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.

// Package path test for getpath_test.go
package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/test"

	"edge-installer/pkg/common/constants"
)

func TestGetCompWorkDir(t *testing.T) {
	convey.Convey("test func GetCompWorkDir success", t, func() {
		_, err := GetCompWorkDir()
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetCompWorkDir failed, os.Executable failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()
		_, err := GetCompWorkDir()
		expErr := fmt.Errorf("get current path failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetCompConfigDir(t *testing.T) {
	convey.Convey("test func GetCompConfigDir success", t, func() {
		_, err := GetCompConfigDir()
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetCompConfigDir failed, GetCompWorkDir failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()
		_, err := GetCompConfigDir()
		expErr := errors.New("get comp work dir failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetCompConfigDir failed, filepath.EvalSymlinks failed", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", test.ErrTest}},

			{Values: gomonkey.Params{"./", nil}},
			{Values: gomonkey.Params{"", test.ErrTest}},
		}
		var p1 = gomonkey.ApplyFuncSeq(filepath.EvalSymlinks, outputs)
		defer p1.Reset()

		_, err := GetCompConfigDir()
		expErr := errors.New("get comp work dir failed")
		convey.So(err, convey.ShouldResemble, expErr)

		_, err = GetCompConfigDir()
		expErr = fmt.Errorf("eval comp config dir symlink failed, %v", test.ErrTest)
		convey.So(err, convey.ShouldResemble, expErr)
	})
}

func TestGetCompSpecificDir(t *testing.T) {
	convey.Convey("test func GetCompSpecificDir success", t, func() {
		specDir, err := GetCompSpecificDir(constants.InnerCertPathName)
		convey.So(specDir, convey.ShouldResemble, constants.InnerCertPathName)
		convey.So(err, convey.ShouldResemble, nil)
	})

	convey.Convey("test func GetCompSpecificDir failed, os.Executable failed", t, func() {
		var p1 = gomonkey.ApplyFuncReturn(os.Executable, "", test.ErrTest)
		defer p1.Reset()
		_, err := GetCompSpecificDir(constants.InnerCertPathName)
		expErr := errors.New("get comp config dir failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})

	convey.Convey("test func GetCompSpecificDir failed, filepath.EvalSymlinks failed", t, func() {
		outputs := []gomonkey.OutputCell{
			{Values: gomonkey.Params{"", test.ErrTest}},

			{Values: gomonkey.Params{"./", nil}},
			{Values: gomonkey.Params{"", test.ErrTest}},
		}
		var p1 = gomonkey.ApplyFuncSeq(filepath.EvalSymlinks, outputs)
		defer p1.Reset()

		specDir, err := GetCompSpecificDir(constants.InnerCertPathName)
		convey.So(specDir, convey.ShouldResemble, "")
		expErr := errors.New("get comp config dir failed")
		convey.So(err, convey.ShouldResemble, expErr)

		specDir, err = GetCompSpecificDir(constants.InnerCertPathName)
		convey.So(specDir, convey.ShouldResemble, "")
		expErr = errors.New("get comp config dir failed")
		convey.So(err, convey.ShouldResemble, expErr)
	})
}
