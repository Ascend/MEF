// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MEF is licensed under Mulan PSL v2.
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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestIsDir(t *testing.T) {
	convey.Convey("test logger", t, func() {
		convey.Convey("test IsDir func", func() {
			res := IsDir("/tmp/")
			convey.So(res, convey.ShouldBeTrue)
			res = IsDir("/utils/")
			convey.So(res, convey.ShouldBeTrue)
			res = IsDir("")
			convey.So(res, convey.ShouldBeFalse)
		})
	})
}

func TestIsFile(t *testing.T) {
	convey.Convey("test IsFile func", t, func() {
		res := IsFile("/tmp/")
		convey.So(res, convey.ShouldBeFalse)
		res = IsFile("")
		convey.So(res, convey.ShouldBeFalse)
	})
}

func TestIsExist(t *testing.T) {
	convey.Convey("test IsExist func", t, func() {
		res := IsExist("/xxxx/")
		convey.So(res, convey.ShouldBeFalse)
	})
}

func TestIsLexist(t *testing.T) {
	convey.Convey("test IsLexist func", t, func() {
		res := IsLexist("/xxxx/")
		convey.So(res, convey.ShouldBeFalse)
	})
}

func TestReadLink(t *testing.T) {
	convey.Convey("test ReadLink func", t, func() {
		convey.Convey("test success", func() {
			tmpDir, _, err := createTestFile(t, "test_file.txt")
			convey.So(err, convey.ShouldBeNil)
			defer removeTmpDir(t, tmpDir)
			linkPath := tmpDir + "/syslink"
			err = os.Symlink(filepath.Dir(tmpDir), linkPath)
			convey.So(err, convey.ShouldBeNil)
			_, err = ReadLink(linkPath)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test failed", func() {
			_, err := ReadLink("/xx/xxx")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestEvalSymlinks(t *testing.T) {
	convey.Convey("test EvalSymlinks func", t, func() {
		convey.Convey("test success", func() {
			_, err := EvalSymlinks("./go.mod")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("test failed", func() {
			_, err := EvalSymlinks("/xx/xxx")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestCheckOriginPathOk(t *testing.T) {
	convey.Convey("test CheckPath func", t, func() {
		convey.Convey("should return itself given empty string", func() {
			res, err := CheckOriginPath("")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("should return resolve path for given abs path", func() {
			absPath, err := filepath.Abs("./go.mod")
			res, err := CheckOriginPath(absPath)
			convey.So(res, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("should return resolve path for given abs dir with a suffix of '/' ", func() {
			absPath, err := filepath.Abs("./utils")
			res, err := CheckOriginPath(absPath + "/")
			convey.So(res, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func TestCheckOriginPathFailed(t *testing.T) {
	convey.Convey("test CheckPath func", t, func() {
		convey.Convey("should return error given not exist path", func() {
			res, err := CheckOriginPath("xxxxxxx")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "file does not exist")
		})

		convey.Convey("should return error for given relative path", func() {
			res, err := CheckOriginPath("./go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual,
				"the input path not equal the realpath")
		})

		convey.Convey("should return err when get abs path failed", func() {
			absStub := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return "", errors.New("abs failed")
			})
			defer absStub.Reset()
			res, err := CheckOriginPath("./go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "get the absolute path failed")
		})

		convey.Convey("should return err when get eval symbol link failed", func() {
			symStub := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", errors.New("symlinks path failed")
			})
			defer symStub.Reset()
			res, err := CheckOriginPath("./go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "get the symlinks path failed")
		})

		convey.Convey("should return err given symbol link", func() {
			symStub := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "xxx", nil
			})
			defer symStub.Reset()
			res, err := CheckOriginPath("./go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "can't support symlinks")
		})

		convey.Convey("should return err given switching flag ", func() {
			res, err := CheckOriginPath("/root/../root")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "the input path contains unsupported flag for parent directory")
		})
	})
}

func TestCheckPath(t *testing.T) {
	convey.Convey("test CheckPath func", t, func() {
		convey.Convey("should return itself given empty string", func() {
			res, err := CheckPath("")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return error given not exist path", func() {
			res, err := CheckPath("xxxxxxx")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "file does not exist")
		})

		convey.Convey("should return resolve path given normal path", func() {
			res, err := CheckPath("./go.mod")
			convey.So(res, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return err when get abs path failed", func() {
			absStub := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return "", errors.New("abs failed")
			})
			defer absStub.Reset()
			res, err := CheckPath("./go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "get the absolute path failed")
		})

		convey.Convey("should return err when get eval symbol link failed", func() {
			symStub := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", errors.New("symlinks path failed")
			})
			defer symStub.Reset()
			res, err := CheckPath("./go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "get the symlinks path failed")
		})

		convey.Convey("should return err given symbol link", func() {
			symStub := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "xxx", nil
			})
			defer symStub.Reset()
			res, err := CheckPath("./go.mod")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err.Error(), convey.ShouldEqual, "can't support symlinks")
		})

	})
}

func TestMakeSureDir(t *testing.T) {
	convey.Convey("test MakeSureDir func", t, func() {
		convey.Convey("normal situation, no err returned", func() {
			err := MakeSureDir("./testdata/tmp/test")
			convey.So(err, convey.ShouldEqual, nil)
		})
		convey.Convey("abnormal situation,err returned", func() {
			mock := gomonkey.ApplyFunc(os.MkdirAll, func(name string, perm os.FileMode) error {
				return fmt.Errorf("error")
			})
			defer mock.Reset()
			err := MakeSureDir("./xxxx/xxx")
			convey.So(err.Error(), convey.ShouldEqual, "create directory failed")
		})
	})
}

func TestCreateDir(t *testing.T) {
	convey.Convey("test CreateDir func", t, func() {
		convey.Convey("normal situation, no err returned", func() {
			err := CreateDir("./testdata/tmp/test", dirMode)
			convey.So(err, convey.ShouldEqual, nil)
			err = CreateDir("./testdata/tmp/test", dirMode)
			convey.So(err, convey.ShouldEqual, nil)
		})

		convey.Convey("abnormal situation,mkdir failed", func() {
			mock := gomonkey.ApplyFunc(os.MkdirAll, func(name string, perm os.FileMode) error {
				return fmt.Errorf("error")
			})
			defer mock.Reset()
			err := CreateDir("./xxxx/xxx", dirMode)
			convey.So(err.Error(), convey.ShouldEqual, "error")
		})
	})
}

func TestGetDriverLibPath(t *testing.T) {
	convey.Convey("test GetDriverLibPath func", t, func() {
		convey.Convey("should return itself given empty string", func() {
			err := os.Setenv(ldLibPath, "")
			convey.So(err, convey.ShouldBeNil)
			res, err := GetDriverLibPath("")
			convey.So(res, convey.ShouldBeEmpty)
			convey.So(err, convey.ShouldBeError)
		})

		convey.Convey("should return path when getLibFromEnv succeed", func() {
			envStub := gomonkey.ApplyFunc(getLibFromEnv, func(libraryName string) (string, error) {
				return "/test", nil
			})
			defer envStub.Reset()
			res, err := GetDriverLibPath("")
			convey.So(res, convey.ShouldEqual, "/test")
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("should return path when getLibFromEnv failed but getLibFromLdCmd succeed", func() {
			envStub := gomonkey.ApplyFunc(getLibFromEnv, func(libraryName string) (string, error) {
				return "", errors.New("failed")
			})
			defer envStub.Reset()
			cmdStub := gomonkey.ApplyFunc(getLibFromLdCmd, func(libraryName string) (string, error) {
				return "/test", nil
			})
			defer cmdStub.Reset()
			res, err := GetDriverLibPath("")
			convey.So(res, convey.ShouldEqual, "/test")
			convey.So(err, convey.ShouldBeNil)
		})

	})
}

func TestCheckAbsPath(t *testing.T) {
	convey.Convey("test checkAbsPath", t, func() {
		convey.Convey("should return abs path given valid path", func() {
			absPath, err := filepath.Abs("./go.mod")
			convey.So(err, convey.ShouldBeNil)
			patch := gomonkey.ApplyFuncReturn(CheckOwnerAndPermission, absPath, nil)
			defer patch.Reset()
			res, err := checkAbsPath("./go.mod")
			convey.So(res, convey.ShouldNotBeEmpty)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
