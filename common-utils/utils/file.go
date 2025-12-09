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
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"huawei.com/mindx/common/fileutils"
	"huawei.com/mindx/common/rand"
)

const (
	// FileMode file privilege
	FileMode = 0600
	// Size10M  bytes of 10M
	Size10M = 10 * 1024 * 1024
	maxSize = 1024 * 1024 * 1024
)

// ReadLimitBytes read limit length of contents from file path
func ReadLimitBytes(path string, limitLength int) ([]byte, error) {
	key, err := CheckPath(path)
	if err != nil {
		return nil, err
	}
	buf, err := fileutils.ReadLimitBytes(key, limitLength, &fileutils.FileBaseChecker{})
	if err != nil {
		// backward compatibility: do not return error if file is too large
		if !errors.Is(err, fileutils.ErrFileTooLarge) {
			return nil, err
		}
		fmt.Printf("ReadLimitBytes: %d bytes was read from [%s] but the file has more content\n", len(buf), key)
	}
	return buf, nil
}

// LoadFile load file content
func LoadFile(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, nil
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, errors.New("the filePath is invalid")
	}
	if !IsExist(absPath) {
		return nil, nil
	}

	return ReadLimitBytes(absPath, Size10M)
}

func closeFile(file *os.File) {
	if file == nil {
		return
	}
	if err := file.Close(); err != nil {
		return
	}
	return
}

// CopyFile copy file
func CopyFile(src, dst string) error {
	var (
		err     error
		srcFile *os.File
		dstFile *os.File
		srcInfo os.FileInfo
	)

	src, err = CheckPath(src)
	if err != nil {
		return err
	}
	if IsExist(dst) {
		dst, err = CheckPath(dst)
		if err != nil {
			return err
		}
	}
	if srcFile, err = os.Open(src); err != nil {
		return err
	}
	defer closeFile(srcFile)
	if srcInfo, err = os.Stat(src); err != nil {
		return err
	}
	if dstFile, err = os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode()); err != nil {
		return err
	}
	defer closeFile(dstFile)
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return os.Chmod(dst, srcInfo.Mode())
}

// CopyDir recursively copy files
func CopyDir(src string, dst string) error {
	var (
		err     error
		fds     []os.FileInfo
		dstInfo os.FileInfo
	)

	if dstInfo, err = os.Stat(src); err != nil {
		return err
	}
	if err = os.MkdirAll(dst, dstInfo.Mode()); err != nil {
		return err
	}
	if subFolder(src, dst) {
		return errors.New("the destination directory is a subdirectory of the source directory")
	}
	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcFile := filepath.Join(src, fd.Name())
		dstFile := filepath.Join(dst, fd.Name())
		if fd.IsDir() {
			if err = CopyDir(srcFile, dstFile); err != nil {
				return err
			}
		} else {
			if err = CopyFile(srcFile, dstFile); err != nil {
				return err
			}
		}
	}
	return nil
}

func subFolder(src, dst string) bool {
	if src == dst {
		return true
	}
	srcReal, err := filepath.EvalSymlinks(src)
	if err != nil {
		return false
	}
	dstReal, err := filepath.EvalSymlinks(dst)
	if err != nil {
		return false
	}
	srcList := strings.Split(srcReal, string(os.PathSeparator))
	dstList := strings.Split(dstReal, string(os.PathSeparator))
	if len(srcList) > len(dstList) {
		return false
	}
	return reflect.DeepEqual(srcList, dstList[:len(srcList)])
}

// DeleteFile is used to delete one file
func DeleteFile(path string) error {
	if !IsLexist(path) {
		return nil
	}

	dirPath := filepath.Dir(path)
	if _, err := CheckOriginPath(dirPath); err != nil {
		return fmt.Errorf("dir path check failed: %s", err.Error())
	}

	return os.Remove(path)
}

