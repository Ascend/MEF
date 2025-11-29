// Package mindxcheckutils is a check utils package
/**
 * Copyright (c) Huawei Technologies Co., Ltd. 2020-2022. All rights reserved.
 */
package utils

import (
	"os"
	"strings"
	"testing"
)

func TestNormalFileCheckRegularFile(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	defer removeTmpDir(t, tmpDir)
	err = os.Symlink(filePath, tmpDir+"/syslink")
	if err != nil {
		t.Fatalf("create symlink failed %q: %s", filePath, err)
	}

	if _, err = normalFileCheck(tmpDir, true, false); err != nil {
		t.Fatalf("check allow dir failed %q: %s", tmpDir+"/__test__", err)
	}

	if _, err = normalFileCheck(tmpDir, false, false); !strings.Contains(err.Error(), "not regular file") {
		t.Fatalf("check not allow dir failed %q: %s", tmpDir+"/__test__", err)
	}

	if _, err = normalFileCheck("/dev/zero", true, false); !strings.Contains(err.Error(), "not regular file/dir") {
		t.Fatalf("check /dev/zero failed %q: %s", tmpDir+"/__test__", err)
	}

	if _, err = normalFileCheck(tmpDir+"/syslink", false, false); !strings.Contains(err.Error(), "symlinks") {
		t.Fatalf("check symlinks failed %q: %s", tmpDir+"/syslink", err)
	}

	if _, err = normalFileCheck(filePath, false, false); err != nil {
		t.Fatalf("check failed %q: %s", filePath, err)
	}

	if _, err = normalFileCheck(tmpDir+"/notexisted", false, false); !strings.Contains(err.Error(), "not existed") {
		t.Fatalf("check symlinks failed %q: %s", tmpDir+"/syslink", err)
	}
}

func TestRealFileChecker(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	defer removeTmpDir(t, tmpDir)
	const permission os.FileMode = 0700
	err = os.WriteFile(filePath, []byte("hello\n"), permission)
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	if _, err = RealFileChecker(filePath, false, true, 0); err == nil {
		t.Fatalf("size check wrong 0 %q: %s", filePath, err)
	}
	if _, err = RealFileChecker(filePath, false, true, 1); err != nil {
		t.Fatalf("size check wrong 1 %q: %s", filePath, err)
	}
}

func TestRealFileCheckerInside(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	defer removeTmpDir(t, tmpDir)
	const permission os.FileMode = 0700
	const deep int = 100
	err = os.WriteFile(filePath, []byte("hello\n"), permission)
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	if err = fileChecker(filePath, false, false, false, deep); err == nil {
		t.Fatalf("size check wrong 0 %q: %s", filePath, err)
	}
}

func TestRealDirChecker(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	defer removeTmpDir(t, tmpDir)
	if _, err = RealDirChecker(filePath, false, true); err == nil {
		t.Fatalf("should be dir 0 %q: %s", filePath, err)
	}
	if _, err = RealDirChecker(tmpDir, false, true); err != nil {
		t.Fatalf("should be dir 1 %q: %s", filePath, err)
	}
}

func TestVerifyFile(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	defer removeTmpDir(t, tmpDir)
	err = os.Symlink(filePath, tmpDir+"/syslink")
	if err != nil {
		t.Fatalf("create symlink failed %q: %s", filePath, err)
	}
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("open file failed")
	}
	defer file.Close()
	linkFile, err := os.Open(tmpDir + "/syslink")
	if err != nil {
		t.Fatalf("open file failed")
	}
	defer linkFile.Close()
	const permission os.FileMode = 0700
	err = os.WriteFile(filePath, []byte("hello\n"), permission)
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	if err = VerifyFile(file, 0); err == nil {
		t.Fatalf("size check wrong 0 %q: %s", filePath, err)
	}
	if err = VerifyFile(file, 1); err != nil {
		t.Fatalf("size check wrong 1 %q: %s", filePath, err)
	}
	if err = VerifyFile(linkFile, 1); err != nil && !strings.Contains(err.Error(), "symlinks") {
		t.Fatalf("check symlinks failed %q: %s", tmpDir+"/syslink", err)
	}
}

func TestStringChecker(t *testing.T) {
	if ok := stringChecker("0123456789abcABC", 0, DefaultStringLength); !ok {
		t.Fatalf("failed on regular letters")
	}
	const testSize = 3
	if ok := stringChecker("123", 0, testSize); ok {
		t.Fatalf("failed on max length")
	}
	if ok := stringChecker("1234", 0, testSize); ok {
		t.Fatalf("failed on max length")
	}
	if ok := stringChecker("12", 0, testSize); !ok {
		t.Fatalf("failed on max length")
	}
	if ok := stringChecker("", 0, testSize); ok {
		t.Fatalf("failed on min length")
	}
	if ok := stringChecker("123", testSize, DefaultStringLength); ok {
		t.Fatalf("failed on min length")
	}
	if ok := stringChecker("123%", 0, DefaultStringLength); ok {
		t.Fatalf("failed on strange words")
	}
	if ok := stringChecker("123.-/~", 0, DefaultStringLength); !ok {
		t.Fatalf("failed on strange words")
	}
}

func TestIsSoftLink(t *testing.T) {
	tmpDir, filePath, err := createTestFile(t, "test_file.txt")
	if err != nil {
		t.Fatalf("create file failed %q: %s", filePath, err)
	}
	defer removeTmpDir(t, tmpDir)
	linkPath := tmpDir + "/syslink"
	err = os.Symlink(filePath, linkPath)
	if err != nil {
		t.Fatalf("create symlink failed %q: %s", filePath, err)
	}
	if err = IsSoftLink(linkPath); err == nil {
		t.Fatalf("failed on soft link check, should be softlink")
	}
	if err = IsSoftLink("./xxx/xxxx"); err == nil {
		t.Fatalf("failed on soft link check, should return err")
	}
	if err = IsSoftLink(filePath); err != nil {
		t.Fatalf("failed on soft link check, should not be softlink")
	}
}

func createTestFile(t *testing.T, fileName string) (string, string, error) {
	const fileMode os.FileMode = 0600
	tmpDir := os.TempDir()
	const permission os.FileMode = 0700
	if os.MkdirAll(tmpDir+"/__test__", permission) != nil {
		t.Fatalf("MkdirAll failed %q", tmpDir+"/__test__")
	}
	f, err := os.Create(tmpDir + "/__test__" + fileName)
	if err != nil {
		t.Fatalf("create file failed %q: %s", tmpDir+"/__test__", err)
	}
	defer f.Close()
	err = f.Chmod(fileMode)
	if err != nil {
		t.Fatalf("change file mode failed %q: %s", tmpDir+"/__test__", err)
	}
	return tmpDir + "/__test__", tmpDir + "/__test__" + fileName, err
}

func removeTmpDir(t *testing.T, tmpDir string) {
	if os.RemoveAll(tmpDir) != nil {
		t.Logf("removeall %v", tmpDir)
	}
}
