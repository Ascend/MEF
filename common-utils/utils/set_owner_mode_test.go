// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils for unit test of set owner and mode functions
package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

var (
	testDir                      = "./test_set_owner"
	testFile                     = "./test_set_owner/tmp/go.mod"
	testNotExistPath             = "/test/xxx"
	testMode         os.FileMode = 0700
	testInvalidMode  os.FileMode = 0777
)

func setup() error {
	if err := MakeSureDir(testFile); err != nil {
		fmt.Printf("create test path failed, error: %v\n", err)
		return err
	}
	if err := CopyFile("./go.mod", testFile); err != nil {
		fmt.Printf("copy test file failed, error: %v\n", err)
		return err
	}
	absDir, err := filepath.Abs(testDir)
	if err != nil {
		return err
	}
	testDir = absDir
	absFile, err := filepath.Abs(testFile)
	if err != nil {
		return err
	}
	testFile = absFile
	return nil
}

func teardown() error {
	if err := os.RemoveAll(testDir); err != nil {
		fmt.Printf("remove test path failed, error: %v\n", err)
		return err
	}
	return nil
}

// TestMain setups environment
func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		return
	}
	exitCode := m.Run()
	if err := teardown(); err != nil {
		return
	}
	fmt.Printf("test complete, exitCode=%d\n", exitCode)
}

func TestSetPathOwnerGroup(t *testing.T) {
	convey.Convey("SetPathOwnerGroup should be success", t, func() {
		convey.Convey("set owner with ignore file recursively", func() {
			err := SetPathOwnerGroup(testDir, uint32(os.Geteuid()), uint32(os.Getegid()), true, true)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("set owner without ignore file recursively", func() {
			err := SetPathOwnerGroup(testDir, uint32(os.Geteuid()), uint32(os.Getegid()), true, false)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("set owner without recursion", func() {
			err := SetPathOwnerGroup(testDir, uint32(os.Geteuid()), uint32(os.Getegid()), false, true)
			convey.So(err, convey.ShouldBeNil)
		})
	})

	convey.Convey("SetPathOwnerGroup should be failed", t, func() {
		convey.Convey("path is empty", func() {
			err := SetPathOwnerGroup("", uint32(os.Geteuid()), uint32(os.Getegid()), true, true)
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})
		convey.Convey("path does not exist", func() {
			err := SetPathOwnerGroup(testNotExistPath, uint32(os.Geteuid()), uint32(os.Getegid()), true, false)
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})
	})
}

func TestSetPathPermission(t *testing.T) {
	convey.Convey("SetPathPermission should be success", t, func() {
		convey.Convey("set permission with ignore file recursively", func() {
			err := SetPathPermission(testDir, testMode, true, true)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("set permission without ignore file recursively", func() {
			err := SetPathPermission(testDir, testMode, true, false)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("set permission without recursion", func() {
			err := SetPathPermission(testDir, testMode, false, true)
			convey.So(err, convey.ShouldBeNil)
		})
	})

	convey.Convey("SetPathPermission should be failed", t, func() {
		convey.Convey("path is empty", func() {
			err := SetPathPermission("", testMode, true, true)
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})
		convey.Convey("path does not exist", func() {
			err := SetPathPermission(testNotExistPath, testMode, true, false)
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})
	})
}

func TestSetParentPathPermission(t *testing.T) {
	convey.Convey("SetParentPathPermission should be success", t, func() {
		p := gomonkey.ApplyFuncReturn(os.Chmod, nil)
		defer p.Reset()
		err := SetParentPathPermission(testDir, testMode)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("SetParentPathPermission should be failed", t, func() {
		convey.Convey("set mode is invalid", func() {
			err := SetParentPathPermission(testDir, testInvalidMode)
			convey.So(err, convey.ShouldResemble, errors.New("cannot set write permission for group or others"))
		})
		convey.Convey("path is empty", func() {
			err := SetParentPathPermission("", testMode)
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})
		convey.Convey("path does not exist", func() {
			err := SetParentPathPermission(testNotExistPath, testMode)
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})
	})
}
