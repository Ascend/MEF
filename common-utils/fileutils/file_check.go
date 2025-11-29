//  Copyright(c) 2023. Huawei Technologies Co.,Ltd.  All rights reserved.

// Package fileutils provides the util func to deal with file
package fileutils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// CheckOwnerAndPermission check path  owner and permission
func CheckOwnerAndPermission(verifyPath string, modeUmask os.FileMode, userId uint32) (string, error) {
	ownerChecker := NewFileOwnerChecker(false, false, userId, userId)
	modeChecker := NewFileModeChecker(false, modeUmask, false, false)
	ownerChecker.SetNext(modeChecker)

	file, err := os.OpenFile(verifyPath, os.O_RDONLY, Mode400)
	if err != nil {
		return "", fmt.Errorf("open file %s failed: %s", verifyPath, err.Error())
	}
	defer CloseFile(file)

	if err = ownerChecker.Check(file, verifyPath); err != nil {
		return "", fmt.Errorf("check file %s failed: %s", verifyPath, err.Error())
	}

	realPath, err := EvalSymlinks(verifyPath)
	if err != nil {
		return "", err
	}

	return realPath, nil
}

// CheckOriginPath valid the path and return the real path,
// can not support the relative path, for example:  ../ in path will not support
func CheckOriginPath(filePath string) (string, error) {
	checker := NewFileLinkChecker(false)
	existPath, err := getExistsDir(filePath)
	if err != nil {
		return "", fmt.Errorf("get exist dir failed: %s", err.Error())
	}
	file, err := os.OpenFile(existPath, os.O_RDONLY, Mode400)
	if err != nil {
		return "", fmt.Errorf("open file %s failed: %s", filePath, err.Error())
	}
	defer CloseFile(file)

	if err = checker.Check(file, existPath); err != nil {
		return "", fmt.Errorf("check file %s failed: %s", filePath, err.Error())
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("get file abs path failed:%s", err.Error())
	}

	return filepath.Clean(absPath), nil
}

// IsExist check whether the path exists, If the file is a symbolic link, the returned the final FileInfo
func IsExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil || errors.Is(err, fs.ErrExist) {
		return true
	}
	return false
}

// IsLexist check whether the path exists, If the file is a symbolic link, the returned FileInfo
// describes the symbolic link
func IsLexist(filePath string) bool {
	_, err := os.Lstat(filePath)
	if err == nil || errors.Is(err, fs.ErrExist) {
		return true
	}
	return false
}

// IsSoftLink is the func to check if a path is a soft link
func IsSoftLink(path string) error {
	if !IsLexist(path) {
		return errors.New("path does not exists")
	}

	file, err := os.OpenFile(path, os.O_RDONLY, Mode400)
	if err != nil {
		return fmt.Errorf("open file %s failed: %s", path, err.Error())
	}
	defer CloseFile(file)
	checker := NewFileLinkChecker(true)
	if err = checker.Check(file, path); err != nil {
		return err
	}

	return nil
}

// CheckMode check input file mode whether includes invalid mode.
// For example, if read operation of group and other is forbidden, then call CheckMode(inputFileMode, 0044).
// All operations are forbidden for group and other, then call CheckMode(inputFileMode, 0077).
// Write operation is forbidden for group and other by default, with calling CheckMode(inputFileMode)
func CheckMode(mode os.FileMode, optional ...os.FileMode) bool {
	var targetMode os.FileMode
	if len(optional) > 0 {
		targetMode = optional[0]
	} else {
		targetMode = DefaultWriteFileMode
	}
	checkMode := uint32(mode) & uint32(targetMode)
	return checkMode == 0
}

// RealFileCheck Check whether a file is valid
func RealFileCheck(path string, checkParent, allowLink bool, maxSizeInMb int64) (string, error) {
	const oneMegabytes = 1024 * 1024
	pathChecker := NewFilePathChecker()
	ownerChecker := NewFileOwnerChecker(checkParent, true, RootUid, RootGid)
	modeChecker := NewFileModeChecker(checkParent, DefaultWriteFileMode, true, true)
	sizeChecker := NewFileSizeChecker(maxSizeInMb * oneMegabytes)
	dirChecker := NewIsDirChecker(false)

	pathChecker.SetNext(ownerChecker)
	pathChecker.SetNext(modeChecker)
	pathChecker.SetNext(sizeChecker)
	pathChecker.SetNext(dirChecker)

	if !allowLink {
		linkChecker := NewFileLinkChecker(true)
		pathChecker.SetNext(linkChecker)
	}

	file, err := check(path, os.O_RDONLY, Mode400, pathChecker)
	if err != nil {
		return "", fmt.Errorf("check file %s failed: %s", path, err.Error())
	}
	defer CloseFile(file)

	realPath, err := GetRealPath(file, path)
	if err != nil {
		return "", fmt.Errorf("get file real path failed: %s", err.Error())
	}

	return realPath, nil
}

// RealDirCheck Check whether the directory is valid
func RealDirCheck(path string, checkParent, allowLink bool) (string, error) {
	pathChecker := NewFilePathChecker()
	ownerChecker := NewFileOwnerChecker(checkParent, true, RootUid, RootGid)
	modeChecker := NewFileModeChecker(checkParent, DefaultWriteFileMode, true, true)
	dirChecker := NewIsDirChecker(true)

	pathChecker.SetNext(ownerChecker)
	pathChecker.SetNext(modeChecker)
	pathChecker.SetNext(dirChecker)

	if !allowLink {
		linkChecker := NewFileLinkChecker(true)
		pathChecker.SetNext(linkChecker)
	}

	file, err := check(path, os.O_RDONLY, Mode400, pathChecker)
	if err != nil {
		return "", fmt.Errorf("check file %s failed: %s", path, err.Error())
	}
	defer CloseFile(file)

	realPath, err := GetRealPath(file, path)
	if err != nil {
		return "", fmt.Errorf("get file real path failed: %s", err.Error())
	}

	return realPath, nil

}
