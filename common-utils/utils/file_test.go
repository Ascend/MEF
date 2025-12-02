// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils provides the util func
package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestReadLimitBytes(t *testing.T) {
	convey.Convey("test ReadLimitBytes func", t, func() {
		convey.Convey("should return nil given empty string", func() {
			emptyString := ""
			const limitLength = 10
			res, err := ReadLimitBytes(emptyString, limitLength)
			convey.So(res, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("should not return nil given valid path", func() {
			const limitLength = 10
			res, err := ReadLimitBytes("./go.mod", limitLength)
			convey.So(res, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return nil given invalid limit length", func() {
			const limitLength = -1
			res, err := ReadLimitBytes("./go.mod", limitLength)
			convey.So(res, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "the limit length is not valid")
		})

		convey.Convey("should return nil when check path failed", func() {
			checkStub := gomonkey.ApplyFunc(CheckPath, func(path string) (string, error) {
				return "", errors.New("check failed")
			})
			defer checkStub.Reset()
			const limitLength = 10
			res, err := ReadLimitBytes("./go.mod", limitLength)
			convey.So(res, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "check failed")
		})

		convey.Convey("should return nil when read file failed", func() {
			var file *os.File
			checkStub := gomonkey.ApplyMethod(reflect.TypeOf(file), "Read",
				func(_ *os.File, _ []byte) (int, error) {
					return 0, errors.New("read file failed")
				})
			defer checkStub.Reset()
			const limitLength = 10
			res, err := ReadLimitBytes("./go.mod", limitLength)
			convey.So(res, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "read file failed: read file failed")
		})
	})
}

func TestLoadFile(t *testing.T) {
	convey.Convey("test LoadFile func", t, func() {
		convey.Convey("should return error given empty path", func() {
			res, err := LoadFile("")
			convey.So(res, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return nil given path not existing", func() {
			res, err := LoadFile("xxxx")
			convey.So(res, convey.ShouldBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should not return nil given valid path", func() {
			res, err := LoadFile("./go.mod")
			convey.So(res, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return nil given invalid path", func() {
			absStub := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return "", errors.New("the path is invalid")
			})
			defer absStub.Reset()
			res, err := LoadFile("./go.mod")
			convey.So(res, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "the filePath is invalid")
		})

		convey.Convey("should return nil when read file failed", func() {
			readStub := gomonkey.ApplyFunc(ReadLimitBytes, func(path string, limitLength int) ([]byte, error) {
				return nil, errors.New("read file failed")
			})
			defer readStub.Reset()
			res, err := LoadFile("./go.mod")
			convey.So(res, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "read file failed")
		})
	})
}

func TestCopyDir(t *testing.T) {
	convey.Convey("test CopyDir func", t, func() {
		convey.Convey("should return error given empty src path", func() {
			err := CopyDir("", "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return error given file src path", func() {
			err := CopyDir("./go.mod", "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return nil given dir src path", func() {
			err := CopyDir("../utils", "../utils_test")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("should return error given file dst path", func() {
			err := CopyDir("../utils", "../utils_test/file_test.go")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestCopyFile(t *testing.T) {
	convey.Convey("test CopyFile func", t, func() {
		convey.Convey("should return error given empty src file path", func() {
			err := CopyFile("", "../utils_test/file_test.go")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return error given empty dst path", func() {
			err := CopyFile("../utils_test/file_test.go", "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return error given dir scr path", func() {
			err := CopyFile("../utils", "../utils_test/file_test.go")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return error given dir dst path", func() {
			err := CopyFile("../utils/file_test.go", "../utils_test")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return nil given file scr and dst path", func() {
			err := CopyFile("../utils/file_test.go", "../utils_test/file_test.go")
			convey.So(err, convey.ShouldBeNil)
		})
	})
	if err := os.RemoveAll("../utils_test"); err != nil {
		fmt.Print("remove util_test file failed")
	}
}

func TestDeleteFile(t *testing.T) {
	convey.Convey("test DeleteFile func", t, func() {
		convey.Convey("should return nil given path does not exist", func() {
			err := DeleteFile("/xxx/xxxx")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("should return error given dir is softlink", func() {
			tmpDir, absPath, err := createTestFile(t, "test_file.txt")
			convey.So(err, convey.ShouldBeNil)
			defer removeTmpDir(t, tmpDir)
			linkPath := tmpDir + "/syslink"
			err = os.Symlink(filepath.Dir(tmpDir), linkPath)
			convey.So(err, convey.ShouldBeNil)
			linkFilePath := filepath.Join(linkPath, filepath.Base(absPath))
			err = DeleteFile(linkFilePath)
			convey.So(err.Error(), convey.ShouldEqual, "dir path check failed: can't support symlinks")
		})
		convey.Convey("should return nil", func() {
			tmpDir, absPath, err := createTestFile(t, "test_file.txt")
			convey.So(err, convey.ShouldBeNil)
			defer removeTmpDir(t, tmpDir)
			err = DeleteFile(absPath)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRenameFile(t *testing.T) {
	convey.Convey("test RenameFile func", t, func() {
		convey.Convey("should return error given src path does not exist", func() {
			err := RenameFile("/xxx/xxxx", "/xxx/xxxx")
			convey.So(err.Error(), convey.ShouldEqual, "rename file failed: src path does not exist")
		})
		convey.Convey("should return error given dst dir path does not exist", func() {
			realPath, err := filepath.Abs("./go.mod")
			convey.So(err, convey.ShouldBeNil)
			err = RenameFile(realPath, "/xxx/xxxx")
			convey.So(err.Error(), convey.ShouldEqual, "rename file failed: dst dir does not exist")
		})
		convey.Convey("should return error given src path is softlink", func() {
			tmpDir, _, err := createTestFile(t, "test_file.txt")
			convey.So(err, convey.ShouldBeNil)
			defer removeTmpDir(t, tmpDir)
			linkPath := tmpDir + "/syslink"
			err = os.Symlink(tmpDir, linkPath)
			convey.So(err, convey.ShouldBeNil)
			err = RenameFile(linkPath, "./xxx/xxxx")
			convey.So(err.Error(), convey.ShouldEqual, "check srcPath failed: can't support symlinks")
		})
		convey.Convey("should return error given dst dir is softlink", func() {
			tmpDir, absPath, err := createTestFile(t, "test_file.txt")
			convey.So(err, convey.ShouldBeNil)
			defer removeTmpDir(t, tmpDir)
			linkPath := tmpDir + "/syslink"
			err = os.Symlink(tmpDir, linkPath)
			convey.So(err, convey.ShouldBeNil)
			err = RenameFile(absPath, filepath.Join(linkPath, "test.txt"))
			convey.So(err.Error(), convey.ShouldEqual, "check dst Path failed: can't support symlinks")
		})
		convey.Convey("should return nil", func() {
			tmpDir, absPath, err := createTestFile(t, "test_file.txt")
			convey.So(err, convey.ShouldBeNil)
			defer removeTmpDir(t, tmpDir)
			err = RenameFile(absPath, filepath.Join(tmpDir, "test.txt"))
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCreateFile(t *testing.T) {
	convey.Convey("test CreateFile func", t, func() {
		convey.Convey("should return nil since file exists", func() {
			err := CreateFile("./go.mod", FileMode)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("should return error given path does not exist", func() {
			err := CreateFile("/xxx/xxxx", FileMode)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("should return error open file failed", func() {
			patch := gomonkey.ApplyFuncReturn(os.OpenFile, nil, errors.New("testErr"))
			defer patch.Reset()
			realPath, err := filepath.Abs("../utils_test/test.txt")
			convey.So(err, convey.ShouldBeNil)
			err = CreateFile(realPath, FileMode)
			convey.So(err.Error(), convey.ShouldEqual, "testErr")
		})
		convey.Convey("should return nil", func() {
			err := MakeSureDir("../utils_test/test.txt")
			convey.So(err, convey.ShouldBeNil)
			realPath, err := filepath.Abs("../utils_test/test.txt")
			err = CreateFile(realPath, FileMode)
			convey.So(err, convey.ShouldBeNil)
		})
		if err := os.RemoveAll("../utils_test"); err != nil {
			fmt.Print("remove util_test file failed")
		}
	})
}

func TestReadDir(t *testing.T) {
	convey.Convey("test ReadDir func", t, func() {
		convey.Convey("should return error given path does not exist", func() {
			_, err := ReadDir("./xxx/xxxx")
			convey.So(err, convey.ShouldEqual, os.ErrNotExist)
		})
		convey.Convey("should return error given path is not realpath", func() {
			_, err := ReadDir("./go.mod")
			convey.So(err.Error(), convey.ShouldEqual, fmt.Sprintf("the input path not equal the realpath"))
		})
		convey.Convey("should return error given path is not dir", func() {
			realPath, err := filepath.Abs("./go.mod")
			convey.So(err, convey.ShouldBeNil)
			_, err = ReadDir(realPath)
			convey.So(err.Error(), convey.ShouldEqual, fmt.Sprintf("path %s is not dir", realPath))
		})
		convey.Convey("should return nil", func() {
			realPath, err := filepath.Abs("./")
			_, err = ReadDir(realPath)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestReadFile(t *testing.T) {
	convey.Convey("test ReadFile func", t, func() {
		convey.Convey("should return error given empty path", func() {
			res, err := ReadFile("")
			convey.So(res, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "the file does not exist")
		})

		convey.Convey("should return nil given path not existing", func() {
			res, err := ReadFile("xxxx")
			convey.So(res, convey.ShouldBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "the file does not exist")
		})

		convey.Convey("should not return nil given valid path", func() {
			res, err := LoadFile("./go.mod")
			convey.So(res, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestWriteData(t *testing.T) {
	convey.Convey("test WriteData func", t, func() {
		convey.Convey("should return error given empty path", func() {
			err := WriteData("", nil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("should return err make sure dir failed", func() {
			patch := gomonkey.ApplyFuncReturn(MakeSureDir, errors.New("testErr"))
			defer patch.Reset()
			err := WriteData("/xxx/xxxx", nil)
			convey.So(err.Error(), convey.ShouldEqual, "testErr")
		})

		convey.Convey("should return err open file failed", func() {
			patch := gomonkey.ApplyFuncReturn(os.OpenFile, nil, errors.New("testErr"))
			defer patch.Reset()
			err := WriteData("/test.txt", nil)
			convey.So(err.Error(), convey.ShouldEqual, "testErr")
		})

		convey.Convey("should return nil", func() {
			path, err := filepath.Abs("./test.txt")
			convey.So(err, convey.ShouldBeNil)
			err = WriteData(path, nil)
			defer func() {
				err = DeleteFile(path)
				convey.So(err, convey.ShouldBeNil)
			}()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDeleteAllFileWithConfusion(t *testing.T) {
	convey.Convey("test DeleteAllFileWithConfusion func", t, func() {
		convey.Convey("should return nil given path not exists", func() {
			err := DeleteAllFileWithConfusion("/xxx/xxxx")
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return nil", func() {
			err := MakeSureDir("../utils_test/test.txt")
			convey.So(err, convey.ShouldBeNil)
			realPath, err := filepath.Abs("../utils_test/test.txt")
			convey.So(err, convey.ShouldBeNil)
			err = CreateFile(realPath, FileMode)
			convey.So(err, convey.ShouldBeNil)
			realPath, err = filepath.Abs("../utils_test")
			convey.So(err, convey.ShouldBeNil)
			err = DeleteAllFileWithConfusion(realPath)
			convey.So(err, convey.ShouldBeNil)
		})

		if err := os.RemoveAll("../utils_test"); err != nil {
			fmt.Print("remove util_test file failed")
		}
	})
}

func TestGetFileSha256(t *testing.T) {
	convey.Convey("test GetFileSha256 func", t, func() {
		convey.Convey("should return nil given valid path", func() {
			const (
				testString       = "123"
				testStringSha256 = "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"
				testFilePath     = "../test/test_sha256"
			)
			err := MakeSureDir(testFilePath)
			convey.So(err, convey.ShouldBeNil)
			err = os.WriteFile(testFilePath, []byte(testString), Mode600)
			convey.So(err, convey.ShouldBeNil)
			defer os.Remove(testFilePath)
			hash, err := GetFileSha256(testFilePath)
			convey.So(err, convey.ShouldBeNil)
			convey.So(hash, convey.ShouldEqual, testStringSha256)
			realPath, err := filepath.Abs(testFilePath)
			convey.So(err, convey.ShouldBeNil)
			err = DeleteAllFileWithConfusion(realPath)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
