// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func singleCheck(path string, checker FileChecker) error {
	file, err := os.OpenFile(path, os.O_RDONLY, Mode400)
	convey.So(err, convey.ShouldBeNil)
	defer CloseFile(file)
	err = checker.Check(file, path)
	return err
}

func TestCheckers(t *testing.T) {
	convey.Convey("test file link checker: ", t, testLinkChecker)
	convey.Convey("test file size checker: ", t, testFileSizeChecker)
	convey.Convey("test file mode checker: ", t, testFileModeChecker)
	convey.Convey("test file owner checker: ", t, testFileOwnerChecker)
	convey.Convey("test path checker: ", t, testFilePathChecker)
	convey.Convey("test is dir checker: ", t, testIsDirChecker)
}

func testLinkChecker() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("no link return nil", func() {
		checker := NewFileLinkChecker(true)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("has link return err", func() {
		checker := NewFileLinkChecker(true)
		linkPath := filepath.Join(tmpDir, "test_link")
		err = os.Symlink(filePath, linkPath)
		convey.So(err, convey.ShouldBeNil)
		err = singleCheck(linkPath, checker)
		convey.So(err, convey.ShouldResemble, errors.New("can't support symlinks"))
	})

	convey.Convey("check relative path return err", func() {
		checker := NewFileLinkChecker(false)
		testFileName := "test_file"
		f, err := os.OpenFile(testFileName, os.O_CREATE, Mode600)
		convey.So(err, convey.ShouldBeNil)
		defer f.Close()
		defer os.Remove(testFileName)
		err = singleCheck(testFileName, checker)
		convey.So(err, convey.ShouldResemble, errors.New("can't support symlinks"))
	})
}

func testFileSizeChecker() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("not exceeding limitation check success", func() {
		checker := NewFileSizeChecker(1)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("exceeding limitation check failed", func() {
		const (
			dataSize  = 2 * 1024 * 1024
			allowSize = 1 * 1024 * 1024
		)
		data := make([]byte, dataSize)
		err = os.WriteFile(filePath, data, Mode600)
		convey.So(err, convey.ShouldBeNil)

		checker := NewFileSizeChecker(allowSize)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldResemble, errors.New("file size exceeds 1.00 MB"))
	})
}

func testFileModeChecker() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("non-recursively mode check success", func() {
		checker := NewFileModeChecker(false, DefaultWriteFileMode, true, true)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("non-recursively mode check failed", func() {
		const umask400 fs.FileMode = 0400
		checker := NewFileModeChecker(false, umask400, true, true)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("path %s's file mode -rw------- unsupported", filePath))
	})

	convey.Convey("recursively mode check failed", func() {
		checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("path %s's file mode dtrwxrwxrwx unsupported", os.TempDir()))
	})

	convey.Convey("recursively mode check success", func() {
		const umask000 fs.FileMode = 0000
		checker := NewFileModeChecker(false, umask000, true, true)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testFileOwnerChecker() {
	convey.Convey("non-recursively owner checker: ", testNonRecursivelyFileOwnerChecker)
	convey.Convey("recursively owner checker: ", testRecursivelyFileOwnerChecker)
}

func testNonRecursivelyFileOwnerChecker() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("check success", func() {
		checker := NewFileOwnerChecker(false, false, 0, 0)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("uid check failed", func() {
		const testUid = 1024
		checker := NewFileOwnerChecker(false, false, testUid, 0)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("the owner of file [%s] [uid=0] is not supported", filePath))
	})

	convey.Convey("gid check failed", func() {
		const testGid = 1024
		checker := NewFileOwnerChecker(false, false, 0, testGid)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("the group of file [%s] [gid=0] is not supported", filePath))
	})
}

func testRecursivelyFileOwnerChecker() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("check success", func() {
		checker := NewFileOwnerChecker(true, false, 0, 0)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("uid check failed", func() {
		const testUid = 1024
		err = os.Chown(tmpDir, testUid, 0)
		convey.So(err, convey.ShouldBeNil)
		checker := NewFileOwnerChecker(true, false, 0, 0)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("the owner of file [%s] [uid=%d] is not supported", tmpDir, testUid))
	})

	convey.Convey("gid check failed", func() {
		const testGid = 1024
		err = os.Chown(tmpDir, 0, testGid)
		convey.So(err, convey.ShouldBeNil)
		checker := NewFileOwnerChecker(true, false, 0, 0)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("the group of file [%s] [gid=%d] is not supported", tmpDir, testGid))
	})

	convey.Convey("check allow current uid success", func() {
		err = os.Chown(tmpDir, testUid, 0)
		convey.So(err, convey.ShouldBeNil)
		err = os.Chown(filePath, testUid, 0)
		convey.So(err, convey.ShouldBeNil)
		p := gomonkey.ApplyFuncReturn(os.Geteuid, testUid).ApplyFuncReturn(os.Getegid, testGid)
		defer p.Reset()

		checker := NewFileOwnerChecker(true, true, 0, 0)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("check allow current gid success", func() {
		const testGid = 1024
		err = os.Chown(tmpDir, 0, testGid)
		convey.So(err, convey.ShouldBeNil)
		err = os.Chown(filePath, 0, testGid)
		convey.So(err, convey.ShouldBeNil)
		p := gomonkey.ApplyFuncReturn(os.Geteuid, testUid).ApplyFuncReturn(os.Getegid, testGid)
		defer p.Reset()
		checker := NewFileOwnerChecker(true, true, 0, 0)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testFilePathChecker() {
	tmpDir, _, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("path contains illegal character", func() {
		path := filepath.Join(tmpDir, "!123")
		f, err := os.OpenFile(path, os.O_CREATE, Mode600)
		convey.So(err, convey.ShouldBeNil)
		defer f.Close()

		checker := NewFilePathChecker()
		err = checker.Check(f, path)
		convey.So(err, convey.ShouldResemble, errors.New("path has unsupported character"))
	})

	convey.Convey("path contains ..", func() {
		checker := NewFilePathChecker()
		err = singleCheck(tmpDir+"/..", checker)
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(err.Error(), convey.ShouldContainSubstring,
			"the input path is not a valid absolute path")
	})
}

func testIsDirChecker() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("check dir failed", func() {
		checker := NewIsDirChecker(true)
		err = singleCheck(filePath, checker)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("path %s is not dir", filePath))
	})

	convey.Convey("check file failed", func() {
		checker := NewIsDirChecker(false)
		err = singleCheck(tmpDir, checker)
		convey.So(err, convey.ShouldResemble, fmt.Errorf("path %s is a dir", tmpDir))
	})
}