// RenameFile is used to rename (move) old path to new path
func RenameFile(oldPath, newPath string) error {
	if !IsLexist(oldPath) {
		return errors.New("rename file failed: src path does not exist")
	}

	oldPath, err := CheckOriginPath(oldPath)
	if err != nil {
		return fmt.Errorf("check srcPath failed: %s", err.Error())
	}

	if !IsLexist(filepath.Dir(newPath)) {
		return errors.New("rename file failed: dst dir does not exist")
	}

	newPath, err = CheckOriginPath(newPath)
	if err != nil {
		return fmt.Errorf("check dst Path failed: %s", err.Error())
	}

	return os.Rename(oldPath, newPath)
}

// CreateFile is used to create the named file with mode
func CreateFile(filePath string, mode os.FileMode) error {
	if IsLexist(filePath) {
		return nil
	}

	dir := filepath.Dir(filePath)
	if _, err := CheckOriginPath(dir); err != nil {
		return err
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func recursiveConfusionFile(path string, info fs.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	if !isKeyFile(path) {
		return nil
	}

	// if a file larger than maxSize, it should be considered as a malicious file so that we do not confuse it
	if info.Size() > maxSize {
		return nil
	}

	if err = IsSoftLink(path); err != nil {
		return nil
	}

	if err = confusionFile(path, info.Size()); err != nil {
		return err
	}

	return nil
}

func isKeyFile(path string) bool {
	sufList := []string{
		".key",
		".ks",
	}

	for _, suf := range sufList {
		if strings.HasSuffix(path, suf) {
			return true
		}
	}

	return false
}

func confusionFile(path string, size int64) error {
	if size > maxSize {
		size = maxSize
	}
	// Override with zero
	overrideByte := make([]byte, size, size)
	if err := WriteData(path, overrideByte); err != nil {
		return fmt.Errorf("confusion file with 0 failed: %s", err.Error())
	}

	for i := range overrideByte {
		overrideByte[i] = 0xff
	}
	if err := WriteData(path, overrideByte); err != nil {
		return fmt.Errorf("confusion file with 1 failed: %s", err.Error())
	}

	if _, err := rand.Read(overrideByte); err != nil {
		return errors.New("get random words failed")
	}
	if err := WriteData(path, overrideByte); err != nil {
		return fmt.Errorf("confusion file with random num failed: %s", err.Error())
	}

	return nil
}

// WriteData is used to write data with path check
func WriteData(filePath string, fileData []byte) error {
	filePath, err := CheckOriginPath(filePath)
	if err != nil {
		return err
	}

	err = MakeSureDir(filePath)
	if err != nil {
		return err
	}

	writer, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FileMode)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write(fileData)
	return err
}

// DeleteAllFileWithConfusion is used to delete all files with confusion
func DeleteAllFileWithConfusion(filePath string) error {
	if !IsLexist(filePath) {
		return nil
	}

	realPath, err := CheckOriginPath(filePath)
	if err != nil {
		return fmt.Errorf("confusion path %s failed: %s", filePath, err.Error())
	}

	if err = filepath.Walk(realPath, recursiveConfusionFile); err != nil {
		return fmt.Errorf("confusion path %s failed: %s", filePath, err.Error())
	}

	return os.RemoveAll(filePath)
}

// ReadFile is used to read file content
func ReadFile(filePath string) ([]byte, error) {
	data, err := LoadFile(filePath)
	if err == nil && data == nil {
		return nil, errors.New("the file does not exist")
	}
	return data, err
}

// ReadDir is used to read the file list in a dir
func ReadDir(path string) ([]os.DirEntry, error) {
	if _, err := CheckOriginPath(path); err != nil {
		return nil, err
	}

	if !IsDir(path) {
		return nil, fmt.Errorf("path %s is not dir", path)
	}

	return os.ReadDir(path)
}

// GetFileSha256 is used to get the sha256sum value of a file
func GetFileSha256(path string) (string, error) {
	path, err := RealFileChecker(path, false, false, maxAllowFileSize)
	if err != nil {
		return "", err
	}

	file, err := LoadFile(path)
	if err != nil {
		return "", fmt.Errorf("open file failed: %s", err.Error())
	}

	hash := sha256.New()
	if _, err := hash.Write(file); err != nil {
		return "", fmt.Errorf("get file sha256sum failed: %s", err.Error())
	}

	// The returned sha256 value should be a hexadecimal number.
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
