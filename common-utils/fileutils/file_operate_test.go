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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"huawei.com/mindx/common/rand"
)

func TestIsEmptyDir(t *testing.T) {
	convey.Convey("check dir is empty", t, func() {
		testDir := "test_dir"
		err := os.Mkdir(testDir, os.ModePerm)
		convey.So(err, convey.ShouldEqual, nil)
		defer os.RemoveAll(testDir)

		ok, err := IsEmptyDir(testDir)
		convey.So(ok, convey.ShouldEqual, true)

		file, err := os.OpenFile(filepath.Join(testDir, "test_file.txt"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, Mode600)
		convey.So(err, convey.ShouldEqual, nil)
		defer file.Close()
		ok, err = IsEmptyDir(testDir)
		convey.So(ok, convey.ShouldEqual, false)
	})
}

func TestEvalSymlinks(t *testing.T) {
	convey.Convey("eval symlinks success", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		linkPath := filepath.Join(tmpDir, "test_link")
		err = os.Symlink(filePath, linkPath)
		realPath, err := EvalSymlinks(linkPath)
		convey.So(err, convey.ShouldBeNil)
		convey.So(realPath, convey.ShouldEqual, filePath)
	})
}

func TestReadLink(t *testing.T) {
	convey.Convey("read link success", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		linkPath := filepath.Join(tmpDir, "test_link")
		err = os.Symlink(filePath, linkPath)
		realPath, err := ReadLink(linkPath)
		convey.So(err, convey.ShouldBeNil)
		convey.So(realPath, convey.ShouldEqual, filePath)
	})
}

