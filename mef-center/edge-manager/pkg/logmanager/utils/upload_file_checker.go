// Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
// MindEdge is licensed under Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//          http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
// EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
// MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
// See the Mulan PSL v2 for more details.

// Package utils
package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"huawei.com/mindxedge/base/common"
)

const (
	maxFileCount = 200
	maxFileSize  = 50 * common.MB
)

// UploadFileChecker checks log package
type UploadFileChecker struct {
	Sha256Checksum string
	File           *os.File
}

// Check checks log package
func (c UploadFileChecker) Check() error {
	checkFns := []func(*UploadFileChecker) error{
		checkTarGz,
		checkSha256sum,
	}
	for _, fn := range checkFns {
		if _, err := c.File.Seek(0, io.SeekStart); err != nil {
			return err
		}
		if err := fn(&c); err != nil {
			return err
		}
	}
	return nil
}

func checkTarGz(checker *UploadFileChecker) error {
	gzipReader, err := gzip.NewReader(checker.File)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(gzipReader)

	var fileCount int
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		fileCount++
		if fileCount > maxFileCount {
			return errors.New("max file count exceeded")
		}

		if err := checkTarEntry(header, tarReader, maxFileSize); err != nil {
			return err
		}
	}
	return nil
}

func checkSha256sum(checker *UploadFileChecker) error {
	hash := sha256.New()
	if _, err := io.Copy(hash, checker.File); err != nil {
		return errors.New("calculate sha256 checksum error")
	}
	if fmt.Sprintf("%x", hash.Sum(nil)) != checker.Sha256Checksum {
		return errors.New("sha256 checksum error")
	}
	return nil
}

func checkTarEntry(header *tar.Header, content io.Reader, maxFileSize int64) error {
	if header.Size > maxFileSize || header.Size < 0 {
		return errors.New("tar entry size exceeded")
	}
	const umask337 = 0337
	if header.Mode&umask337 != 0 {
		return errors.New("bad entry mode")
	}
	if header.Typeflag != tar.TypeReg {
		return errors.New("bad entry type")
	}
	if filepath.IsAbs(header.Name) || strings.Contains(header.Name, "..") {
		return errors.New("invalid entry name")
	}

	switch filepath.Ext(header.Name) {
	case ".gz":
		return checkGz(content, maxFileSize)
	case ".log":
		return nil
	default:
		return errors.New("bad tar entry extension")
	}
}

func checkGz(content io.Reader, maxFileSize int64) error {
	gzReader, err := gzip.NewReader(content)
	if err != nil {
		return err
	}

	nRead, err := io.CopyN(io.Discard, gzReader, maxFileSize+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if nRead > maxFileSize {
		return errors.New("gzip is too large")
	}
	return nil
}
