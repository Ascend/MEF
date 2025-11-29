//  Copyright(c) 2022. Huawei Technologies Co.,Ltd.  All rights reserved.

package fileutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

var (
	tmpFileCount = 0
	testErr      = errors.New("test error")
)

func TestCheckOwnerAndPermission(t *testing.T) {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	if err != nil {
		t.Fatalf("create tmp dir failed: %s", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	convey.Convey("Check Owner And Permission: user not correct", t, func() {
		const (
			mode000    = 0000
			notUserUid = 1024
		)
		_, err = CheckOwnerAndPermission(os.TempDir(), mode000, uint32(notUserUid))
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: the owner of file [%s] [uid=%d] is not supported",
				os.TempDir(), os.TempDir(), 0))
	})

	convey.Convey("Check Owner And Permission: right not correct", t, func() {
		_, err = CheckOwnerAndPermission(os.TempDir(), DefaultWriteFileMode, 0)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file /tmp failed: path /tmp's file mode drwxrwxrwx unsupported"))
	})

	convey.Convey("Check Owner And Permission: success", t, func() {
		_, err = CheckOwnerAndPermission(filePath, DefaultWriteFileMode, 0)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestCheckOriginPath(t *testing.T) {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	if err != nil {
		t.Fatalf("create tmp dir failed: %s", err.Error())
	}
	defer os.RemoveAll(tmpDir)

	convey.Convey("Check Origin Path: has softlink", t, func() {
		linkPath := filepath.Join(tmpDir, "syslink")
		err = os.Symlink(filePath, linkPath)
		convey.So(err, convey.ShouldBeNil)
		defer os.Remove(linkPath)

		_, err = CheckOriginPath(linkPath)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: can't support symlinks", linkPath))
	})

	convey.Convey("Check Origin Path: has relative path", t, func() {
		testFileName := "test_file"
		f, err := os.OpenFile(testFileName, os.O_CREATE, Mode600)
		convey.So(err, convey.ShouldBeNil)
		defer f.Close()
		defer os.Remove(testFileName)

		_, err = CheckOriginPath(testFileName)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: can't support symlinks", testFileName))
	})

	convey.Convey("Check Origin Path: check success", t, func() {
		cleanPath, err := CheckOriginPath(filePath)
		convey.So(err, convey.ShouldBeNil)
		convey.So(cleanPath, convey.ShouldEqual, filePath)
	})
}

func TestRealFileCheck(t *testing.T) {
	convey.Convey("Real File check: normal check", t, testRealFileCheckNormal)
	convey.Convey("Real File check: not check parent", t, testRealFileCheckNoParent)
	convey.Convey("Real File check: check parent", t, testRealFileCheckParent)
}

func testRealFileCheckNormal() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("Path contains illegal characters", func() {
		_, err = RealFileCheck(fmt.Sprintf("%s/..", tmpDir), false, true, 1)
		convey.So(err.Error(), convey.ShouldContainSubstring, "the input path is not a valid absolute path")
	})
	convey.Convey("not a file", func() {
		_, err = RealFileCheck(tmpDir, false, true, 1)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: path %s is a dir", tmpDir, tmpDir))
	})

	convey.Convey("oversize", func() {
		const dataSize = 2 * 1024 * 1024
		data := make([]byte, dataSize)
		err = os.WriteFile(filePath, data, Mode600)
		convey.So(err, convey.ShouldBeNil)

		_, err = RealFileCheck(filePath, false, true, 1)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: file size exceeds 1.00 MB", filePath))
	})

	convey.Convey("allow link", func() {
		linkPath := filepath.Join(os.TempDir(), "syslink")
		err = os.Symlink(filePath, linkPath)
		convey.So(err, convey.ShouldBeNil)
		defer os.Remove(linkPath)

		_, err = RealFileCheck(linkPath, false, true, 1)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("file is softlink", func() {
		linkPath := filepath.Join(tmpDir, "syslink")
		err = os.Symlink(filePath, linkPath)
		convey.So(err, convey.ShouldBeNil)
		defer os.Remove(linkPath)

		_, err = RealFileCheck(linkPath, false, false, 1)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: can't support symlinks", linkPath))
	})

	convey.Convey("check success", func() {
		_, err = RealFileCheck(filePath, false, true, 1)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testRealFileCheckNoParent() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("file mode check failed", func() {
		_, err = RealFileCheck(os.TempDir(), false, true, 1)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file /tmp failed: path /tmp's file mode drwxrwxrwx unsupported"))
	})

	convey.Convey("file owner check failed", func() {
		const testUid = 1024
		err = os.Chown(filePath, testUid, testUid)
		convey.So(err, convey.ShouldBeNil)
		defer os.Chown(filePath, 0, 0)

		_, err = RealFileCheck(filePath, false, true, 1)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: the owner of file [%s] [uid=%d] is not supported",
				filePath, filePath, testUid))
	})
}

func testRealFileCheckParent() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("file mode check failed", func() {
		_, err = RealFileCheck(filePath, true, true, 1)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: path /tmp's file mode drwxrwxrwx unsupported", filePath))
	})

	convey.Convey("file owner check failed", func() {
		const testUid = 1024
		convey.So(err, convey.ShouldBeNil)
		err = os.Chown(tmpDir, testUid, testUid)
		convey.So(err, convey.ShouldBeNil)
		var p = gomonkey.ApplyMethodReturn(&FileModeChecker{}, "Check", nil)
		defer p.Reset()

		_, err = RealFileCheck(filePath, true, true, 1)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: the owner of file [%s] [uid=%d] is not supported",
				filePath, tmpDir, testUid))
	})
}

func TestRealDirCheck(t *testing.T) {
	convey.Convey("Real Dir check: normal check", t, testRealDirCheckNormal)
	convey.Convey("Real Dir check: not check parent", t, testRealDirCheckNoParent)
	convey.Convey("Real Dir check: check parent", t, testRealDirCheckParent)
}

func testRealDirCheckNormal() {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("Path contains illegal characters", func() {
		_, err = RealDirCheck(fmt.Sprintf("%s/..", tmpDir), false, true)
		convey.So(err.Error(), convey.ShouldContainSubstring,
			"the input path is not a valid absolute path")
	})
	convey.Convey("not a dir", func() {
		_, err = RealDirCheck(filePath, false, true)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: path %s is not dir", filePath, filePath))
	})

	convey.Convey("allow link", func() {
		linkPath := filepath.Join(tmpDir, "syslink")
		err = os.Symlink(tmpDir, linkPath)
		convey.So(err, convey.ShouldBeNil)
		defer os.Remove(linkPath)

		_, err = RealDirCheck(linkPath, false, true)
		convey.So(err, convey.ShouldBeNil)
	})

	convey.Convey("file is softlink", func() {
		linkPath := filepath.Join(tmpDir, "syslink")
		err = os.Symlink(tmpDir, linkPath)
		convey.So(err, convey.ShouldBeNil)
		defer os.Remove(linkPath)

		_, err = RealDirCheck(linkPath, false, false)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: can't support symlinks", linkPath))
	})

	convey.Convey("check success", func() {
		_, err = RealDirCheck(tmpDir, false, true)
		convey.So(err, convey.ShouldBeNil)
	})
}

func testRealDirCheckNoParent() {
	tmpDir, _, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("file mode check failed", func() {
		_, err = RealDirCheck(os.TempDir(), false, true)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file /tmp failed: path /tmp's file mode drwxrwxrwx unsupported"))
	})

	convey.Convey("file owner check failed", func() {
		const testUid = 1024
		err = os.Chown(tmpDir, testUid, testUid)
		convey.So(err, convey.ShouldBeNil)
		defer os.Chown(tmpDir, 0, 0)

		_, err = RealDirCheck(tmpDir, false, true)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: the owner of file [%s] [uid=%d] is not supported",
				tmpDir, tmpDir, testUid))
	})
}

