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
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxExtractFileCount = 1000
	maxTotalExtractSize = 1000 * 1024 * 1024
	maxPkgSizeInMb      = 500
	fileNameRegex       = "^[a-zA-Z0-9_/.-]{1,256}$"
)

var symlinkPaths []string

// ExtraTarGzFile extract tar.gz file
func ExtraTarGzFile(tarGzFile, extractPath string, allowLink bool) error {
	realPkgPath, cleanExtractPath, err := checkAndPrepare(tarGzFile, extractPath)
	if err != nil {
		return fmt.Errorf("check and prepare failed, error: %v", err)
	}

	srcFile, err := os.Open(realPkgPath)
	if err != nil {
		return errors.New("open tar.gz file failed")
	}
	defer CloseFile(srcFile)

	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return errors.New("create gzip reader failed")
	}
	defer func() {
		if err = gzReader.Close(); err != nil {
			return
		}
	}()

	tarReader := tar.NewReader(gzReader)
	if err = extraTarGzFile(tarReader, cleanExtractPath, allowLink); err != nil {
		return fmt.Errorf("extract tar.gz file failed, error: %v", err)
	}
	return nil
}

func extraTarGzFile(tarReader *tar.Reader, extractPath string, allowLink bool) error {
	var (
		fileCount int
		totalSize int64
	)

	symlinkPaths = []string{}
	defer func() {
		symlinkPaths = []string{}
	}()
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("get next entry in tar file failed: %v", err)
		}

		fileCount += 1
		if fileCount > maxExtractFileCount {
			return errors.New("too many file will be uncompressed")
		}
		totalSize += header.Size
		// compare header.Size with maxTotalExtractSize to detect integer overflow
		if totalSize > maxTotalExtractSize || header.Size > maxTotalExtractSize || header.Size < 0 {
			return errors.New("too big file will be uncompressed")
		}
		if err = fileNameCheck(header.Name); err != nil {
			return fmt.Errorf("check file name [%s] failed: %v", header.Name, err)
		}
		if err = copyTarFile(extractPath, header, tarReader, allowLink); err != nil {
			return fmt.Errorf("extract [%s] failed: %v", header.Name, err)
		}
	}
	if err := symlinkCheck(extractPath); err != nil {
		return fmt.Errorf("check symlink failed: %v", err)
	}
	return nil
}

func fileNameCheck(fileName string) error {
	compile := regexp.MustCompile(fileNameRegex)
	if !compile.MatchString(fileName) {
		return errors.New("file name not match regex requirement")
	}

	excludeWord := ".."
	if strings.Contains(fileName, excludeWord) {
		return fmt.Errorf("file name contains exclude words [%s]", excludeWord)
	}
	return nil
}

func checkAndPrepare(tarGzFile, extractPath string) (string, string, error) {
	realPkgPath, err := RealFileCheck(tarGzFile, true, false, maxPkgSizeInMb)
	if err != nil {
		return "", "", fmt.Errorf("check compressed file failed, error: %v", err)
	}

	extractPath = filepath.Clean(extractPath)
	if err = CreateDir(extractPath, Mode700); err != nil {
		return "", "", fmt.Errorf("prepare extract path failed, error: %v", err)
	}
	return realPkgPath, extractPath, nil
}

func symlinkCheck(extractPath string) error {
	for _, symlinkPath := range symlinkPaths {
		realPath, err := EvalSymlinks(symlinkPath)
		if err != nil {
			return fmt.Errorf("eval symlinks of path [%s] failed: %v", symlinkPath, err)
		}

		if !strings.HasPrefix(realPath, extractPath) {
			return fmt.Errorf("symlink [%s] cannot point to [%s], it is outside the extraction path [%s]",
				symlinkPath, realPath, extractPath)
		}
	}
	return nil
}

func copyTarFile(extractPath string, header *tar.Header, tarReader *tar.Reader, allowLink bool) error {
	extraFilePath := filepath.Join(extractPath, header.Name)
	switch header.Typeflag {
	case tar.TypeDir:
		if err := CreateDir(extraFilePath, header.FileInfo().Mode()&Mode755, NewFileLinkChecker(false)); err != nil {
			return fmt.Errorf("create path [%s] failed, error: %v", extraFilePath, err)
		}
	case tar.TypeReg:
		targetFile, err := os.OpenFile(extraFilePath, os.O_WRONLY|os.O_CREATE, os.FileMode(header.Mode))
		if err != nil {
			return fmt.Errorf("open dst file [%s] failed, error: %v", header.Name, err)
		}

		defer CloseFile(targetFile)
		modeChecker := NewFileModeChecker(false, DefaultWriteFileMode, true, false)
		modeChecker.SetNext(NewFileLinkChecker(false))
		if err := modeChecker.Check(targetFile, extraFilePath); err != nil {
			return fmt.Errorf("copy src file [%s] failed, error: %v", header.Name, err)
		}
		if err := targetFile.Truncate(0); err != nil {
			return fmt.Errorf("truncate file failed, %v", err)
		}
		if _, err = io.CopyN(targetFile, tarReader, header.Size); err != nil {
			return fmt.Errorf("copy src file [%s] failed, error: %v", header.Name, err)
		}
	case tar.TypeSymlink:
		if !allowLink {
			return fmt.Errorf("do not support symlink[%s]", header.Name)
		}
		if filepath.IsAbs(header.Linkname) || fileNameCheck(header.Linkname) != nil {
			return fmt.Errorf("invalid link name[%s]", header.Linkname)
		}
		if err := os.Symlink(header.Linkname, extraFilePath); err != nil {
			return fmt.Errorf("create symlink[%s] failed,error:%v", header.Name, err)
		}
		symlinkPaths = append(symlinkPaths, extraFilePath)
	default:
		return fmt.Errorf("do not support the type of [%c]", header.Typeflag)
	}
	return nil
}
