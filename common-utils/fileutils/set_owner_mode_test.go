// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

package fileutils

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

const (
	testUid = 1024
	testGid = 2028
	mode755 = 0755
	mode777 = 0777
)

func TestSetPathOwnerGroup(t *testing.T) {
	convey.Convey("set owner non-recursively success", t, testRecursivelySetFileOwnerGroup)
	convey.Convey("set owner recursively success", t, testNonRecursivelySetFileOwnerGroup)
	convey.Convey("set owner ignore file success", t, testIgnoreFileSetFileOwnerGroup)
	convey.Convey("set soft link owner failed", t, testSetSoftlinkFileOwnerGroup)
	convey.Convey("set owner with recursively mode checker failed", t, testSetFileOwnerGroupWithChecker)
}

func testRecursivelySetFileOwnerGroup() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	param := SetOwnerParam{
		Path:       tmpDir,
		Uid:        testUid,
		Gid:        testGid,
		Recursive:  false,
		IgnoreFile: false,
	}
	err = SetPathOwnerGroup(param)
	convey.So(err, convey.ShouldBeNil)
	dirInfo, err := os.Stat(tmpDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(dirInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, testUid)
	convey.So(dirInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, testGid)
	fileInfo, err := os.Stat(filePath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, 0)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, 0)
}

func testNonRecursivelySetFileOwnerGroup() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	param := SetOwnerParam{
		Path:       tmpDir,
		Uid:        testUid,
		Gid:        testGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	err = SetPathOwnerGroup(param)
	convey.So(err, convey.ShouldBeNil)
	dirInfo, err := os.Stat(tmpDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(dirInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, testUid)
	convey.So(dirInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, testGid)
	fileInfo, err := os.Stat(filePath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, testUid)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, testGid)
}

func testIgnoreFileSetFileOwnerGroup() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	param := SetOwnerParam{
		Path:       tmpDir,
		Uid:        testUid,
		Gid:        testGid,
		Recursive:  true,
		IgnoreFile: true,
	}
	err = SetPathOwnerGroup(param)
	convey.So(err, convey.ShouldBeNil)
	dirInfo, err := os.Stat(tmpDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(dirInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, testUid)
	convey.So(dirInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, testGid)
	fileInfo, err := os.Stat(filePath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, 0)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, 0)
}

func testSetSoftlinkFileOwnerGroup() {
	tmpDir, _, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	tgtPath := filepath.Join(os.TempDir(), "test_target")
	f, err := os.OpenFile(tgtPath, os.O_CREATE, Mode400)
	convey.So(err, convey.ShouldBeNil)
	defer f.Close()
	defer os.Remove(tgtPath)

	param := SetOwnerParam{
		Path:       tmpDir,
		Uid:        testUid,
		Gid:        testGid,
		Recursive:  true,
		IgnoreFile: false,
	}
	err = SetPathOwnerGroup(param)
	convey.So(err, convey.ShouldBeNil)
	fileInfo, err := os.Stat(tgtPath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, 0)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, 0)

	linkPath := filepath.Join(tmpDir, "test_link")
	err = os.Symlink(tgtPath, linkPath)
	err = SetPathOwnerGroup(param)
	convey.So(err, convey.ShouldNotBeNil)
}

func testSetFileOwnerGroupWithChecker() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	param := SetOwnerParam{
		Path:       tmpDir,
		Uid:        testUid,
		Gid:        testGid,
		Recursive:  true,
		IgnoreFile: false,
		CheckerParam: []FileChecker{
			NewFileModeChecker(true, 0, true, true),
		},
	}
	err = SetPathOwnerGroup(param)
	convey.So(err, convey.ShouldBeNil)
	dirInfo, err := os.Stat(tmpDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(dirInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, testUid)
	convey.So(dirInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, testGid)
	fileInfo, err := os.Stat(filePath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Uid, convey.ShouldEqual, testUid)
	convey.So(fileInfo.Sys().(*syscall.Stat_t).Gid, convey.ShouldEqual, testGid)
}

func TestSetPathPermission(t *testing.T) {
	convey.Convey("set path permission failed since mode not support", t, testSetPathInvalidMode)
	convey.Convey("set mode non-recursively success", t, testRecursivelySetFileMode)
	convey.Convey("set mode recursively success", t, testNonRecursivelySetFileMode)
	convey.Convey("set mode ignore file success", t, testIgnoreFileSetFileMode)
	convey.Convey("set soft link mode failed", t, testSetSoftlinkFileMode)
	convey.Convey("set mode with recursively mode checker failed", t, testSetFileModeWithChecker)
}

func testSetPathInvalidMode() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	err = SetPathPermission(filePath, mode777, false, false)
	convey.So(err, convey.ShouldResemble, errors.New("cannot set write permission for group or others"))
}

func testRecursivelySetFileMode() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	err = SetPathPermission(tmpDir, mode755, true, false)
	convey.So(err, convey.ShouldBeNil)
	dirInfo, err := os.Stat(tmpDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(dirInfo.Mode()&mode777, convey.ShouldEqual, mode755)
	fileInfo, err := os.Stat(filePath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Mode()&mode777, convey.ShouldEqual, mode755)
}

func testNonRecursivelySetFileMode() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	err = SetPathPermission(tmpDir, mode755, false, false)
	convey.So(err, convey.ShouldBeNil)
	dirInfo, err := os.Stat(tmpDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(dirInfo.Mode()&mode777, convey.ShouldEqual, mode755)
	fileInfo, err := os.Stat(filePath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Mode()&mode777, convey.ShouldEqual, Mode600)
}

func testIgnoreFileSetFileMode() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	err = SetPathPermission(tmpDir, mode755, true, true)
	convey.So(err, convey.ShouldBeNil)
	dirInfo, err := os.Stat(tmpDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(dirInfo.Mode()&mode777, convey.ShouldEqual, mode755)
	fileInfo, err := os.Stat(filePath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Mode()&mode777, convey.ShouldEqual, Mode600)
}

func testSetSoftlinkFileMode() {
	tmpDir, _, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	tgtPath := filepath.Join(os.TempDir(), "test_target")
	f, err := os.OpenFile(tgtPath, os.O_CREATE, Mode400)
	convey.So(err, convey.ShouldBeNil)
	defer f.Close()
	defer os.Remove(tgtPath)

	err = SetPathPermission(tmpDir, mode755, true, false)
	convey.So(err, convey.ShouldBeNil)
	fileInfo, err := os.Stat(tgtPath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Mode()&mode777, convey.ShouldEqual, Mode400)

	linkPath := filepath.Join(tmpDir, "test_link")
	err = os.Symlink(tgtPath, linkPath)
	err = SetPathPermission(tmpDir, mode755, true, false)
	convey.So(err, convey.ShouldNotBeNil)
}

func testSetFileModeWithChecker() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	checker := NewFileModeChecker(true, 0, true, true)
	err = SetPathPermission(tmpDir, mode755, true, false, checker)
	convey.So(err, convey.ShouldBeNil)
	dirInfo, err := os.Stat(tmpDir)
	convey.So(err, convey.ShouldBeNil)
	convey.So(dirInfo.Mode()&mode777, convey.ShouldEqual, fs.FileMode(Mode755))
	fileInfo, err := os.Stat(filePath)
	convey.So(err, convey.ShouldBeNil)
	convey.So(fileInfo.Mode()&mode777, convey.ShouldEqual, fs.FileMode(Mode755))
}