func testRealDirCheckParent() {
	tmpDir, _, err := createTestFile("test_file.txt")
	convey.So(err, convey.ShouldBeNil)
	defer os.RemoveAll(tmpDir)

	convey.Convey("file mode check failed", func() {
		_, err = RealDirCheck(tmpDir, true, true)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: path /tmp's file mode drwxrwxrwx unsupported", tmpDir))
	})

	convey.Convey("file owner check failed", func() {
		const testUid = 1024
		tmpTestDir := tmpDir + "/test_dir"
		err = os.Mkdir(tmpTestDir, Mode700)
		convey.So(err, convey.ShouldBeNil)
		err = os.Chown(tmpDir, testUid, testUid)
		convey.So(err, convey.ShouldBeNil)
		var p = gomonkey.ApplyMethodReturn(&FileModeChecker{}, "Check", nil)
		defer p.Reset()

		_, err = RealDirCheck(tmpTestDir, true, true)
		convey.So(err, convey.ShouldResemble,
			fmt.Errorf("check file %s failed: the owner of file [%s] [uid=%d] is not supported",
				tmpTestDir, tmpDir, testUid))
	})
}

func TestIsSoftLink(t *testing.T) {
	tmpDir, filePath, err := createTestFile("test_file.txt")
	if err != nil {
		t.Fatalf("create test file failed: %s", err.Error())
	}
	defer os.RemoveAll(tmpDir)
	linkPath := tmpDir + "/syslink"
	err = os.Symlink(filePath, linkPath)
	if err != nil {
		t.Fatalf("create symlink failed %q: %s", filePath, err)
	}
	convey.Convey("Is softLink test should be softlink", t, func() {
		err = IsSoftLink(linkPath)
		convey.So(err, convey.ShouldResemble, errors.New("can't support symlinks"))
	})

	convey.Convey("Is softLink test file does not exist", t, func() {
		err = IsSoftLink("./xxx/xxxx")
		convey.So(err, convey.ShouldResemble, errors.New("path does not exists"))
	})

	convey.Convey("Is softLink test should success", t, func() {
		err = IsSoftLink(filePath)
		convey.So(err, convey.ShouldBeNil)
	})
}

func createTestFile(fileName string) (string, string, error) {
	const fileMode os.FileMode = 0600
	tmpDir := os.TempDir()
	const permission os.FileMode = 0700
	tmpTestDir := filepath.Join(tmpDir, fmt.Sprintf("__test__%s", strconv.Itoa(tmpFileCount)))
	if os.MkdirAll(tmpTestDir, permission) != nil {
		return "", "", fmt.Errorf("mkdirAll failed %q", tmpTestDir)
	}
	tmpFileCount++
	f, err := os.Create(filepath.Join(tmpTestDir, fileName))
	if err != nil {
		return "", "", fmt.Errorf("create file failed %q: %s", tmpTestDir, err)
	}
	defer f.Close()
	err = f.Chmod(fileMode)
	if err != nil {
		return "", "", fmt.Errorf("change file mode failed %q: %s", tmpTestDir, err)
	}
	return tmpTestDir, filepath.Join(tmpTestDir, fileName), err
}