func TestReadLimitBytes(t *testing.T) {
	convey.Convey("test Read Limit Bytes: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)
		testData := []byte{1, 0, 0}

		convey.Convey("read 1 byte success", func() {
			err = os.WriteFile(filePath, testData, Mode600)
			convey.So(err, convey.ShouldBeNil)
			content, err := ReadLimitBytes(filePath, 1)
			convey.So(err.Error(), convey.ShouldContainSubstring, "but the file has more content")
			convey.So(content, convey.ShouldResemble, []byte{1})
		})

		convey.Convey("read soft link failed", func() {
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(filePath, linkPath)
			convey.So(err, convey.ShouldBeNil)
			_, err = ReadLimitBytes(linkPath, 1)
			convey.So(err, convey.ShouldResemble, errors.New("file check failed: can't support symlinks"))
		})

		convey.Convey("read soft link failed since recursively mode checker is set", func() {
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			_, err = ReadLimitBytes(filePath, 1, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("file check failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestLoadFile(t *testing.T) {
	convey.Convey("test Load File: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)
		testData := []byte("testdata")

		convey.Convey("failed since file does not exist", func() {
			_, err = LoadFile(filepath.Join(tmpDir, "not_exist"))
			convey.So(err, convey.ShouldResemble, errors.New("path does not exist"))
		})

		convey.Convey("Read success with path checker", func() {
			err = os.WriteFile(filePath, testData, Mode600)
			convey.So(err, convey.ShouldBeNil)
			checker := NewFilePathChecker()
			content, err := LoadFile(filePath, checker)
			convey.So(err, convey.ShouldBeNil)
			convey.So(content, convey.ShouldResemble, testData)
		})
	})
}

func TestMakeSureDir(t *testing.T) {
	convey.Convey("test MakeSure Dir: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		convey.Convey("success since file exists", func() {
			tgtFile := filepath.Join(tmpDir, "not_exist")
			err = MakeSureDir(tgtFile)
			convey.So(err, convey.ShouldBeNil)
			convey.So(IsLexist(tgtFile), convey.ShouldBeFalse)
		})

		convey.Convey("success create dir", func() {
			tgtDir := filepath.Join(tmpDir, "tmp_dir")
			tgtFile := filepath.Join(tgtDir, "tmp_file")
			err = MakeSureDir(tgtFile)
			convey.So(err, convey.ShouldBeNil)
			convey.So(IsLexist(tgtDir), convey.ShouldBeTrue)
			convey.So(IsLexist(tgtFile), convey.ShouldBeFalse)
		})

		convey.Convey("failed since exists dir is soft link", func() {
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(filePath, linkPath)
			convey.So(err, convey.ShouldBeNil)
			tgtFile := filepath.Join(linkPath, "test_dir", "test_file")
			err = MakeSureDir(tgtFile)
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
		})

		convey.Convey("write failed since recursively mode checker is set", func() {
			tgtDir := filepath.Join(tmpDir, "test_dir")
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			err = MakeSureDir(tgtDir, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestCreateDir(t *testing.T) {
	convey.Convey("test Create Dir: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		convey.Convey("success since file exists", func() {
			err = CreateDir(tmpDir, Mode700)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("create dir success", func() {
			tgtDir := filepath.Join(tmpDir, "test_dir")
			err = CreateDir(tgtDir, Mode700)
			convey.So(err, convey.ShouldBeNil)
			convey.So(IsExist(tgtDir), convey.ShouldBeTrue)
			fileInfo, err := os.Stat(tgtDir)
			convey.So(err, convey.ShouldBeNil)
			convey.So(fileInfo.IsDir(), convey.ShouldBeTrue)
			convey.So(fileInfo.Mode()&Mode700, convey.ShouldEqual, Mode700)
		})

		convey.Convey("failed since exists dir is soft link", func() {
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(filePath, linkPath)
			convey.So(err, convey.ShouldBeNil)
			tgtDir := filepath.Join(linkPath, "test_dir")
			err = CreateDir(tgtDir, Mode700)
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
		})

		convey.Convey("write failed since recursively mode checker is set", func() {
			tgtDir := filepath.Join(tmpDir, "test_dir")
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			err = CreateDir(tgtDir, Mode700, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestWriteData(t *testing.T) {
	convey.Convey("test Write Data: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)
		testData := []byte("testdata")

		convey.Convey("success", func() {
			tgtPath := filepath.Join(tmpDir, "test_tgt")
			err = WriteData(tgtPath, testData)
			convey.So(err, convey.ShouldBeNil)
			content, err := os.ReadFile(tgtPath)
			convey.So(err, convey.ShouldBeNil)
			convey.So(content, convey.ShouldResemble, testData)
		})

		convey.Convey("write failed since tgt is soft link", func() {
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(filePath, linkPath)
			convey.So(err, convey.ShouldBeNil)
			err = WriteData(linkPath, testData)
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
			content, err := os.ReadFile(filePath)
			convey.So(err, convey.ShouldBeNil)
			convey.So(content, convey.ShouldResemble, []byte{})
		})

		convey.Convey("write failed since recursively mode checker is set", func() {
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			err = WriteData(filePath, testData, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestDeleteFile(t *testing.T) {
	convey.Convey("test Delete File： ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		convey.Convey("delete success", func() {
			err = DeleteFile(filePath)
			convey.So(err, convey.ShouldBeNil)
			convey.So(IsExist(filePath), convey.ShouldBeFalse)
		})

		convey.Convey("delete success since file does not exist", func() {
			err = DeleteFile(filepath.Join(tmpDir, "not_exist"))
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("delete failed since dir is not empty", func() {
			err = DeleteFile(tmpDir)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("delete failed since dir is soft link", func() {
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(tmpDir, linkPath)
			convey.So(err, convey.ShouldBeNil)
			err = DeleteFile(filepath.Join(linkPath, filepath.Base(filePath)))
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
		})

		convey.Convey("delete failed since recursively checker is set", func() {
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			err = DeleteFile(filePath, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestReadDir(t *testing.T) {
	convey.Convey("test Read Dir: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		convey.Convey("read success", func() {
			f, _, err := ReadDir(tmpDir)
			convey.So(err, convey.ShouldBeNil)
			defer CloseFile(f)
		})

		convey.Convey("read failed since param is not dir", func() {
			_, _, err := ReadDir(filePath)
			convey.So(err, convey.ShouldResemble, fmt.Errorf("check file failed: path %s is not dir", filePath))
		})

		convey.Convey("read failed since param is soft link", func() {
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(tmpDir, linkPath)
			convey.So(err, convey.ShouldBeNil)
			_, _, err = ReadDir(linkPath)
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
		})

		convey.Convey("read failed since recursively mode check is set", func() {
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			_, _, err = ReadDir(tmpDir, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestCreateFile(t *testing.T) {
	convey.Convey("test Create File: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		convey.Convey("return nil since file exists", func() {
			err = CreateFile(filePath, Mode600)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("return nil since create success", func() {
			tgtPath := filepath.Join(tmpDir, "test_tgt")
			err = CreateFile(tgtPath, Mode600)
			convey.So(err, convey.ShouldBeNil)
			convey.So(IsLexist(tgtPath), convey.ShouldBeTrue)
			fileInfo, err := os.Stat(tgtPath)
			convey.So(err, convey.ShouldBeNil)
			mode := fileInfo.Mode()
			convey.So(mode, convey.ShouldEqual, Mode600)
		})

		convey.Convey("dir is soft link", func() {
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(tmpDir, linkPath)
			convey.So(err, convey.ShouldBeNil)
			err = CreateFile(filepath.Join(linkPath, "test_tgt"), Mode600)
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
		})

		convey.Convey("failed since set recursively mode checker", func() {
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			tgtPath := filepath.Join(tmpDir, "test_tgt")
			err = CreateFile(tgtPath, Mode600, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestRenameFile(t *testing.T) {
	convey.Convey("test Rename File: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		convey.Convey("rename file success", func() {
			tgtPath := filepath.Join(tmpDir, "test_tgt")
			err = RenameFile(filePath, tgtPath)
			convey.So(err, convey.ShouldBeNil)
			convey.So(IsExist(tgtPath), convey.ShouldBeTrue)
			convey.So(IsExist(filePath), convey.ShouldBeFalse)
		})

		convey.Convey("rename file failed: src is soft link", func() {
			tgtPath := filepath.Join(tmpDir, "test_tgt")
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(filePath, linkPath)
			convey.So(err, convey.ShouldBeNil)
			err = RenameFile(linkPath, tgtPath)
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
		})

		convey.Convey("rename file failed: set recursively mode checker", func() {
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			tgtPath := filepath.Join(tmpDir, "test_tgt")
			err = RenameFile(filePath, tgtPath, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestCopyFile(t *testing.T) {
	convey.Convey("test Copy File: ", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(tmpDir)

		convey.Convey("copy file success", func() {
			tgtPath := filepath.Join(tmpDir, "test_tgt")
			err = CopyFile(filePath, tgtPath)
			convey.So(err, convey.ShouldBeNil)
			convey.So(IsExist(tgtPath), convey.ShouldBeTrue)
			convey.So(IsExist(filePath), convey.ShouldBeTrue)
		})

		convey.Convey("copy file failed: src is soft link", func() {
			srcDir := filepath.Join(tmpDir, "test_src")
			linkPath := filepath.Join(srcDir, "test_link")
			err = CreateDir(srcDir, Mode700)
			convey.So(err, convey.ShouldBeNil)
			err = os.Symlink(filePath, linkPath)
			convey.So(err, convey.ShouldBeNil)
			err = CopyFile(linkPath, filepath.Join(tmpDir, "test_dst"))
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
		})

		convey.Convey("copy file failed: dst is dir", func() {
			err = CopyFile(filePath, tmpDir)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: open %s: is a directory", tmpDir))
		})

		convey.Convey("copy file failed: dst is soft link", func() {
			linkPath := filepath.Join(tmpDir, "test_link")
			err = os.Symlink(filePath, linkPath)
			convey.So(err, convey.ShouldBeNil)
			err = CopyFile(filePath, linkPath)
			convey.So(err, convey.ShouldResemble, errors.New("check file failed: can't support symlinks"))
		})

		convey.Convey("copy file failed: with recursively mode checker", func() {
			checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
			tgtPath := filepath.Join(tmpDir, "test_tgt")
			err = CopyFile(filePath, tgtPath, checker)
			convey.So(err, convey.ShouldResemble,
				fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
		})
	})
}

func TestCopyDir(t *testing.T) {
	convey.Convey("test CopyDir func: normal situation", t, testNormalCopyDir)
	convey.Convey("test CopyDir func: set checker", t, testSetCheckerCopyDir)
}

func testNormalCopyDir() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("copy dir success", func() {
		srcDir := filepath.Join(tmpDir, "test_src")
		dstDir := filepath.Join(tmpDir, "test_tgt")
		err = CreateDir(srcDir, Mode700)
		convey.So(err, convey.ShouldBeNil)
		err = RenameFile(filePath, filepath.Join(srcDir, "test_file.txt"))
		convey.So(err, convey.ShouldBeNil)
		err = CopyDir(srcDir, dstDir)
		convey.So(err, convey.ShouldBeNil)
		convey.So(IsExist(srcDir), convey.ShouldBeTrue)
		convey.So(IsExist(dstDir), convey.ShouldBeTrue)
		convey.So(IsExist(filepath.Join(srcDir, "test_file.txt")), convey.ShouldBeTrue)
		convey.So(IsExist(filepath.Join(dstDir, "test_file.txt")), convey.ShouldBeTrue)
	})

	convey.Convey("copy dir failed: copy to subDir", func() {
		err = CopyDir(os.TempDir(), tmpDir)
		convey.So(err, convey.ShouldResemble,
			errors.New("the destination directory is a subdirectory of the source directory"))
	})

	convey.Convey("copy dir failed: src contains softlink", func() {
		srcDir := filepath.Join(tmpDir, "test_src")
		linkPath := filepath.Join(srcDir, "test_link")
		dstDir := filepath.Join(tmpDir, "test_tgt")
		err = CreateDir(srcDir, Mode700)
		convey.So(err, convey.ShouldBeNil)
		err = os.Symlink(filePath, linkPath)
		convey.So(err, convey.ShouldBeNil)
		err := CopyDir(linkPath, dstDir)
		convey.So(err.Error(), convey.ShouldContainSubstring, "is not dir")
	})
}

func testSetCheckerCopyDir() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("copy dir failed: recursively check mode", func() {
		checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
		srcDir := filepath.Join(tmpDir, "test_src")
		dstDir := filepath.Join(tmpDir, "test_tgt")
		err = CreateDir(srcDir, Mode700)
		convey.So(err, convey.ShouldBeNil)
		err = RenameFile(filePath, filepath.Join(srcDir, "test_file.txt"))
		convey.So(err, convey.ShouldBeNil)
		err = CopyDir(srcDir, dstDir, checker)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file failed: path %s's file mode drwxrwxrwx unsupported", os.TempDir()))
	})
}

func TestGetFileSha256(t *testing.T) {
	convey.Convey("test GetFileSha256 func", t, func() {
		tmpDir, filePath, err := createTestFile("test_file.txt")
		if err != nil {
			t.Fatalf("create tmp dir failed: %s", err.Error())
		}
		defer os.RemoveAll(tmpDir)

		convey.Convey("should return nil given valid path", func() {
			const (
				testString       = "123"
				testStringSha256 = "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3"
			)

			err = os.WriteFile(filePath, []byte(testString), Mode600)
			convey.So(err, convey.ShouldBeNil)
			hash, err := GetFileSha256(filePath)
			convey.So(err, convey.ShouldBeNil)
			convey.So(hash, convey.ShouldEqual, testStringSha256)
		})
	})
}

func TestDeleteAllFileWithConfusion(t *testing.T) {
	convey.Convey("test DeleteAllFileWithConfusion: default checker", t, testDefaultDeleteWithConfusion)
	convey.Convey("test DeleteAllFileWithConfusion non default checker", t, testSetCheckerDeleteWithConfusion)
}

func testDefaultDeleteWithConfusion() {
	tmpDir, filePath, err := createTestFile("test_file.key")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("should return nil given path not exists", func() {
		err := DeleteAllFileWithConfusion("/xxx/xxxx")
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("should return nil", func() {
		err = DeleteAllFileWithConfusion(tmpDir)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("should return err check symlink failed", func() {
		testSymlinkDir := filepath.Join(filepath.Dir(tmpDir), "test_symlink")
		err = os.Symlink(tmpDir, testSymlinkDir)
		convey.So(err, convey.ShouldBeNil)
		defer os.RemoveAll(testSymlinkDir)
		testFilePath := filepath.Join(testSymlinkDir, "test_file.key")
		err = DeleteAllFileWithConfusion(testFilePath)
		convey.So(err, convey.ShouldResemble, errors.New("check path failed: can't support symlinks"))
	})

	convey.Convey("should return err once Write data failed", func() {
		p := gomonkey.ApplyFuncReturn(rand.Read, 0, testErr)
		defer p.Reset()

		err = DeleteAllFileWithConfusion(tmpDir)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("confusion path %s failed: get random words failed", tmpDir))
	})

	convey.Convey("should return nil once Write data will fail but no key file in dir", func() {
		p := gomonkey.ApplyFuncReturn(rand.Read, 0, testErr)
		defer p.Reset()

		err = RenameFile(filePath, filepath.Join(tmpDir, "test_file"))
		convey.So(err, convey.ShouldBeNil)
		err = DeleteAllFileWithConfusion(tmpDir)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testSetCheckerDeleteWithConfusion() {
	tmpDir, filePath, err := createTestFile("test_file.key")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("should return err once check whole path mode", func() {
		checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
		err = DeleteAllFileWithConfusion(tmpDir, checker)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("confusion path %s failed: confusion file with 0 failed: check file failed: "+
				"path %s's file mode drwxrwxrwx unsupported", tmpDir, os.TempDir()))
	})

	convey.Convey("should return nil once check whole path but no key file", func() {
		checker := NewFileModeChecker(true, DefaultWriteFileMode, true, true)
		err = RenameFile(filePath, filepath.Join(tmpDir, "test_file"))
		convey.So(err, convey.ShouldBeNil)
		err = DeleteAllFileWithConfusion(tmpDir, checker)
		convey.So(err, convey.ShouldBeNil)
	})
}
